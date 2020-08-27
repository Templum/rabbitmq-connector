/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package connector

import (
	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"log"
)

type RabbitToOpenFaaS interface {
	Run() error
	Shutdown()
}

func New(manager rabbitmq.Manager, factory rabbitmq.Factory, invoker types.Invoker, conf *config.Controller) RabbitToOpenFaaS {
	return &Connector{
		client: invoker,

		factory:    factory,
		conManager: manager,
		conf:       conf,
	}
}

type Connector struct {
	client types.Invoker

	factory    rabbitmq.Factory
	conManager rabbitmq.Manager
	conf       *config.Controller
	exchanges  []rabbitmq.ExchangeOrganizer
}

func (c *Connector) Run() error {
	log.Println("Started RabbitMQ <=> OpenFaaS Connector")

	log.Printf("Will now establish connection to %s", c.conf.RabbitSanitizedURL)
	conErr := c.conManager.Connect(c.conf.RabbitConnectionURL)

	if conErr != nil {
		return conErr
	}

	c.factory.WithConnection(c.conManager)

	for _, topology := range c.conf.Topology {
		exchange, err := c.factory.WithExchange(types.Exchange(topology)).Build()

		if err != nil {
			return err
		}

		c.exchanges = append(c.exchanges, exchange)
	}

	for _, ex := range c.exchanges {
		err := ex.Start(c.client)

		if err != nil {
			log.Printf("Received %s", err)
		}
	}

	return nil
}

func (c *Connector) Shutdown() {
	log.Println("Shutdown RabbitMQ <=> OpenFaaS Connector")

	// Loop over Exchanges to close
	for _, ex := range c.exchanges {
		ex.Stop()
	}

	// Close Connection
	c.conManager.Disconnect()
}
