package main

import (
	"context"
	"fmt"
	"github.com/dgoujard/uptimeWorker/app/app"
	"github.com/dgoujard/uptimeWorker/app/router"
	"github.com/dgoujard/uptimeWorker/app/services"
	"github.com/dgoujard/uptimeWorker/config"
	dbConn "github.com/dgoujard/uptimeWorker/adapter/gorm"
	lr "github.com/dgoujard/uptimeWorker/util/logger"
	vr "github.com/dgoujard/uptimeWorker/util/validator"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	var appConfig = config.AppConfig()
	logger := lr.New(appConfig.Debug)

	db, err := dbConn.New(appConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
		return
	}
	databaseService := services.CreateDatabaseConnection(&appConfig.Database)
	queueService := services.CreateQueueService(&appConfig.Amq)
	awsService := services.CreateAwsService(&appConfig.Aws)
	uptimeCheckerService := services.CreateUptimeCheckerService()
	uptimeService := services.CreateUptimeService(&appConfig.Worker,uptimeCheckerService,awsService,queueService,databaseService)
	realtimeService := services.CreateRealtimeClient(&appConfig.Realtime)

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

	var asWebServer = false
	var srv http.Server
	if _, ok := cliOptions["withWebserver"]; ok {
		asWebServer = true
		validator := vr.New()
		application := app.New(logger, db, validator)
		appRouter := router.New(application)
		address := fmt.Sprintf(":%d", appConfig.ApiServer.Port)

		logger.Info().Msgf("Starting server %v", address)
		srv := &http.Server{
			Addr:         address,
			Handler:      appRouter,
			ReadTimeout:  appConfig.ApiServer.TimeoutRead.Duration,
			WriteTimeout: appConfig.ApiServer.TimeoutWrite.Duration,
			IdleTimeout:  appConfig.ApiServer.TimeoutIdle.Duration,
		}
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal().Err(err).Msg("Server startup failed")
			}
		}()

	}
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			if asWebServer{
				fmt.Printf("\nReceived an interrupt, Stoping webserver\n\n")
				ctx, cancel := context.WithTimeout(context.Background(), appConfig.ApiServer.TimeoutIdle.Duration)
				defer cancel()

				if err := srv.Shutdown(ctx); err != nil {
					logger.Warn().Err(err).Msg("Server shutdown failure")
				}

				sqlDB, err := db.DB()
				if err == nil {
					if err = sqlDB.Close(); err != nil {
						logger.Warn().Err(err).Msg("Db connection closing failure")
					}
				}

			}
			queueWorker.Close()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
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