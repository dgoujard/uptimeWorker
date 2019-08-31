package services

import (
	"encoding/json"
	"github.com/dgoujard/uptimeWorker/config"
	"github.com/streadway/amqp"
	"log"
)

type QueueService struct  {
	amqCo *amqp.Connection
	amqCh *amqp.Channel
	amqQueueName string
}

func CreateQueueService(config *config.AmqConfig) *QueueService {
	connection, err := amqp.Dial(config.Uri)
	if err != nil {
		log.Println("Dial: %s", err)
	}

	if err != nil {
		log.Println(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	channel, err := connection.Channel()
	if err != nil {
		log.Println("Channel: %s", err)

	}
	return &QueueService{amqCo:connection,amqQueueName:config.QueueName,amqCh:channel}
}

func (q *QueueService)AddSiteToAmqQueue(site SiteBdd, isPriority bool) {
	jsonData, _ := json.Marshal(site)
	//log.Println(string(jsonData))

	err := q.amqCh.Publish("",q.amqQueueName,false,false,amqp.Publishing{
		Headers:         amqp.Table{},
		ContentType:     "application/json",
		ContentEncoding: "",
		Body:            jsonData,
		DeliveryMode:    amqp.Persistent, // 1=non-persistent, 2=persistent
		Priority:        0,              // 0-9
		// a bunch of application/implementation-specific fields
	})
	if err != nil {
		log.Println(err.Error())
		return
	}
}