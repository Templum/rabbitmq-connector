/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"errors"
	"testing"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type invokerMock struct {
	mock.Mock
}

func (i *invokerMock) Invoke(topic string, invocation *types.OpenFaaSInvocation) error {
	i.Called(topic, invocation)
	return nil
}

func TestExchange_Start(t *testing.T) {
	definition := types.Exchange{
		Name:   "Nasdaq",
		Topics: []string{"Billing", "Transport"},
	}

	t.Run("Should successfully start consuming for defined topics", func(t *testing.T) {
		channel := new(channelMock)
		channel.On("Consume", "OpenFaaS_Nasdaq_Billing", "", true, false, false, false, amqp.Table{}).Return(make(<-chan amqp.Delivery), nil)
		channel.On("Consume", "OpenFaaS_Nasdaq_Transport", "", true, false, false, false, amqp.Table{}).Return(make(<-chan amqp.Delivery), nil)
		channel.On("NotifyClose", mock.Anything).Return(make(chan *amqp.Error))

		invoker := new(invokerMock)

		target := NewExchange(channel, invoker, &definition)

		err := target.Start()
		assert.NoError(t, err, "should not throw")

		invoker.AssertExpectations(t)
		channel.AssertExpectations(t)
	})

	t.Run("Should return occurred error when starting consume failed", func(t *testing.T) {
		channel := new(channelMock)
		channel.On("Consume", "OpenFaaS_Nasdaq_Billing", "", true, false, false, false, amqp.Table{}).Return(make(<-chan amqp.Delivery), errors.New("expected"))
		channel.On("NotifyClose", mock.Anything).Return(make(chan *amqp.Error))

		invoker := new(invokerMock)

		target := NewExchange(channel, invoker, &definition)

		err := target.Start()
		assert.Error(t, err, "expected")

		invoker.AssertExpectations(t)
		channel.AssertExpectations(t)
	})
}

func createDeliveries(message amqp.Delivery) <-chan amqp.Delivery {
	deliveries := make(chan amqp.Delivery, 1)
	deliveries <- message

	go func(ch chan amqp.Delivery) {
		time.Sleep(10 * time.Millisecond)
		close(ch)
	}(deliveries)

	return deliveries
}

func TestExchange_StartConsuming(t *testing.T) {
	definition := types.Exchange{
		Name:   "Nasdaq",
		Topics: []string{"Billing"},
	}

	t.Run("Should invoke function when message with matching routingkey was received", func(t *testing.T) {
		invoker := new(invokerMock)
		invoker.On("Invoke", "Billing", mock.Anything)

		target := Exchange{
			client:     invoker,
			definition: &definition,
		}

		target.StartConsuming("Billing", createDeliveries(amqp.Delivery{
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			RoutingKey:      "Billing",
			Body:            []byte("Hello World"),
		}))

		invoker.AssertExpectations(t)
	})

	t.Run("Should not invoke when received message is of no registered topic", func(t *testing.T) {
		invoker := new(invokerMock)
		invoker.On("Invoke", "Billing", mock.Anything)

		target := Exchange{
			client:     invoker,
			definition: &definition,
		}

		target.StartConsuming("Account", createDeliveries(amqp.Delivery{
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			RoutingKey:      "Billing",
			Body:            []byte("Hello World"),
		}))

		invoker.AssertNotCalled(t, "Invoke", mock.Anything)
	})
}

func TestExchange_Stop(t *testing.T) {
	t.Run("Should stop channel", func(t *testing.T) {
		channel := new(channelMock)
		channel.On("Close", nil).Return(nil)

		target := Exchange{
			channel: channel,
		}

		target.Stop()
		channel.AssertExpectations(t)
	})

	t.Run("Should ignore issues that occur during channel closing", func(t *testing.T) {
		channel := new(channelMock)
		channel.On("Close", nil).Return(errors.New("ignore me"))

		target := Exchange{
			channel: channel,
		}

		target.Stop()
		channel.AssertExpectations(t)
	})
}
