package subscriber

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

import (
	"log"

	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/types"
)

type subscriber struct {
	name  string
	topic string

	consumer rabbitmq.QueueConsumer
	client   types.Invoker

	active bool
}

// Subscriber subscribes to an topic on Rabbit MQ. It can also be seen as a worker.
type Subscriber interface {
	Start() error
	Stop() error

	IsRunning() bool
}

// NewSubscriber setups a new subscriber for the given topic.
func NewSubscriber(name string, topic string, consumer rabbitmq.QueueConsumer, client types.Invoker) Subscriber {
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
		errors := s.consumer.ListenForErrors()

		for received := range errors {
			if received.Recover {
				log.Printf("Received non critical error %s. Will try to recover worker %s", received, s.name)
				_ = s.Stop()
				_ = s.Start()
			} else {
				log.Printf("Received critical error %s. Shutting down Consumer %s", received, s.name)
				_ = s.Stop()
			}
		}
	}()

	go func() {
		for invocation := range invocations {
			if s.topic == invocation.Topic {
				go func() {
					s.client.Invoke(s.topic, *invocation.Message)
					log.Printf("Finished invocations of functions on topic %s", s.topic)
				}()
			} else {
				log.Printf("Should not happen, but subscriber topic - %s is not equals to invocation topic - %s", s.topic, invocation.Topic)
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
