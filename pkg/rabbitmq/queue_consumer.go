// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package rabbitmq

import (
	"log"

	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
)

type queueConsumer struct {
	queueName string
	channel   *amqp.Channel
}

type QueueConsumer interface {
	Consume() (<-chan *types.OpenFaaSInvocation, error)
	Stop()
	ListenForErrors() <-chan error
}

func NewQueueConsumer(channel *amqp.Channel) QueueConsumer {
	return &queueConsumer{channel: channel}
}

func (c *queueConsumer) Consume() (<-chan *types.OpenFaaSInvocation, error) {
	ch, err := c.channel.Consume(
		c.queueName,
		"", // Let Rabbit MQ Generate a Tag
		false,
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
		log.Printf("Recieved %s during close", err)
	}
}

func (c *queueConsumer) ListenForErrors() <-chan error {
	return nil
}
