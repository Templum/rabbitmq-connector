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

// Starter defines something that can be started
type Starter interface {
	Start() error
}

// Stopper defines something that can be stopped
type Stopper interface {
	Stop()
}

// ExchangeOrganizer combines the ability to start & stop exchanges
type ExchangeOrganizer interface {
	Starter
	Stopper
}

// Exchange contains all of the relevant units to handle communication with an exchange
type Exchange struct {
	channel ChannelConsumer
	client  types.Invoker

	definition *types.Exchange
	lock       sync.RWMutex
}

// MaxAttempts of retries that will be performed
const MaxAttempts = 3

// NewExchange creates a new exchange instance using the provided parameter
func NewExchange(channel ChannelConsumer, client types.Invoker, definition *types.Exchange) ExchangeOrganizer {
	return &Exchange{
		channel: channel,
		client:  client,

		definition: definition,
		lock:       sync.RWMutex{},
	}
}

// Start s consuming deliveries from a unique queue for the specific exchange.
// Further creating a listener for channel errors
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

// Stop s consuming messages
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

// StartConsuming will consume deliveries from the provided channel and if the received delivery
// is for the target topic it will invoke it. If the delivery is not for the correct topic it will
// reject it so that the delivery is returned to the exchange. Retries are exponential and up to 3 times.
func (e *Exchange) StartConsuming(topic string, deliveries <-chan amqp.Delivery) {
	for delivery := range deliveries {
		if topic == delivery.RoutingKey {
			// TODO: Maybe we want to send the deliveries into a general queue
			// https://medium.com/justforfunc/two-ways-of-merging-n-channels-in-go-43c0b57cd1de
			go e.handleInvocation(topic, delivery)
		} else {
			log.Printf("Received message for topic %s that did not match subscribed topic %s will reject it", delivery.RoutingKey, topic)

			for retry := 0; retry < MaxAttempts; retry++ {
				err := delivery.Reject(true)
				if err == nil {
					return
				}

				log.Printf("Failed to reject delivery %d due to %s. Attempt %d/3", delivery.DeliveryTag, err, retry+1)
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
		for retry := 0; retry < MaxAttempts; retry++ {
			ackErr := delivery.Ack(false)
			if ackErr == nil {
				return
			}

			log.Printf("Failed to acknowledge delivery %d due to %s. Attempt %d/3", delivery.DeliveryTag, ackErr, retry+1)
			time.Sleep(time.Duration(retry+1*250) * time.Millisecond)
		}

		log.Printf("Failed to acknowledge delivery %d, will abort ack now", delivery.DeliveryTag)
	} else {
		for retry := 0; retry < MaxAttempts; retry++ {
			nackErr := delivery.Nack(false, true)
			if nackErr == nil {
				return
			}

			log.Printf("Failed to nack delivery %d due to %s. Attempt %d/3", delivery.DeliveryTag, nackErr, retry+1)
			time.Sleep(time.Duration(retry+1*250) * time.Millisecond)
		}

		log.Printf("Failed to nack delivery %d, will abort nack now", delivery.DeliveryTag)
	}

}
