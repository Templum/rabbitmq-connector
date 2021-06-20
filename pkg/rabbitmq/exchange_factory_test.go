/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"errors"
	"testing"

	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type creatorMock struct {
	mock.Mock
}

func (c *creatorMock) Channel() (RabbitChannel, error) {
	args := c.Called(nil)
	return args.Get(0).(RabbitChannel), args.Error(1)
}

type channelMock struct {
	mock.Mock
}

func (ch *channelMock) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	params := ch.Called(name, durable, autoDelete, exclusive, noWait, args)
	return params.Get(0).(amqp.Queue), params.Error(1)
}

func (ch *channelMock) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	params := ch.Called(name, key, exchange, noWait, args)
	return params.Error(0)
}

func (ch *channelMock) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	params := ch.Called(name, kind, durable, autoDelete, internal, noWait, args)
	return params.Error(0)
}

func (ch *channelMock) Consume(queue string, consumer string, autoAck, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	params := ch.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return params.Get(0).(<-chan amqp.Delivery), params.Error(1)
}

func (ch *channelMock) NotifyClose(c chan *amqp.Error) chan *amqp.Error {
	args := ch.Called(c)
	return args.Get(0).(chan *amqp.Error)
}

func (ch *channelMock) Close() error {
	args := ch.Called(nil)
	return args.Error(0)
}

func TestGenerateQueueName(t *testing.T) {
	const expected = "OpenFaaS_Dax_Wirecard"
	actual := GenerateQueueName("Dax", "Wirecard")

	assert.EqualValues(t, expected, actual)
}

func TestExchangeFactory_Build(t *testing.T) {
	exchange := &types.Exchange{
		Name:        "Dax",
		Topics:      []string{"Wirecard", "BMW"},
		Declare:     true,
		Type:        "direct",
		Durable:     true,
		AutoDeleted: true,
	}

	t.Run("Should successfully build exchanges", func(t *testing.T) {
		invoker := new(invokerMock)
		channel := new(channelMock)
		channel.On("ExchangeDeclare", "Dax", "direct", true, true, false, false, amqp.Table{}).Return(nil)
		channel.On("QueueDeclare", "OpenFaaS_Dax_Wirecard", true, true, false, false, amqp.Table{}).Return(amqp.Queue{}, nil)
		channel.On("QueueDeclare", "OpenFaaS_Dax_BMW", true, true, false, false, amqp.Table{}).Return(amqp.Queue{}, nil)
		channel.On("QueueBind", "OpenFaaS_Dax_Wirecard", "Wirecard", "Dax", false, amqp.Table{}).Return(nil)
		channel.On("QueueBind", "OpenFaaS_Dax_BMW", "BMW", "Dax", false, amqp.Table{}).Return(nil)

		creator := new(creatorMock)
		creator.On("Channel", nil).Return(channel, nil)

		target := NewFactory()
		target.WithChanCreator(creator)
		target.WithInvoker(invoker)
		target.WithExchange(exchange)

		organizer, err := target.Build()

		assert.NoError(t, err, "should not throw")
		assert.NotNil(t, organizer, "should not be nil")

		creator.AssertExpectations(t)
		channel.AssertExpectations(t)
	})

	t.Run("Should raise error if no creator was provided", func(t *testing.T) {
		target := NewFactory()
		organizer, err := target.Build()

		assert.Nil(t, organizer, "should be nil in error case")
		assert.Error(t, err, "no channel creator was provided")
	})

	t.Run("Should raise error if no invoker was provided", func(t *testing.T) {
		creator := new(creatorMock)

		target := NewFactory()
		target.WithChanCreator(creator)
		organizer, err := target.Build()

		assert.Nil(t, organizer, "should be nil in error case")
		assert.Error(t, err, "no openfaas client was provided")
	})

	t.Run("Should raise error if no exchange was provided", func(t *testing.T) {
		creator := new(creatorMock)
		invoker := new(invokerMock)

		target := NewFactory()
		target.WithChanCreator(creator)
		target.WithInvoker(invoker)

		organizer, err := target.Build()

		assert.Nil(t, organizer, "should be nil in error case")
		assert.Error(t, err, "no exchange configured")
	})

	t.Run("Should raise error if channel creation fails", func(t *testing.T) {
		channel := new(channelMock)
		invoker := new(invokerMock)
		creator := new(creatorMock)
		creator.On("Channel", nil).Return(channel, errors.New("channel creation failed"))

		target := NewFactory()
		target.WithChanCreator(creator)
		target.WithInvoker(invoker)
		target.WithExchange(exchange)

		organizer, err := target.Build()

		assert.Nil(t, organizer, "should be nil in error case")
		assert.Error(t, err, "channel creation failed")
		creator.AssertExpectations(t)
	})

	t.Run("Should raise error if exchange declaration fails", func(t *testing.T) {
		invoker := new(invokerMock)
		channel := new(channelMock)
		channel.On("ExchangeDeclare", "Dax", "direct", true, true, false, false, amqp.Table{}).Return(errors.New("failure"))

		creator := new(creatorMock)
		creator.On("Channel", nil).Return(channel, nil)

		target := NewFactory()
		target.WithChanCreator(creator)
		target.WithInvoker(invoker)
		target.WithExchange(exchange)

		organizer, err := target.Build()

		assert.Nil(t, organizer, "should be nil in error case")
		assert.Error(t, err, "failure")
		creator.AssertExpectations(t)
		channel.AssertExpectations(t)
	})

	t.Run("Should raise error if queue declaration fails", func(t *testing.T) {
		invoker := new(invokerMock)
		channel := new(channelMock)
		channel.On("ExchangeDeclare", "Dax", "direct", true, true, false, false, amqp.Table{}).Return(nil)
		channel.On("QueueDeclare", "OpenFaaS_Dax_Wirecard", true, true, false, false, amqp.Table{}).Return(amqp.Queue{}, errors.New("failure"))

		creator := new(creatorMock)
		creator.On("Channel", nil).Return(channel, nil)

		target := NewFactory()
		target.WithChanCreator(creator)
		target.WithInvoker(invoker)
		target.WithExchange(exchange)

		organizer, err := target.Build()

		assert.Nil(t, organizer, "should be nil in error case")
		assert.Error(t, err, "failure")
		creator.AssertExpectations(t)
		channel.AssertExpectations(t)
	})

	t.Run("Should raise error if queue binding fails", func(t *testing.T) {
		invoker := new(invokerMock)
		channel := new(channelMock)
		channel.On("ExchangeDeclare", "Dax", "direct", true, true, false, false, amqp.Table{}).Return(nil)
		channel.On("QueueDeclare", "OpenFaaS_Dax_Wirecard", true, true, false, false, amqp.Table{}).Return(amqp.Queue{}, nil)
		channel.On("QueueBind", "OpenFaaS_Dax_Wirecard", "Wirecard", "Dax", false, amqp.Table{}).Return(errors.New("failure"))

		creator := new(creatorMock)
		creator.On("Channel", nil).Return(channel, nil)

		target := NewFactory()
		target.WithChanCreator(creator)
		target.WithInvoker(invoker)
		target.WithExchange(exchange)

		organizer, err := target.Build()

		assert.Nil(t, organizer, "should be nil in error case")
		assert.Error(t, err, "failure")
		creator.AssertExpectations(t)
		channel.AssertExpectations(t)
	})
}
