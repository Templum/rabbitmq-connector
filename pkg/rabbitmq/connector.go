// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package rabbitmq

import (
	"log"
	"math"
	"runtime"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/streadway/amqp"
)

type connector struct {
	uri    string
	closed bool

	con    *amqp.Connection
	client *types.Controller

	// Consumers
	workers []*worker

	// Sig Channel
	errorChannel chan *amqp.Error
}

func MakeConnector(uri string, client *types.Controller) *connector {
	return &connector{
		uri,
		false,

		nil,
		client,

		nil,
		nil,
	}
}

func (c *connector) Start() {
	log.Println("Starting Connector")

	c.init()
}

func (c *connector) Close() {
	if !c.closed {
		log.Println("Shutting down Connector")
		c.closed = true
		defer c.con.Close()

		for _, worker := range c.workers {
			worker.Close()
		}
		c.workers = nil
	}
}

func (c *connector) init() {
	var err error
	c.con, err = connectToRabbitMQ(c.uri, 3)
	if err != nil {
		log.Panicf("Failed to connect to %s, recieved %s", c.uri, err)
	} else {
		log.Printf("Successfully connected to %s", c.uri)
	}

	// Related to Self Healing
	c.errorChannel = make(chan *amqp.Error)
	c.con.NotifyClose(c.errorChannel)
	go c.registerSelfHealer() // Leaking Go Routine ?

	// Queues: 1 Topic === 1 Queue
	topics := config.GetTopics()
	amountOfTopics := len(topics)
	for _, topic := range topics {
		workerCount := int(math.Round(float64(runtime.NumCPU()*2)/float64(amountOfTopics))) + 1
		log.Printf("Spawning %d Workers for Topic: %s", workerCount, topic)

		for i := 0; i < workerCount; i++ {
			worker := NewWorker(c.con, c.client, topic)
			worker.Start()
			// TODO: Maybe add Thread Lock here
			c.workers = append(c.workers, worker)
		}
	}

}

func (c *connector) registerSelfHealer() {
	for {
		err := <-c.errorChannel
		if !c.closed {
			log.Printf("Recieved following error %s", err)

			time.Sleep(30 * time.Second)
			c.recover()
		}
	}
}

func (c *connector) recover() {
	log.Printf("Performing a Recovery")

	for _, topicQueue := range c.workers {
		topicQueue.Close()
	}
	c.workers = nil

	c.init()
}

func connectToRabbitMQ(uri string, retries int) (*amqp.Connection, error) {
	con, err := amqp.Dial(uri)

	if err != nil && retries > 0 {
		log.Printf("Failed to connect to %s with error %s. Retries left %d", uri, err, retries)
		time.Sleep(5 * time.Second)
		return connectToRabbitMQ(uri, retries-1)
	}

	return con, err
}
