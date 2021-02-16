/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package connector

import (
	"log"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
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

	failureChan, conErr := c.conManager.Connect(c.conf.RabbitConnectionURL)
	if conErr != nil {
		return conErr
	}

	go c.HandleConnectionError(failureChan)

	genErr := c.generateExchangesFrom(c.conf.Topology)
	if genErr != nil {
		return genErr
	}

	for _, ex := range c.exchanges {
		err := ex.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Connector) HandleConnectionError(ch <-chan *amqp.Error) {
	err := <-ch
	log.Printf("Rabbit MQ Connection failed with %s Code: %d [Server=%t Recover=%t]", err.Reason, err.Code, err.Server, err.Recover)

	if err.Recover {
		for _, ex := range c.exchanges {
			ex.Stop()
		}

		// Release old exchange refs to garbage collection
		c.exchanges = nil
		err := c.Run()
		if err != nil {
			log.Panicf("Received critical error: %s during restart, shutting down", err)
		}
	} else {
		log.Panicf("Received critical error: %s, shutting down", err)
	}
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

func (c *Connector) generateExchangesFrom(t types.Topology) error {
	// Do we want to use a connection per Exchange or continue with channels ?
	c.factory.WithChanCreator(c.conManager).WithInvoker(c.client)

	for _, topology := range c.conf.Topology {
		tmp := types.Exchange(topology)
		exchange, buildErr := c.factory.WithExchange(&tmp).Build()

		if buildErr != nil {
			return buildErr
		}

		c.exchanges = append(c.exchanges, exchange)
	}

	return nil
}
