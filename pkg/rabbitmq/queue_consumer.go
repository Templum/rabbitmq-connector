package rabbitmq

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

import (
	"log"

	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
)

type queueConsumer struct {
	queueName string
	channel   *amqp.Channel
}

// QueueConsumer is an Consumer that listens on a specified queue and forwards incoming messages.
type QueueConsumer interface {
	Consume() (<-chan *types.OpenFaaSInvocation, error)
	Stop()
	ListenForErrors() <-chan *amqp.Error
}

// NewQueueConsumer creates a new instance of QueueConsumer and assigns the passed channel to it.
func NewQueueConsumer(queueName string, channel *amqp.Channel) QueueConsumer {
	return &queueConsumer{queueName: queueName, channel: channel}
}

func (c *queueConsumer) Consume() (<-chan *types.OpenFaaSInvocation, error) {
	ch, err := c.channel.Consume(
		c.queueName,
		"", // Let Rabbit MQ Generate a Tag
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	out := make(chan *types.OpenFaaSInvocation)

	go func() {
		for message := range ch {
			out <- types.NewInvocation(message)
		}
	}()

	return out, nil
}

func (c *queueConsumer) Stop() {
	err := c.channel.Close()
	if err != nil {
		log.Printf("Received %s during close", err)
	}
}

func (c *queueConsumer) ListenForErrors() <-chan *amqp.Error {
	errors := make(chan *amqp.Error)
	c.channel.NotifyClose(errors)

	return errors
}
