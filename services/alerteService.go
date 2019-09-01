package services

import (
	"encoding/json"
	"github.com/dgoujard/uptimeWorker/config"
	"log"
)


type AlerteParamUptime struct {
	ResultUptime *CheckSiteResponse
	IsCurrentlyDown bool
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

func (a *AlerteService)handleAlerteUptimeTask(alerteMessage *Alerte)  {
	if alerteMessage.Param != nil {
		param := AlerteParamUptime{}
		if err := json.Unmarshal(alerteMessage.Param, &param); err != nil {
			log.Printf("param parsing error %s,", err.Error())
			return
		}
		log.Println("Alerte a faire",alerteMessage.Site.Name," Down? ",param.IsCurrentlyDown)
		//TODO faire alerte par Email ou autre selon la configuration du site
	}
}