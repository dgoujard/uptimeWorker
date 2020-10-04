package main

import (
	"github.com/BurntSushi/toml"
	"github.com/dgoujard/uptimeWorker/app/services"
	"github.com/dgoujard/uptimeWorker/config"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	var baseCacheDir = dir
	if strings.HasPrefix(dir, "/private/") || strings.HasPrefix(dir, "/var/folders/") {
		baseCacheDir = "/Users/damien/uptimeWorker"
	}

	l := &lumberjack.Logger{
		Filename:   baseCacheDir + "/app.log",
		MaxSize:    1, // megabytes
		MaxAge:     7, //days
		MaxBackups: 3,
		LocalTime:  true,
		Compress:   true, // disabled by default
	}
	//log.SetOutput(l)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for {
			<-c
			l.Rotate()
		}
	}()

	log.Println("Reading configuration")
	var configFile config.TomlConfig
	if _, err := toml.DecodeFile("config.toml", &configFile); err != nil {
		log.Println(err)
		return
	}

	databaseService := services.CreateDatabaseConnection(&configFile.Database)
	queueService := services.CreateQueueService(&configFile.Amq)
	awsService := services.CreateAwsService(&configFile.Aws)
	uptimeCheckerService := services.CreateUptimeCheckerService()
	uptimeService := services.CreateUptimeService(&configFile.Worker,uptimeCheckerService,awsService,queueService,databaseService)
	realtimeService := services.CreateRealtimeClient(&configFile.Realtime)
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
		alerteService = services.CreateAlerteService(&configFile.Alert,awsService,databaseService,realtimeService)
		log.Println("Alerte enabled")
	}
	queueWorker := services.CreateQueueWorker(&configFile.Amq,queueService,uptimeService,alerteService)
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