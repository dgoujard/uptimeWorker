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

	if config.EnableInfluxDb2Reporting {
		influx, err = influxdb.New(nil,
			influxdb.WithAddress(config.InfluxDb2Url),
			influxdb.WithToken(config.InfluxDb2Token),
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
	if result.Err != "" || result.HttpCode != 200 {
		if u.config.EnableLambdaRemoteCheck && site.Status != 9 { //Si c'est pas déjà en panne
			err, resultLambda := u.awsService.CallUptimeCheckLambdaFunc(u.config.LambdaArn,site)
			if err != nil {
				log.Println("Requéte via Lambda en erreur. Probable probléme connexion woker")
			}else {
				if resultLambda.Err == "" && resultLambda.HttpCode == 200 {
					log.Println(site.Url, " LAMBDA up (", resultLambda.Duration, ")")
					u.logResponseTime(site,*resultLambda)
					return
				} else {
					log.Println(site.Url, " LAMBDA DOWN (", resultLambda.Duration, ")")
				}
			}
		}

		if site.Status == 2  {
			log.Println(site.Url," DOWN ",result.HttpCode," ",result.Err ,"(",result.Duration,")")
			u.logResponse(site,result)
		}
		return
	} else if result.Err  == "" && result.HttpCode == 200 {
		u.logResponseTime(site,result)
		if site.Status == 9  {
			log.Println(site.Url," up (",result.Duration,")")
			u.logResponse(site,result)
		}

		return
	}
}
func (u *UptimeService)logResponse(site *SiteBdd, result CheckSiteResponse)  {
	var isDown bool

	//Mise à jour du site
	if result.Err != "" || result.HttpCode != 200 {
		u.databaseService.UpdateSiteStatus(site,9)
		isDown=true
	}else{
		//TODO calcul de la duration car sera utile pour le mail de notification (possible d'afficher la durée)
		u.databaseService.UpdateSiteStatus(site,2)
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
	}
	err := u.databaseService.AddLogForSite(site,&logSite)
	if err != nil {
		log.Println(err)
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
		if err := u.influx.Write(context.Background(), u.influxBucket, u.influxOrg, myMetrics...); err != nil {
			log.Fatal(err) // as above use your own error handling here.
		}
	}
}