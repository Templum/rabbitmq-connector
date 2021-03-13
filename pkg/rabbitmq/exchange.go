/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"log"
	"sync"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
)

type Starter interface {
	Start() error
}

type Stopper interface {
	Stop()
}

type ExchangeOrganizer interface {
	Starter
	Stopper
}

type Exchange struct {
	channel ChannelConsumer
	client  types.Invoker

	definition *types.Exchange
	lock       sync.RWMutex
}

const MAX_ATTEMPTS = 3

func NewExchange(channel ChannelConsumer, client types.Invoker, definition *types.Exchange) ExchangeOrganizer {
	return &Exchange{
		channel: channel,
		client:  client,

		definition: definition,
		lock:       sync.RWMutex{},
	}
}

func (e *Exchange) Start() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	closeChannel := make(chan *amqp.Error)
	e.channel.NotifyClose(closeChannel)
	go e.handleChanFailure(closeChannel)

	for _, topic := range e.definition.Topics {
		queueName := GenerateQueueName(e.definition.Name, topic)
		deliveries, err := e.channel.Consume(queueName, "", false, false, false, false, amqp.Table{})
		if err != nil {
			return err
		}

		go e.StartConsuming(topic, deliveries)
	}

	return nil
}

func (e *Exchange) Stop() {
	e.lock.Lock()
	defer e.lock.Unlock()

	// We ignore the issue since this method is usually called after connection failure.
	_ = e.channel.Close()
}

func (e *Exchange) handleChanFailure(ch <-chan *amqp.Error) {
	err := <-ch
	log.Printf("Received following error %s on channel for exchange %s", err, e.definition.Name)
}

func (e *Exchange) StartConsuming(topic string, deliveries <-chan amqp.Delivery) {
	for delivery := range deliveries {
		if topic == delivery.RoutingKey {
			// TODO: Maybe we want to send the deliveries into a general queue
			// https://medium.com/justforfunc/two-ways-of-merging-n-channels-in-go-43c0b57cd1de
			go e.handleInvocation(topic, delivery)
		} else {
			log.Printf("Received message for topic %s that did not match subscribed topic %s will reject it", delivery.RoutingKey, topic)

			for retry := 0; retry < MAX_ATTEMPTS; retry++ {
				err := delivery.Reject(true)
				if err == nil {
					return
				}

				log.Printf("Failed to reject delivery %d due to %s. Attemp %d/3", delivery.DeliveryTag, err, retry+1)
				time.Sleep(time.Duration(retry+1*250) * time.Millisecond)
			}

			log.Printf("Failed to reject delivery %d, will abort reject now", delivery.DeliveryTag)
		}
	}
}

func (e *Exchange) handleInvocation(topic string, delivery amqp.Delivery) {
	// Call Function via Client
	err := e.client.Invoke(topic, types.NewInvocation(delivery))
	if err == nil {
		for retry := 0; retry < MAX_ATTEMPTS; retry++ {
			ackErr := delivery.Ack(false)
			if ackErr == nil {
				return
			}

			log.Printf("Failed to acknowledge delivery %d due to %s. Attemp %d/3", delivery.DeliveryTag, ackErr, retry+1)
			time.Sleep(time.Duration(retry+1*250) * time.Millisecond)
		}

		log.Printf("Failed to acknowledge delivery %d, will abort ack now", delivery.DeliveryTag)
	} else {
		for retry := 0; retry < MAX_ATTEMPTS; retry++ {
			nackErr := delivery.Nack(false, true)
			if nackErr == nil {
				return
			}

			log.Printf("Failed to nack delivery %d due to %s. Attemp %d/3", delivery.DeliveryTag, nackErr, retry+1)
			time.Sleep(time.Duration(retry+1*250) * time.Millisecond)
		}

		log.Printf("Failed to nack delivery %d, will abort nack now", delivery.DeliveryTag)
	}

}
