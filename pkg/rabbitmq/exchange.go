/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
	"log"
	"sync"
)

type Starter interface {
	Start() error
}

type Stopper interface {
	Stop()
}

type ExchangeOrganizer interface {
	IsConnected() bool
	IsRunning() bool

	Starter
	Stopper
}

type Exchange struct {
	channel *amqp.Channel
	maker   ChannelMaker

	definition types.Exchange
	consumers  []interface{} // TODO: Create Type/Interface for this jizz
	lock       sync.RWMutex

	isRunning   bool
	isConnected bool
}

func NewExchange(channel *amqp.Channel, maker ChannelMaker, definition types.Exchange) ExchangeOrganizer {

	ref := Exchange{
		channel: channel,
		maker:   maker,

		definition: definition,
		lock:       sync.RWMutex{},

		isConnected: true,
		isRunning:   false,
	}

	// This will also take care of recovering in case of issues
	go ref.observeChannelState()
	return &ref
}

func (e *Exchange) Start() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, topic := range e.definition.Topics {
		deliveries, err := e.channel.Consume(GenerateQueueName(e.definition.Name, topic), "", true, false, false, false, nil)
		if err != nil {
			return err
		}

		go func(deliveries <-chan amqp.Delivery) {
			// TODO: Handle Deliveries
		}(deliveries)
	}

	e.isRunning = true
	return nil
}

func (e *Exchange) Stop() {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.isRunning = false

}

func (e *Exchange) IsConnected() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.isConnected
}

func (e *Exchange) IsRunning() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.isRunning
}

func (e *Exchange) observeChannelState() {
	closeChannel := make(chan *amqp.Error)
	e.channel.NotifyClose(closeChannel)

	for {
		received := <-closeChannel

		if received.Recover {
			log.Printf("Received non critical error %s.", received)
		} else {
			log.Printf("Received critical error %s.", received)
		}
		e.lock.Lock()
		e.isConnected = false

		newChannel, err := e.maker.CreateChannel()

		if err != nil {
			e.channel.Close() // TODO: Maybe not needed as already closed
			close(closeChannel)
			e.lock.Unlock()

			log.Printf("Was not able to recover channel for exchange %s due to %s", e.definition.Name, err)
			return
		}

		e.channel = newChannel
		e.isConnected = true
		e.lock.Unlock()

		e.Stop()
		e.Start() // Ignoring Error here
	}

}
