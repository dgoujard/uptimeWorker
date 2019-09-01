package services

import (
	"context"
	"github.com/dgoujard/uptimeWorker/config"
	"github.com/influxdata/influxdb-client-go"
	"log"
	"time"
)

type AlerteParamUptime struct {
	ResultUptime *CheckSiteResponse
	IsCurrentlyDown bool
}

type Alerte struct {
	Site *SiteBdd
	Type string
	Param interface{}
}

type UptimeService struct {
	checker *uptimeCheckerService
	awsService *AwsService
	queueService *QueueService
	config *config.WorkerConfig
	influx *influxdb.Client
	influxBucket string
	influxOrg string
}
func CreateUptimeService(config *config.WorkerConfig, uptimeChecker *uptimeCheckerService, awsService *AwsService, queueService *QueueService) *UptimeService {

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
		config:config,
		influx:influx,
		influxBucket:config.InfluxDb2Bucket,
		influxOrg:config.InfluxDb2Org,
	}
}

func (u *UptimeService) CheckSite(site *SiteBdd){
	result := u.checker.CheckSite(site)
	if result.Err != "" || result.HttpCode != 200 {
		if u.config.EnableLambdaRemoteCheck {
			err, resultLambda := u.awsService.CallUptimeCheckLambdaFunc(u.config.LambdaArn,site)
			if err != nil {
				log.Println("Requéte via Lambda en erreur. Probable probléme connexion woker")
			}else {
				if resultLambda.Err == "" && resultLambda.HttpCode == 200 {
					log.Println(site.Url, " LAMBDA up (", resultLambda.Duration, ")")
					return
				} else {
					log.Println(site.Url, " LAMBDA DOWN (", resultLambda.Duration, ")")
				}
			}
		}

		//TODO si site.Status n'est pas en erreur alors enregistrement dans mongo
		log.Println(site.Url," DOWN ",result.HttpCode," ",result.Err ,"(",result.Duration,")")
		u.queueService.AddAlertToAmqQueue(&Alerte{
			Site:site,
			Type: "uptime",
			Param: &AlerteParamUptime{
				ResultUptime:    &result,
				IsCurrentlyDown: true,
			},
		})
		return
	} else if result.Err  == "" && result.HttpCode == 200 {
		//TODO si site.Status indique erreur alors que je n'ai plus d'erreur alors enregistrement dans mongo
		log.Println(site.Url," up (",result.Duration,")")
		u.queueService.AddAlertToAmqQueue(&Alerte{
			Site:site,
			Type: "uptime",
			Param: &AlerteParamUptime{
				ResultUptime:    &result,
				IsCurrentlyDown: false,
			},
		})
		u.logResponseType(site,result)
		return
	}
}

func (u *UptimeService)logResponseType(site *SiteBdd, result CheckSiteResponse)  {
	if u.influx != nil {
		localFrance, _ := time.LoadLocation("Europe/Paris")

		myMetrics := []influxdb.Metric{
			influxdb.NewRowMetric(
				map[string]interface{}{"total": result.Duration.Nanoseconds()/ 1e6},
				"reponse_time",
				map[string]string{"name": site.Name,"url":site.Url,"id":site.Id.Hex()},
				time.Now().In(localFrance)),
		}
		if err := u.influx.Write(context.Background(), u.influxBucket, u.influxOrg, myMetrics...); err != nil {
			log.Fatal(err) // as above use your own error handling here.
		}
	}
}