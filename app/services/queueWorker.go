package services

import (
	"encoding/json"
	"fmt"
	"github.com/dgoujard/uptimeWorker/config"
	"github.com/streadway/amqp"
	"log"
)

type QueueWorker struct {
	queueService *QueueService
	amqCo *amqp.Connection
	amqUptimeQueueName string
	amqAlerteQueueName string
	amqConcurentRuntime int
	uptimeService *UptimeService
	alerteService *AlerteService
}

func CreateQueueWorker(config *config.AmqConfig, queueService *QueueService, uptime *UptimeService,alerte *AlerteService) *QueueWorker {
	connection, err := amqp.Dial(config.Uri)
	if err != nil {
		log.Println("Dial: %s", err)
	}

	if err != nil {
		log.Println(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}

	return &QueueWorker{
		amqCo:connection,
		amqUptimeQueueName:config.QueueName,
		amqAlerteQueueName:config.QueueAlertName,
		amqConcurentRuntime:config.ConcurentRuntime,
		queueService: queueService,
		uptimeService: uptime,
		alerteService:alerte,
	}
}

func (q *QueueWorker) StartAmqWatching() {
	channel, err := q.amqCo.Channel()
	if err != nil {
		log.Println("Channel: %s", err)
	}
	defer channel.Close()
	defer q.amqCo.Close()

	q.listenUptimeChannel(channel)
	if q.alerteService != nil {
		q.listenAlerteChannel(channel)
	}
}
func (q *QueueWorker) Close(){
	//Nothing todo
	fmt.Printf("\nReceived an interrupt, Closing subscriptions ?\n\n")
}


func (q *QueueWorker) listenUptimeChannel(channel *amqp.Channel)  {
	msgs, err := channel.Consume(
		q.amqUptimeQueueName, // queue
		"",             // consumer
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Println("fail consume: %s", err)
	}
	for i := 0; i < q.amqConcurentRuntime; i++ {
		go func(runtimeIndex int) {
			for d := range msgs {
				//log.Println("Runtime ID: " + strconv.Itoa(runtimeIndex))
				//log.Println(string(d.Body))

				site := &SiteBdd{}
				if err := json.Unmarshal(d.Body, &site); err != nil {
					log.Printf("TaskMessage json not valid  %v.", err)
				}
				q.uptimeService.CheckSite(site)
				//TODO créer/utiliser le service uptimeService
				d.Ack(false)
			}
		}(i)
	}
}


func (q *QueueWorker) listenAlerteChannel(channel *amqp.Channel)  {
	msgs, err := channel.Consume(
		q.amqAlerteQueueName, // queue
		"",             // consumer
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Println("fail consume: %s", err)
	}
	for i := 0; i < q.amqConcurentRuntime; i++ {
		go func(runtimeIndex int) {
			for d := range msgs {
				//log.Println("Runtime ID Alerte: " + strconv.Itoa(runtimeIndex))
				//log.Println(string(d.Body))

				alertMessage := &Alerte{}
				if err := json.Unmarshal(d.Body, &alertMessage); err != nil {
					log.Printf("alertMessage json not valid  %v.", err)
				}
				switch alertMessage.Type {
				case "uptime":
					q.alerteService.handleAlerteUptimeTask(alertMessage)
				case "sslExpire":
					q.alerteService.handleAlerteSSLExpireTask(alertMessage)
				}

				d.Ack(false)
			}
		}(i)
	}
}
