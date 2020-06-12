/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"errors"
	"fmt"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
	"log"
)

type Factory interface {
	WithConnection(maker ChannelMaker) Factory
	WithExchange(ex types.Exchange) Factory
	Build() (ExchangeOrganizer, error)
}

func NewFactory() Factory {
	return &ExchangeFactory{}
}

type ExchangeFactory struct {
	maker    ChannelMaker
	exchange types.Exchange
}

func (f *ExchangeFactory) WithConnection(maker ChannelMaker) Factory {
	f.maker = maker
	return f
}

func (f *ExchangeFactory) WithExchange(ex types.Exchange) Factory {
	log.Printf("Factory is configured for exchange %s", ex.Name)
	ex.EnsureCorrectType()
	f.exchange = ex
	return f
}

func (f *ExchangeFactory) Build() (ExchangeOrganizer, error) {
	if f.maker == nil {
		return nil, errors.New("no channel maker was provided")
	}

	channel, err := f.maker.CreateChannel()
	if err != nil {
		return nil, err
	}

	topologyErr := declareTopology(channel, f.exchange)
	if topologyErr != nil {
		return nil, topologyErr
	}

	return NewExchange(channel, f.maker, f.exchange), nil
}

func declareTopology(con *amqp.Channel, ex types.Exchange) error {
	if ex.Declare {
		err := con.ExchangeDeclare(ex.Name, ex.Type, ex.Durable, ex.AutoDeleted, false, false, nil)
		if err != nil {
			return err
		}
		log.Printf("Successfully declared exchange %s of type %s { Durable: %t Auto-Delete: %t }", ex.Name, ex.Type, ex.Durable, ex.AutoDeleted)
	}

	for _, topic := range ex.Topics {
		name := GenerateQueueName(ex.Name, topic)

		_, declareErr := con.QueueDeclare(
			name,
			ex.Durable,
			ex.AutoDeleted,
			false,
			false,
			nil,
		)
		if declareErr != nil {
			return declareErr
		}
		log.Printf("Successfully declared Queue %s", name)

		bindErr := con.QueueBind(
			name,
			topic,
			ex.Name,
			false,
			nil,
		)

		if bindErr != nil {
			return bindErr
		}
		log.Printf("Successfully bound Queue %s to exchange %s", name, ex.Name)
	}

	return nil
}

func GenerateQueueName(ex string, topic string) string {
	const PreFix = "OpenFaaS"
	return fmt.Sprintf("%s_%s_%s", PreFix, ex, topic)
}
