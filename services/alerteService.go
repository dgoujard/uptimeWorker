package services

import (
	"bytes"
	"encoding/json"
	"github.com/dgoujard/uptimeWorker/config"
	"html/template"
	"log"
	"strconv"
	"time"
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
type alerteEmailVariablesTmpl struct {
	Site *SiteBdd
	Param *AlerteParamUptime
}

type AlerteService struct {
	AwsService          *AwsService
	config              *config.AlertConfig
	db		*DatabaseService
	realtime *RealtimeService
}

func CreateAlerteService(config *config.AlertConfig, awsService *AwsService, databaseService *DatabaseService, realtime *RealtimeService) *AlerteService {
	return &AlerteService{
		config:config,
		AwsService:awsService,
		db:databaseService,
		realtime: realtime,
	}
}
func generateEmailSubject(site *SiteBdd,log *LogBdd) (subject string) {
	if(log.Code != 200 || log.Detail != ""){
		subject = "[Alerte Uptime DOWN] "+site.Name
	}else{
		subject = "[Alerte Uptime UP] "+site.Name
	}
	return
}
func generateEmailTemplate() (html string) {
	html = "<h2>Alerte {{.Site.Name}} {{ if .Param.IsCurrentlyDown }}DOWN{{ else }}UP{{end}}</h2>"
	html += "<ul>"
	html += "<li>Url : <a href='{{.Site.Url}}' target='_blank'>{{.Site.Url}}</a></li>"
	html += "<li>Code : {{.Param.LogSite.Code}} {{.Param.LogSite.Detail}}</li>"
	html += "<li>Event timestamp : {{readableDatetime .Param.LogSite.Datetime}}</li>"
	html += "</ul>"
	return
}
/* Exemple d'une alerte
{"Site":{"_id":"5d39cf70a7f30900062f589f","Account":"5d15e76baf18e1087b9cc379","NotificationGroup":"5d87308befeb2c009b5aada7","createDatetime":1562681347,"Name":"Outil Navitia Kisio","Url":"https://api.navitia.io/v1/coverage/fr-cen/networks/network:Semtao/traffic_reports?start_page=0","Status":9,"uptimeId":783062088},"Type":"uptime","param":{"IsCurrentlyDown":true,"LogSite":{"_id":"5d876ce3303899a57f5cea93","datetime":1569156323,"Site":"5d39cf70a7f30900062f589f","Type":"5d15e76baf18e1087b9cc379","code":401,"Detail":"Unauthorized"}}}
 */
func (a *AlerteService)handleAlerteUptimeTask(alerteMessage *Alerte)  {
	if alerteMessage.Param != nil {
		param := AlerteParamUptime{}
		if err := json.Unmarshal(alerteMessage.Param, &param); err != nil {
			log.Printf("param parsing error %s,", err.Error())
			return
		}
		log.Println("Alerte a faire",alerteMessage.Site.Name," Down? ",param.IsCurrentlyDown," Detail ",param.LogSite.Detail," TS ",param.LogSite.Datetime)
		if len(a.config.Realtimechannel)> 0 {
			var messageToRT = make(map[string]string)
			messageToRT["site_id"] = alerteMessage.Site.Id.Hex()
			messageToRT["site_name"] = alerteMessage.Site.Name
			messageToRT["site_url"] = alerteMessage.Site.Url
			if param.IsCurrentlyDown {
				messageToRT["isDown"] = "1"
			}else{
				messageToRT["isDown"] = "0"
			}
			messageToRT["detail"] = param.LogSite.Detail
			messageToRT["code"] = strconv.Itoa(param.LogSite.Code)
			messageToRT["datetime"] = strconv.Itoa(int(param.LogSite.Datetime))
			err := a.realtime.Publish(a.config.Realtimechannel,messageToRT)
			if err != nil {
				log.Println("Probleme publication realtime",err)
			}
		}

		if alerteMessage.Site.NotificationGroup.IsZero() {
			log.Println("Pas de cibles pour les alertes")
			return
		}

		tEmail, err := template.New("").Funcs(template.FuncMap{
			"readableDatetime": func(timestamp int64) string {
				loc, _ := time.LoadLocation("Europe/Paris")

				tm := time.Unix(timestamp, 0)
				return tm.In(loc).Format("02/01/2006 15:04:05")
			},
		}).Parse(generateEmailTemplate())
		if err != nil {
			log.Println("Template parsing: ", err)
		}
		tEmailVariable := alerteEmailVariablesTmpl{
			Site:alerteMessage.Site,
			Param:&param,
		}
		notificationgroup := a.db.GetNotificationGroup(alerteMessage.Site.NotificationGroup.Hex())
		for _, cible := range notificationgroup.Cibles {
			switch cible.Type {
				case "email":

					var tpl bytes.Buffer
					err = tEmail.Execute(&tpl,tEmailVariable)
					if err != nil {
						log.Println("Template execution: ", err)
					}
					mailHtml := tpl.String()
					err = a.AwsService.SendEmail(a.config.EmailFrom,
						cible.Cible,
						generateEmailSubject(alerteMessage.Site,param.LogSite),
						mailHtml,
						"",
					)
					if err != nil {
						log.Println("Erreur envoi email",err)
					}
			}
		}
	}
}