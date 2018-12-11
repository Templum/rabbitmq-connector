// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package rabbitmq

import (
	"log"
	"math"
	"runtime"
	"strings"
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
}

func MakeConnector(uri string, client *types.Controller) Connector {
	return &connector{
		uri,
		false,

		nil,
		client,

		nil,
	}
}

type Recoverer interface {
	recover(receivedError *amqp.Error)
}

type Connector interface {
	Start()
	Close()
	Recoverer
}

func (c *connector) Start() {
	log.Println("Starting Init Process of Connector")
	var err error
	c.con, err = connectToRabbitMQ(c.uri, 3)
	if err != nil {
		log.Panicf("Failed to connect to %s after 3 retries", SanitizeConnectionUri(c.uri))
	}

	log.Printf("Successfully established connection with %s", SanitizeConnectionUri(c.uri))

	// Related to Self Healing
	oneTimeErrorChannel := make(chan *amqp.Error) // Maybe switch later to an struct holding the channel
	c.con.NotifyClose(oneTimeErrorChannel)
	go Healer(c, oneTimeErrorChannel)

	c.spawnWorkers(config.GetTopics())

	log.Println("Connector finished Init Process and is now running")
}

func (c *connector) recover(receivedError *amqp.Error) {
	if receivedError.Recover {
		log.Printf("Performing a recovery from following recoverable error: [Status: %d Reason: %s]", receivedError.Code, receivedError.Reason)
	} else {
		log.Panicf("Recieved unrecovarable error: [Status: %d Reason: %s]", receivedError.Code, receivedError.Reason)
	}

	log.Println("Will now clear worker pool")
	c.clearWorkers()

	c.Start()
}

func (c *connector) Close() {
	if !c.closed {
		log.Println("Shutting down Connector")
		c.closed = true
		defer c.con.Close()
		log.Println("Clearing worker pool")
		c.clearWorkers()
	}
}

// spawnWorkers will spawn Workers per topic based on the available resources. Further it
// assigns the newly created worker to the worker pool before starting them.
func (c *connector) spawnWorkers(topics []string) {
	amountOfTopics := len(topics)
	workerCount := CalculateWorkerCount(amountOfTopics)
	log.Printf("%d Topics are registered. Will be spawning %d per Topic. ", amountOfTopics, workerCount)

	for _, topic := range topics {
		for i := 0; i < workerCount; i++ {
			worker := NewWorker(c.con, c.client, topic)
			c.workers = append(c.workers, worker)
			go worker.Start()
		}
	}
}

// clearWorkers will stop Workers from the worker pool and afterwards cleanup the references
func (c *connector) clearWorkers() {
	for _, worker := range c.workers {
		worker.Close()
	}
	c.workers = nil
}

// Helper Functions

// TODO: Documentation & Eventuell Type für Connection (AMK)
func connectToRabbitMQ(uri string, retries int) (*amqp.Connection, error) {
	con, err := amqp.Dial(uri)

	if err != nil && retries > 0 {
		log.Printf("Failed to connect to %s with error %s. Retries left %d", SanitizeConnectionUri(uri), err, retries)
		time.Sleep(5 * time.Second)
		return connectToRabbitMQ(uri, retries-1)
	}

	return con, err
}

// Healer will take the reference to something "Recoverable"  along with an error channel for amqp errors.
// The method is only meant for one time usage.
func Healer(c Recoverer, errorStream chan *amqp.Error) {
	err := <-errorStream
	// Give it some time before beginning recovering
	time.Sleep(30 * time.Second)
	c.recover(err)
}

// CalculateWorkerCount will calculate the amount of workers that will be spawned
// based on the following formula: Available CPU's * 2 / Amount of Topics
func CalculateWorkerCount(amountOfTopics int) int {
	targetGoRoutines := float64(runtime.NumCPU() * 2)

	if int(targetGoRoutines) < amountOfTopics {
		return int(math.Floor(targetGoRoutines)/float64(amountOfTopics)) + 1
	} else {
		return int(math.Floor(targetGoRoutines) / float64(amountOfTopics))
	}

}

// SanitizeConnectionUri takes an uri in the format user:pass@domain and removes the sensitive credentials.
// Prior it will perform a check for @, if it is not included it will return the unchanged uri
func SanitizeConnectionUri(uri string) string {
	const SEPARATOR = "@"

	if idx := strings.Index(uri, SEPARATOR); idx != -1 {
		return strings.Split(uri, SEPARATOR)[1]
	} else {
		return uri
	}
}
