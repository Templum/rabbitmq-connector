// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package subscriber

import (
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"log"
)

type subscriber struct {
	name  string
	topic string

	consumer rabbitmq.QueueConsumer
	client   types.Invoker

	active bool
}

type Subscriber interface {
	Start() error
	Stop() error

	IsRunning() bool
}

func NewSubscriber(name string, topic string, consumer rabbitmq.QueueConsumer, client types.Invoker) *subscriber {
	return &subscriber{
		name:  name,
		topic: topic,

		consumer: consumer,
		client:   client,

		active: false,
	}
}

func (s *subscriber) Start() error {
	invocations, err := s.consumer.Consume()
	if err != nil {
		log.Printf("Received %s during registering for messages on %s in Consumer[%s]", err, s.topic, s.name)
		return err
	}

	go func() {
		for invocation := range invocations {
			if s.topic == invocation.Topic {
				go func() {
					s.client.Invoke(s.topic, invocation.Message)
					invocation.Finished()
					// TEMP
					log.Printf("Message %s", *invocation.Message)
					log.Printf("Finished invocations of functions on topic %s", s.topic)
				}()
			}
		}
	}()

	log.Printf("Subscriber %s is now listening for messages on %s", s.name, s.topic)
	s.active = true
	return nil
}

func (s *subscriber) Stop() error {
	s.consumer.Stop()

	log.Printf("Subscriber %s stopped listening for messages on %s", s.name, s.topic)
	s.active = false
	return nil
}

func (s *subscriber) IsRunning() bool {
	return s.active
}
