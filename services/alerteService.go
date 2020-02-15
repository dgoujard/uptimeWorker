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
	Messageduree string
}

type AlerteService struct {
	AwsService          *AwsService
	Config              *config.AlertConfig
	Db		*DatabaseService
	Realtime *RealtimeService
}

func CreateAlerteService(config *config.AlertConfig, awsService *AwsService, databaseService *DatabaseService, realtime *RealtimeService) *AlerteService {
	return &AlerteService{
		Config:config,
		AwsService:awsService,
		Db:databaseService,
		Realtime: realtime,
	}
}
func generateUptimeEmailSubject(site *SiteBdd, isDown bool) (subject string) {
	if isDown{
		subject = "[Alerte Uptime DOWN] "+site.Name
	}else{
		subject = "[Alerte Uptime UP] "+site.Name
	}
	return
}
func generateUptimeEmailTemplate() (html string) {
	html = "<h2>Alerte {{.Site.Name}} {{ if .Param.IsCurrentlyDown }}DOWN{{ else }}UP{{end}}</h2>"
	html += "<ul>"
	html += "<li>Url : <a href='{{.Site.Url}}' target='_blank'>{{.Site.Url}}</a></li>"
	html += "<li>Code : {{.Param.LogSite.Code}} {{.Param.LogSite.Detail}}</li>"
	html += "<li>Event timestamp : {{readableDatetime .Param.LogSite.Datetime}}</li>"
	html += "{{ if not .Param.IsCurrentlyDown }}<li>Duration : {{.Messageduree}}</li>{{end}}"
	html += "</ul>"
	return
}
func generateSslExpireEmailTemplate() (html string) {
	html = "<h2>Alerte {{.Site.Name}} Ssl expiration</h2>"
	html += "<ul>"
	html += "<li>Url : <a href='{{.Site.Url}}' target='_blank'>{{.Site.Url}}</a></li>"
	html += "<li>SSL Expiration : {{readableDatetime .Site.SslExpireDatetime}}</li>"
	html += "</ul>"
	return
}
/* Exemple d'une alerte uptime
{"Site":{"_id":"5d39cf70a7f30900062f589f","Account":"5d15e76baf18e1087b9cc379","NotificationGroup":"5d87308befeb2c009b5aada7","createDatetime":1562681347,"Name":"Outil Navitia Kisio","Url":"https://api.navitia.io/v1/coverage/fr-cen/networks/network:Semtao/traffic_reports?start_page=0","Status":9,"uptimeId":783062088},"Type":"uptime","param":{"IsCurrentlyDown":true,"LogSite":{"_id":"5d876ce3303899a57f5cea93","datetime":1569156323,"Site":"5d39cf70a7f30900062f589f","Type":"5d15e76baf18e1087b9cc379","code":401,"Detail":"Unauthorized"}}}
 */
/* Exemple alerte SSL
{"Site":{"_id":"5df8da7c2aa418000d5fb355","Account":"5dd6c0ed857e052c8437fdaa","NotificationGroup":"5d87308befeb2c009b5aada7","createDatetime":1574262428,"lastlog":1576149564,"Name":"Site Transdev STD Gard (Digeek--)","Url":"https://www.stdgard.fr/","Status":2,"uptimeId":783847742,"ssl_monitored":true,"ssl_subject":"stdgard.fr","ssl_issuer":"Let's Encrypt Authority X3","ssl_algo":"SHA256-RSA","ssl_expireDatetime":1578358402},"Type":"sslExpire"}
 */
func (a *AlerteService) handleAlerteSSLExpireTask(alerteMessage *Alerte) {
	log.Println("Alerte a faire",alerteMessage.Site.Name," SSL Expiration")
	if len(a.Config.Realtimechannel)> 0 {
		var messageToRT = make(map[string]string)
		messageToRT["site_id"] = alerteMessage.Site.Id.Hex()
		messageToRT["site_name"] = alerteMessage.Site.Name
		messageToRT["site_url"] = alerteMessage.Site.Url
		messageToRT["type"] = "sslexpire"
		messageToRT["ssl_expiredate"] = strconv.Itoa(int(alerteMessage.Site.SslExpireDatetime))
		err := a.Realtime.Publish(a.Config.Realtimechannel,messageToRT)
		if err != nil {
			log.Println("Probleme publication realtime",err)
		}
	}
	if alerteMessage.Site.NotificationGroup.IsZero() {
		log.Println("Pas de cibles pour les alertes")
		return
	}
	tEmail, err := template.New("").Funcs(template.FuncMap{
		"readableDatetime": func(timestamp int32) string {
			loc, _ := time.LoadLocation("Europe/Paris")

			tm := time.Unix(int64(timestamp), 0)
			return tm.In(loc).Format("02/01/2006 15:04:05")
		},
	}).Parse(generateSslExpireEmailTemplate())
	if err != nil {
		log.Println("Template parsing: ", err)
	}
	tEmailVariable := alerteEmailVariablesTmpl{
		Site:alerteMessage.Site,
	}
	notificationgroup := a.Db.GetNotificationGroup(alerteMessage.Site.NotificationGroup.Hex())
	for _, cible := range notificationgroup.Cibles {
		switch cible.Type {
		case "email":

			var tpl bytes.Buffer
			err = tEmail.Execute(&tpl,tEmailVariable)
			if err != nil {
				log.Println("Template execution: ", err)
			}
			mailHtml := tpl.String()
			err = a.AwsService.SendEmail(a.Config.EmailFrom,
				cible.Cible,
				"[Alerte SSL Expire] "+alerteMessage.Site.Name,
				mailHtml,
				"",
			)
			if err != nil {
				log.Println("Erreur envoi email",err)
			}
		}
	}

}
func (a *AlerteService)handleAlerteUptimeTask(alerteMessage *Alerte)  {
	if alerteMessage.Param != nil {
		param := AlerteParamUptime{}
		if err := json.Unmarshal(alerteMessage.Param, &param); err != nil {
			log.Printf("param parsing error %s,", err.Error())
			return
		}
		log.Println("Alerte a faire",alerteMessage.Site.Name," Down? ",param.IsCurrentlyDown," Detail ",param.LogSite.Detail," TS ",param.LogSite.Datetime)
		//Notification serveur temps réel pour interface web
		if len(a.Config.Realtimechannel)> 0 {
			var messageToRT = make(map[string]string)
			messageToRT["site_id"] = alerteMessage.Site.Id.Hex()
			messageToRT["site_name"] = alerteMessage.Site.Name
			messageToRT["site_url"] = alerteMessage.Site.Url
			messageToRT["type"] = "uptime"
			if param.IsCurrentlyDown {
				messageToRT["isDown"] = "1"
			}else{
				messageToRT["isDown"] = "0"
			}
			messageToRT["detail"] = param.LogSite.Detail
			messageToRT["code"] = strconv.Itoa(param.LogSite.Code)
			messageToRT["datetime"] = strconv.Itoa(int(param.LogSite.Datetime))
			err := a.Realtime.Publish(a.Config.Realtimechannel,messageToRT)
			if err != nil {
				log.Println("Probleme publication realtime",err)
			}
		}

		//Notification si besoin à un groupe
		if alerteMessage.Site.NotificationGroup.IsZero() {
			log.Println("Pas de cibles pour les alertes")
			return
		}

		//Calcul du texte de durée de panne si message up
		var messageduree = ""
		if !param.IsCurrentlyDown {
			currentLogTime := time.Unix(param.LogSite.Datetime, 0)
			//Recherche de la derniére panne dans les logs du site
			lastLog := a.Db.GetLastSiteLog(alerteMessage.Site.Id,true)
			if lastLog != nil && lastLog.Datetime != 0 {
				previousLogTime := time.Unix(lastLog.Datetime, 0)
				year, month, day, hour, min, sec := timeDiff(previousLogTime,currentLogTime)
				if year > 0 {
					messageduree += " "+strconv.Itoa(year)+" ans"
				}
				if month > 0 {
					messageduree += " "+strconv.Itoa(month)+" mois"
				}
				if day > 0 {
					messageduree += " "+strconv.Itoa(day)+" jours"
				}
				if hour > 0 {
					messageduree += " "+strconv.Itoa(hour)+" heures"
				}
				if min > 0 {
					messageduree += " "+strconv.Itoa(min)+" minutes et"
				}
				if sec > 0 {
					messageduree += " "+strconv.Itoa(sec)+" secondes"
				}
			}
		}
		//Preparation de l'email pour les cibles email
		tEmail, err := template.New("").Funcs(template.FuncMap{
			"readableDatetime": func(timestamp int64) string {
				loc, _ := time.LoadLocation("Europe/Paris")

				tm := time.Unix(timestamp, 0)
				return tm.In(loc).Format("02/01/2006 15:04:05")
			},
		}).Parse(generateUptimeEmailTemplate())
		if err != nil {
			log.Println("Template parsing: ", err)
		}
		tEmailVariable := alerteEmailVariablesTmpl{
			Site:alerteMessage.Site,
			Param:&param,
			Messageduree: messageduree,
		}
		notificationgroup := a.Db.GetNotificationGroup(alerteMessage.Site.NotificationGroup.Hex())
		for _, cible := range notificationgroup.Cibles {
			switch cible.Type {
				case "email":

					var tpl bytes.Buffer
					err = tEmail.Execute(&tpl,tEmailVariable)
					if err != nil {
						log.Println("Template execution: ", err)
					}
					mailHtml := tpl.String()
					err = a.AwsService.SendEmail(a.Config.EmailFrom,
						cible.Cible,
						generateUptimeEmailSubject(alerteMessage.Site,param.IsCurrentlyDown),
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

func timeDiff(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}