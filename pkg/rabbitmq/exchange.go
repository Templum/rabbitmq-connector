/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"log"
	"sync"

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
	channel Channel
	client  types.Invoker

	definition *types.Exchange
	lock       sync.RWMutex

	isRunning   bool
	isConnected bool
}

func NewExchange(channel Channel, client types.Invoker, definition *types.Exchange) ExchangeOrganizer {
	return &Exchange{
		channel: channel,
		client:  client,

		definition: definition,
		lock:       sync.RWMutex{},

		isConnected: true,
		isRunning:   false,
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
		deliveries, err := e.channel.Consume(queueName, "", true, false, false, false, nil)
		if err != nil {
			return err
		}

		go e.startConsuming(topic, deliveries)
	}

	e.isRunning = true
	return nil
}

func (e *Exchange) Stop() {
	e.lock.Lock()
	defer e.lock.Unlock()

	// We ignore the issue since this method is usually called after connection failure.
	_ = e.channel.Close()
	e.isRunning = false
}

func (e *Exchange) handleChanFailure(ch <-chan *amqp.Error) {
	err := <-ch
	log.Printf("Received following error %s on channel for exchange %s", err, e.definition.Name)
}

func (e *Exchange) startConsuming(topic string, deliveries <-chan amqp.Delivery) {
	for delivery := range deliveries {
		if topic == delivery.RoutingKey {
			// TODO: Maybe we want to send the deliveries into a general queue
			// https://medium.com/justforfunc/two-ways-of-merging-n-channels-in-go-43c0b57cd1de
			go e.client.Invoke(topic, types.NewInvocation(delivery))
		} else {
			// TODO: Debug Log
			log.Printf("Received message for topic %s that did not match subsribed topic %s", delivery.RoutingKey, topic)
		}
	}
}
