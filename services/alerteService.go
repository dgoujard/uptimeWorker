package services

import (
	"encoding/json"
	"github.com/dgoujard/uptimeWorker/config"
	"log"
)


type AlerteParamUptime struct {
	IsCurrentlyDown bool
	LogSite *LogBdd
}

type Alerte struct {
	Site *SiteBdd
	Type string
	Param json.RawMessage `json:"param,omitempty"`
}

type AlerteService struct {
	AwsService          *AwsService
	config              *config.AlertConfig
}

func CreateAlerteService(config *config.AlertConfig, awsService *AwsService) *AlerteService {
	return &AlerteService{
		config:config,
		AwsService:awsService,
	}
}
/* Exemple d'une alerte
{"Site":{"_id":"5d39cf70a7f30900062f589f","Account":"5d15e76baf18e1087b9cc379","createDatetime":1562681347,"Name":"Outil Navitia Kisio","Url":"https://api.navitia.io/v1/coverage/fr-cen/networks/network:Semtao/traffic_reports?start_page=0","Status":9,"uptimeId":783062088},"Type":"uptime","param":{"IsCurrentlyDown":true,"LogSite":{"_id":"5d868336c824927a742b8e15","datetime":1569096502,"Site":"5d39cf70a7f30900062f589f","Type":"5d15e76baf18e1087b9cc379","code":401,"Detail":"Unauthorized"}}}
 */
func (a *AlerteService)handleAlerteUptimeTask(alerteMessage *Alerte)  {
	if alerteMessage.Param != nil {
		param := AlerteParamUptime{}
		if err := json.Unmarshal(alerteMessage.Param, &param); err != nil {
			log.Printf("param parsing error %s,", err.Error())
			return
		}
		log.Println("Alerte a faire",alerteMessage.Site.Name," Down? ",param.IsCurrentlyDown," Detail ",param.LogSite.Detail," TS ",param.LogSite.Datetime)
		//TODO faire alerte par Email ou autre selon la configuration du site
	}
}