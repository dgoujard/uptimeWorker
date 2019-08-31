package services

import "log"

type UptimeService struct {
	checker *uptimeCheckerService
}
func CreateUptimeService(uptimeChecker *uptimeCheckerService) *UptimeService {
	return &UptimeService{
		checker:uptimeChecker,
	}
}

func (u *UptimeService) CheckSite(site *SiteBdd){
	//TODO log des temps de réponse dans influxdb
	err, httpCode, responseTime := u.checker.CheckSite(site)
	if err != nil || httpCode != 200 {
		//TODO confirmer l'erreur en appelant une fonction lambda afin d'avoir un test depuis une autre localisation
		//TODO si site.Status n'est pas en erreur alors alerte a faire et enregistrement dans mongo
		log.Println(site.Url," DOWN ",httpCode," ",err,"(",responseTime,")")
	} else if err == nil && httpCode == 200 {
		//TODO si site.Status indique erreur alors que je n'ai plus d'erreur alors alerte à faire et enregistrement dans mongo
		log.Println(site.Url," up (",responseTime,")")
	}else{
		//TODO pas de changement entre l'état actuel et le dernier état enregistré en base
	}
}