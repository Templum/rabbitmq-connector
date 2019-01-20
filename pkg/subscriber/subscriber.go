package subscriber

import (
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"log"
)

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

type subscriber struct {
	name  string
	topic string

	consumer types.QueueConsumer
	client   types.Invoker

	active bool
}

type Subscriber interface {
	Start() error
	Stop() error

	IsRunning() bool
}

func NewSubscriber(name string, topic string, consumer types.QueueConsumer, client types.Invoker) *subscriber {
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
		println("Received %s during registering for messages on %s in Consumer[%s]", err, s.topic, s.name)
		return err
	}

	go func() {
		for invocation := range invocations {
			if s.topic == invocation.Topic {
				s.client.Invoke(s.topic, invocation.Message)
				invocation.Finished()
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
