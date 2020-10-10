package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dgoujard/uptimeWorker/app/services"
)



func HandleLambdaEvent(site services.SiteBdd) (services.CheckSiteResponse, error) {
	service := services.CreateUptimeCheckerService()
	response := service.CheckSite(&site)
	return response, nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}