package services

import (
	"time"
)

type cronService struct {
	queueService *QueueService
	databaseService *DatabaseService
}

func CreateCronService(database *DatabaseService,queue *QueueService) *cronService {
	return &cronService{
		queueService:    queue,
		databaseService: database,
	}
}
func (c *cronService)StartCronProcess() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		c.cronAddSitesToQueue()
		for t := range ticker.C {
			_ = t // we don't print the ticker time, so assign this `t` variable to underscore `_` to avoid error
			c.cronAddSitesToQueue()
		}
	}()
}

func (c *cronService)cronAddSitesToQueue()  {
	liste := c.databaseService.GetSitesLis()
	for _, site := range liste {
		//TODO selon la config du site et le minute en cours ajouté dans les check ou pas
		c.queueService.AddSiteToAmqQueue(site,false)
	}
}