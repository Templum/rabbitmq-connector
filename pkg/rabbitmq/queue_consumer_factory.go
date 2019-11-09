package rabbitmq

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

import (
	"fmt"
	"log"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/streadway/amqp"
)

type queueConsumerFactory struct {
	con    *amqp.Connection
	config *config.Controller
}

// QueueConsumerFactory will build QueueConsumer for the given topic.
type QueueConsumerFactory interface {
	Build(topic string) (QueueConsumer, error)
}

// NewQueueConsumerFactory Initializes a new factory using the provided config.
func NewQueueConsumerFactory(config *config.Controller) (QueueConsumerFactory, error) {
	factory := queueConsumerFactory{
		con:    nil,
		config: config,
	}

	con, err := factory.establishConnection(config.RabbitConnectionURL, 5)
	if err != nil {
		log.Printf("Failed to establish a connection to %s. Last Recieved error is %s", config.RabbitSanitizedURL, err)
		return nil, err
	}

	factory.con = con
	return &factory, nil
}

func (f *queueConsumerFactory) Build(topic string) (QueueConsumer, error) {
	var err error
	ch, err := f.establishChannel(5)
	if err != nil {
		return nil, err
	}

	err = f.declareTopology(ch, topic)
	if err != nil {
		return nil, err
	}

	return NewQueueConsumer(generateQueueName(topic), ch), nil
}

func (f *queueConsumerFactory) establishConnection(connectionURL string, retries uint) (*amqp.Connection, error) {
	con, err := amqp.Dial(connectionURL)

	if err != nil && retries > 0 {
		log.Printf("Failed to establish connection due to %s. %d tries left", err, retries)
		time.Sleep(5 * time.Second)
		return f.establishConnection(connectionURL, retries-1)
	}

	return con, err
}

func (f *queueConsumerFactory) establishChannel(retries uint) (*amqp.Channel, error) {
	channel, err := f.con.Channel()

	if err != nil && retries > 0 {
		log.Printf("The attempt to open a Channel failed with %s. Retries left %d", err, retries)
		time.Sleep(2 * time.Second)
		return f.establishChannel(retries - 1)
	}
	return channel, err
}

func (f *queueConsumerFactory) declareTopology(c *amqp.Channel, topic string) error {
	var err error
	cfg := f.config
	queueName := generateQueueName(topic)

	err = c.ExchangeDeclare(
		cfg.ExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	_, err = c.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	err = c.QueueBind(
		queueName,
		topic,
		cfg.ExchangeName,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	log.Printf("Binding Queue %s to Exchange %s for Topic: %s", queueName, cfg.ExchangeName, topic)
	return nil
}

// generateQueueName will return the QueueName which consists of a prefix and the topic
func generateQueueName(topic string) string {
	const PreFix = "OpenFaaS"
	return fmt.Sprintf("%s_%s", PreFix, topic)
}
