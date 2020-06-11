/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package subscriber

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
	factory rabbitmq.ExchangeFactory

	subscribers []Subscriber
}

// Connector is the interface for the RabbitMQ Connector. Allowing for starting & stopping it.
type Connector interface {
	Start()
	End()
}

// NewConnector generates a connector using the provided parameters.
func NewConnector(config *config.Controller, client types.Invoker, factory rabbitmq.ExchangeFactory) Connector {
	return &connector{
		config:      config,
		client:      client,
		factory:     factory,
		subscribers: nil,
	}
}

func (m *connector) Start() {
	// Start Consumer
	m.spawnSubscribers(m.config.Topology)

	for _, subscriber := range m.subscribers {
		_ = subscriber.Start()
	}
}

func (m *connector) End() {
	m.clearWorkers()
}

func (m *connector) spawnSubscribers(topology types.Topology) {

	m.factory.WithConnection(m.config.RabbitConnectionURL, m.config.RabbitSanitizedURL)

	for _, exchange := range topology {
		subs, err := m.factory.WithExchange(exchange).Build()
		if err != nil {
			log.Printf("Failed to setup subscriber on exchange %s due to %s", exchange.Name, err)
		} else {
			subscriber := NewSubscriber(Generator.GetRandomName(3), topic, consumer, m.client)
			m.subscribers = append(m.subscribers, subscriber)
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
