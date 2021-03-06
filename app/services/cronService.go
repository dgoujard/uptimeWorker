package services

import (
	"time"
)

type cronService struct {
	queueService *QueueService
	databaseService *DatabaseService
}

func CreateCronService(database *DatabaseService, queue *QueueService) *cronService {
	return &cronService{
		queueService:    queue,
		databaseService: database,
	}
}
func (c *cronService)StartCronProcess() {
	ticker := time.NewTicker(60 * time.Second)
	//TODO faire cron qui toutes les heures pour regarder état de la date expiration SSL
	//TODO si < 7 jours et que date dernier email alerte (todo ajout champ) > 7 jours date expiration alors alerte
	go func() {
		c.cronAddSitesToQueue()
		for t := range ticker.C {
			_ = t // we don't print the ticker time, so assign this `t` variable to underscore `_` to avoid error
			c.cronAddSitesToQueue()
		}
	}()
}

func (c *cronService)cronAddSitesToQueue() {
	//TODO peut etre controllé la taille de la queue. S'il reste encore des checks à faire alors il n'y a pas assez de workers il faut pas ajouté les controles + faire un mail
	liste := c.databaseService.GetSitesList(false)
	for _, site := range liste {
		//TODO selon la config du site et le minute en cours ajouté dans les check ou pas
		c.queueService.AddSiteToAmqQueue(site)
	}
}