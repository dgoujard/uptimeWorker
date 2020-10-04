package services

import (
	"encoding/json"
	"github.com/dgoujard/uptimeWorker/config"
	"github.com/streadway/amqp"
	"log"
)

type QueueService struct  {
	AmqCo *amqp.Connection
	AmqCh *amqp.Channel
	AmqQueueName string
	AmqAlertQueueName string
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
	return &QueueService{
		AmqCo:connection,
		AmqQueueName:config.QueueName,
		AmqAlertQueueName:config.QueueAlertName,
		AmqCh:channel,
	}
}

func (q *QueueService)AddSiteToAmqQueue(site SiteBdd) {
	jsonData, _ := json.Marshal(site)
	err := q.AmqCh.Publish("",q.AmqQueueName,false,false,amqp.Publishing{
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

func (q *QueueService)AddAlertToAmqQueue(alerte *Alerte,param interface{}) {
	if param != nil {
		jsonData, _ := json.Marshal(param)
		alerte.Param = jsonData
	}
	jsonData, _ := json.Marshal(alerte)

	err := q.AmqCh.Publish("",q.AmqAlertQueueName,false,false,amqp.Publishing{
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