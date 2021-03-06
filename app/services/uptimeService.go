package services

import (
	"context"
	"github.com/dgoujard/uptimeWorker/config"
	"github.com/influxdata/influxdb-client-go"
	"log"
	"net/http"
	"time"
)


type UptimeService struct {
	checker *uptimeCheckerService
	awsService *AwsService
	queueService *QueueService
	databaseService *DatabaseService
	config *config.WorkerConfig
	influx *influxdb.Client
	influxBucket string
	influxOrg string
}
func CreateUptimeService(config *config.WorkerConfig, uptimeChecker *uptimeCheckerService, awsService *AwsService, queueService *QueueService, databaseservice *DatabaseService) *UptimeService {
	var influx *influxdb.Client
	var err error

	if config.EnableInfluxDb2Reporting && len(config.InfluxDb2Url) > 0 {
		influx, err = influxdb.New(
			config.InfluxDb2Url,
			config.InfluxDb2Token,
		)
		if err != nil {
			panic(err) // error handling here; normally we wouldn't use fmt but it works for the example
		}
	}

	return &UptimeService{
		checker:uptimeChecker,
		awsService: awsService,
		queueService: queueService,
		databaseService: databaseservice,
		config:config,
		influx:influx,
		influxBucket:config.InfluxDb2Bucket,
		influxOrg:config.InfluxDb2Org,
	}
}

func (u *UptimeService) CheckSite(site *SiteBdd){
	result := u.checker.CheckSite(site)
	if result.Err != "" || result.HttpCode != 200 { //TODO mutialiser le test up/down
		if u.config.EnableLambdaRemoteCheck && site.Status != SiteStatusDown { //Si c'est pas déjà en panne
			err, resultLambda := u.awsService.CallUptimeCheckLambdaFunc(u.config.LambdaArn,site)
			if err != nil {
				log.Println("Requéte via Lambda en erreur. Probable probléme connexion woker")
			}else {
				if resultLambda.Err == "" && resultLambda.HttpCode == 200 { //TODO mutialiser le test up/down
					log.Println(site.Url, " LAMBDA up (", resultLambda.Duration, ")")
					u.logResponseTime(site,*resultLambda)
					return
				} else {
					log.Println(site.Url, " LAMBDA DOWN (", resultLambda.Duration, ")")
				}
			}
		}

		if site.Status != SiteStatusDown  { //Si pas déjà en état down je log
			log.Println(site.Url," DOWN ",result.HttpCode," ",result.Err ,"(",result.Duration,")")
			u.logResponse(site,result)
		}
		return
	} else if result.Err  == "" && result.HttpCode == 200 { //TODO mutialiser le test up/down
		u.logResponseTime(site,result)
		if site.Status != SiteStatusUp  { //Si le site n'étair pas marqué up alors je le marque up
			log.Println(site.Url," up (",result.Duration,")")
			u.logResponse(site,result)
		}

		return
	}
}
func (u *UptimeService)logResponse(site *SiteBdd, result CheckSiteResponse)  {
	var isDown bool

	//Mise à jour du site
	if result.Err != "" || result.HttpCode != 200 { //TODO mutialiser le test up/down
		isDown=true
	}else{
		isDown = false
	}

	//Creation du log
	logCode := result.HttpCode
	if logCode == 0{
		logCode = 333333 //Si c'est 0 alors j'ai annulé la requéte donc ont met 333333
	}
	logDetail := result.Err
	if logDetail == ""{
		logDetail = http.StatusText(result.HttpCode)
	}
	loc, _ := time.LoadLocation("Europe/Paris")
	logSite := LogBdd{
		Datetime: time.Now().In(loc).Unix(),
		Code:     logCode,
		Detail:   logDetail,
		TakeIntoAccount: true,
	}
	err := u.databaseService.AddLogForSite(site,&logSite,isDown)
	if err != nil {
		log.Println(err)
	}

	//Mise à jour site
	if isDown {
		u.databaseService.UpdateSiteStatus(site,SiteStatusDown,logSite.Datetime)
	}else{
		u.databaseService.UpdateSiteStatus(site,SiteStatusUp,logSite.Datetime)
	}
	//Creation de l'alerte
	u.queueService.AddAlertToAmqQueue(&Alerte{
		Site:site,
		Type: "uptime",
	},&AlerteParamUptime{
		IsCurrentlyDown: isDown,
		LogSite: &logSite,
	})
}
func (u *UptimeService)logResponseTime(site *SiteBdd, result CheckSiteResponse)  {
	if u.influx != nil {
		myMetrics := []influxdb.Metric{
			influxdb.NewRowMetric(
				map[string]interface{}{"total": result.Duration.Nanoseconds()/ 1e6},
				"reponse_time",
				map[string]string{"name": site.Name,"url":site.Url,"id":site.Id.Hex()},
				time.Now()),
		}
		if _, err := u.influx.Write(context.Background(), u.influxBucket, u.influxOrg, myMetrics...); err != nil {
			log.Fatal(err) // as above use your own error handling here.
		}
	}
}