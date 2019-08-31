package services

import (
	"github.com/dgoujard/uptimeWorker/config"
	"log"
)

type UptimeService struct {
	checker *uptimeCheckerService
	awsService *AwsService
	config *config.WorkerConfig
}
func CreateUptimeService(config *config.WorkerConfig,uptimeChecker *uptimeCheckerService, awsService *AwsService) *UptimeService {
	return &UptimeService{
		checker:uptimeChecker,
		awsService: awsService,
		config:config,
	}
}

func (u *UptimeService) CheckSite(site *SiteBdd){
	//TODO log des temps de réponse dans influxdb
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

		//TODO si site.Status n'est pas en erreur alors alerte a faire et enregistrement dans mongo
		log.Println(site.Url," DOWN ",result.HttpCode," ",result.Err ,"(",result.Duration,")")
		return
	} else if result.Err  == "" && result.HttpCode == 200 {
		//TODO si site.Status indique erreur alors que je n'ai plus d'erreur alors alerte à faire et enregistrement dans mongo
		log.Println(site.Url," up (",result.Duration,")")
		return
	}
}