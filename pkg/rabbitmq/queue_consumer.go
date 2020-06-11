/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
	"log"
)

type topicSubscriber struct {
	ch          *amqp.Channel
	targetQueue string
}

func NewTopicSubscriber(ch *amqp.Channel, targetQueue string) Subscriber {
	return &topicSubscriber{
		ch:          ch,
		targetQueue: targetQueue,
	}
}

type Subscriber interface {
	Subscribe() (<-chan types.OpenFaaSInvocation, error)
	Stop()
}

func (s *topicSubscriber) Subscribe() (<-chan types.OpenFaaSInvocation, error) {
	incoming, err := s.ch.Consume(
		s.targetQueue,
		"",
		true,
		false,
		false,
		false,
		nil)

	if err != nil {
		return nil, err
	}

	out := make(chan types.OpenFaaSInvocation)
	errors := make(chan *amqp.Error)
	s.ch.NotifyClose(errors)

	go func() {
		for message := range incoming {
			out <- types.NewInvocation(message)
		}
		close(out)
	}()

	go func() {
		for received := range errors {
			if received.Recover {
				log.Printf("Received non critical error %s.", received)
			} else {
				log.Printf("Received critical error %s.", received)
			}
		}
	}()

	return out, nil
}

func (s *topicSubscriber) Stop() {
}

// https://golang.hotexamples.com/de/examples/github.com.streadway.amqp/Connection/NotifyClose/golang-connection-notifyclose-method-examples.html
