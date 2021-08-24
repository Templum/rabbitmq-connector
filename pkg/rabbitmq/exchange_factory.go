/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"errors"
	"fmt"
	"log"

	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
)

// Factory for building a Exchange
type Factory interface {
	WithInvoker(client types.Invoker) Factory
	WithChanCreator(creator ChannelCreator) Factory
	WithExchange(ex *types.Exchange) Factory
	Build() (ExchangeOrganizer, error)
}

// NewFactory creates a new instance with no defaults set.
func NewFactory() Factory {
	return &ExchangeFactory{}
}

// ExchangeFactory keeps tracks of all the build options provided to it during construction
type ExchangeFactory struct {
	creator  ChannelCreator
	client   types.Invoker
	exchange *types.Exchange
}

// WithChanCreator sets the channel creator that will be used
func (f *ExchangeFactory) WithChanCreator(creator ChannelCreator) Factory {
	f.creator = creator
	return f
}

// WithInvoker sets the invoker which will interact with OpenFaaS
func (f *ExchangeFactory) WithInvoker(client types.Invoker) Factory {
	f.client = client
	return f
}

// WithExchange sets the exchange definition and further ensures that the correct type is used
func (f *ExchangeFactory) WithExchange(ex *types.Exchange) Factory {
	log.Printf("Factory is configured for exchange %s", ex.Name)
	ex.EnsureCorrectType()
	f.exchange = ex
	return f
}

// Build uses the set values and builds a new exchange from them
func (f *ExchangeFactory) Build() (ExchangeOrganizer, error) {
	if f.creator == nil {
		return nil, errors.New("no channel creator was provided")
	}
	if f.client == nil {
		return nil, errors.New("no openfaas client was provided")
	}
	if f.exchange == nil {
		return nil, errors.New("no exchange configured")
	}

	channel, err := f.creator.Channel()
	if err != nil {
		return nil, err
	}

	topologyErr := declareTopology(channel, f.exchange)
	if topologyErr != nil {
		return nil, topologyErr
	}

	return NewExchange(channel, f.client, f.exchange), nil
}

func declareTopology(con RabbitChannel, ex *types.Exchange) error {
	if ex.Declare {
		err := con.ExchangeDeclare(ex.Name, ex.Type, ex.Durable, ex.AutoDeleted, false, false, amqp.Table{})
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
			amqp.Table{},
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
			amqp.Table{},
		)

		if bindErr != nil {
			return bindErr
		}
		log.Printf("Successfully bound Queue %s to exchange %s", name, ex.Name)
	}

	return nil
}

// GenerateQueueName is responsible to generate a unique queue for the connector to use
// It follows the naming schema OpenFaaS_[EXCHANGE_NAME]_[TOPIC]
func GenerateQueueName(ex string, topic string) string {
	const PreFix = "OpenFaaS"
	return fmt.Sprintf("%s_%s_%s", PreFix, ex, topic)
}
