package services

import (
	"encoding/json"
	"fmt"
	"github.com/dgoujard/uptimeWorker/config"
	"github.com/streadway/amqp"
	"log"
	"os"
	"os/signal"
)

type QueueWorker struct {
	queueService *QueueService
	amqCo *amqp.Connection
	amqQueueName string
	amqConcurentRuntime int
	uptimeService *UptimeService
}

func CreateQueueWorker(config *config.AmqConfig, queueService *QueueService, uptime *UptimeService) *QueueWorker {
	connection, err := amqp.Dial(config.Uri)
	if err != nil {
		log.Println("Dial: %s", err)
	}

	if err != nil {
		log.Println(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}

	return &QueueWorker{
		amqCo:connection,
		amqQueueName:config.QueueName,
		amqConcurentRuntime:config.ConcurentRuntime,
		queueService: queueService,
		uptimeService: uptime,
	}
}

func (q *QueueWorker) StartAmqWatching() {
	channel, err := q.amqCo.Channel()
	if err != nil {
		log.Println("Channel: %s", err)
	}
	msgs, err := channel.Consume(
		q.amqQueueName, // queue
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
	defer channel.Close()
	defer q.amqCo.Close()
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

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Printf("\nReceived an interrupt, unsubscribing and closing connection...\n\n")
			// Do not unsubscribe a durable on exit, except if asked to.
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}