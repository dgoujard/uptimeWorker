package main

import (
	"github.com/BurntSushi/toml"
	"github.com/dgoujard/uptimeWorker/config"
	"github.com/dgoujard/uptimeWorker/services"
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
	uptimeService := services.CreateUptimeService()
	//queueWorker := services.CreateQueueWorker(&configFile.Amq,queueService,uptimeService)

	if len(os.Args) > 1 && os.Args[1] == "withCron" {
		cronService := services.CreateCronService(databaseService,queueService)
		cronService.StartCronProcess()
	}
	test := services.SiteBdd{Url:"http://www.actigraph.com"}
	//test := services.SiteBdd{Url:"http://ccinormandiedev.cartographie.pro"}
	uptimeService.CheckSite(&test)

	//queueWorker.StartAmqWatching()
}
