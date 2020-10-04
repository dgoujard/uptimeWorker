package main

import (
	"github.com/dgoujard/uptimeWorker/app/services"
	"github.com/dgoujard/uptimeWorker/config"
	"log"
	"os"
)

func main() {
	var appConfig = config.AppConfig()

	databaseService := services.CreateDatabaseConnection(&appConfig.Database)
	queueService := services.CreateQueueService(&appConfig.Amq)
	awsService := services.CreateAwsService(&appConfig.Aws)
	uptimeCheckerService := services.CreateUptimeCheckerService()
	uptimeService := services.CreateUptimeService(&appConfig.Worker,uptimeCheckerService,awsService,queueService,databaseService)
	realtimeService := services.CreateRealtimeClient(&appConfig.Realtime)
	//	test := services.SiteBdd{Url:"http://www.actigraphSS.com"}
	//fmt.Println(awsService.CallUptimeCheckLambdaFunc("arn:aws:lambda:eu-west-1:312046026144:function:uptimeCheck",test))
	/*test := services.SiteBdd{Url:"http://www.tgl-longwy.fr"}
	fmt.Println(uptimeCheckerService.CheckSite(&test))
	os.Exit(0)*/
	/*
		for i := 0; i < 10; i++ {
			go func() {
				uptimeService.CheckSite(&test)
			}()
		}
		time.Sleep(2*time.Second)*/

	cliOptions := getCliParams()
	if _, ok := cliOptions["withCron"]; ok {
		cronService := services.CreateCronService(databaseService,queueService)
		cronService.StartCronProcess()
		log.Println("Cron enabled")
	}

	var alerteService *services.AlerteService
	if _, ok := cliOptions["withAlerte"]; ok {
		alerteService = services.CreateAlerteService(&appConfig.Alert,awsService,databaseService,realtimeService)
		log.Println("Alerte enabled")
	}
	queueWorker := services.CreateQueueWorker(&appConfig.Amq,queueService,uptimeService,alerteService)
	queueWorker.StartAmqWatching()
}


func getCliParams() (result map[string]bool) {
	result = make(map[string]bool)
	if len(os.Args) > 1 {
		for _, item := range os.Args {
			result[item] = true
		}
	}
	return result
}