package subscriber

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

import (
	"log"
	"math"
	"runtime"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	Generator "github.com/docker/docker/pkg/namesgenerator"
)

type connector struct {
	config  *config.Controller
	client  types.Invoker
	factory rabbitmq.QueueConsumerFactory

	subscribers []Subscriber
}

// Connector is the interface for the RabbitMQ Connector. Allowing for starting & stopping it.
type Connector interface {
	Start()
	End()
}

// NewConnector generates a connector using the provided parameters.
func NewConnector(config *config.Controller, client types.Invoker, factory rabbitmq.QueueConsumerFactory) Connector {
	return &connector{
		config:      config,
		client:      client,
		factory:     factory,
		subscribers: nil,
	}
}

func (m *connector) Start() {
	// Start Consumer
	m.spawnWorkers(m.config.Topics)

	for _, subscriber := range m.subscribers {
		_ = subscriber.Start()
	}
}

func (m *connector) End() {
	m.clearWorkers()
}

func (m *connector) spawnWorkers(topics []string) {
	amountOfTopics := len(topics)
	workerCount := CalculateWorkerCount(amountOfTopics)
	log.Printf("%d Topic(s) are registered. Will be spawning %d Worker(s) per Topic. ", amountOfTopics, workerCount)

	for _, topic := range topics {
		for i := 0; i < workerCount; i++ {
			consumer, err := m.factory.Build(topic)
			if err != nil {
				log.Printf("Failed to setup a worker for %s. Recieved %s", topic, err)
			} else {
				subscriber := NewSubscriber(Generator.GetRandomName(3), topic, consumer, m.client)
				m.subscribers = append(m.subscribers, subscriber)
			}
		}
	}
}

func (m *connector) clearWorkers() {
	for _, subscriber := range m.subscribers {
		if ok := subscriber.IsRunning(); ok {
			_ = subscriber.Stop()
		}
	}
	m.subscribers = nil
}

// CalculateWorkerCount will calculate the amount of workers that will be spawned
// based on the following formula: Available CPU's * 2 / Amount of Topics
func CalculateWorkerCount(amountOfTopics int) int {
	targetGoRoutines := float64(runtime.NumCPU() * 2)

	if int(targetGoRoutines) < amountOfTopics {
		return int(math.Floor(targetGoRoutines)/float64(amountOfTopics)) + 1
	}

	return int(math.Floor(targetGoRoutines) / float64(amountOfTopics))
}
