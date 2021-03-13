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

type acknowledgerMock struct {
	mock.Mock
}

func (a *acknowledgerMock) Ack(tag uint64, multiple bool) error {
	args := a.Called(tag, multiple)
	return args.Error(0)
}

func (a *acknowledgerMock) Nack(tag uint64, multiple bool, requeue bool) error {
	args := a.Called(tag, multiple, requeue)
	return args.Error(0)
}

func (a *acknowledgerMock) Reject(tag uint64, requeue bool) error {
	args := a.Called(tag, requeue)
	return args.Error(0)
}

type invokerMock struct {
	mock.Mock
}

func (i *invokerMock) Invoke(topic string, invocation *types.OpenFaaSInvocation) error {
	args := i.Called(topic, invocation)
	return args.Error(0)
}

func TestExchange_Start(t *testing.T) {
	definition := types.Exchange{
		Name:   "Nasdaq",
		Topics: []string{"Billing", "Transport"},
	}

	t.Run("Should successfully start consuming for defined topics", func(t *testing.T) {
		channel := new(channelMock)
		channel.On("Consume", "OpenFaaS_Nasdaq_Billing", "", false, false, false, false, amqp.Table{}).Return(make(<-chan amqp.Delivery), nil)
		channel.On("Consume", "OpenFaaS_Nasdaq_Transport", "", false, false, false, false, amqp.Table{}).Return(make(<-chan amqp.Delivery), nil)
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
		channel.On("Consume", "OpenFaaS_Nasdaq_Billing", "", false, false, false, false, amqp.Table{}).Return(make(<-chan amqp.Delivery), errors.New("expected"))
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

	t.Run("Should invoke function when message is for registered routing key and further ack processing of message if no error occurred", func(t *testing.T) {
		invoker := new(invokerMock)
		invoker.On("Invoke", "Billing", mock.Anything).Return(nil)

		acker := new(acknowledgerMock)
		acker.On("Ack", mock.Anything, false).Return(nil)

		target := Exchange{
			client:     invoker,
			definition: &definition,
		}

		target.StartConsuming("Billing", createDeliveries(amqp.Delivery{
			Acknowledger:    acker,
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			RoutingKey:      "Billing",
			Body:            []byte("Hello World"),
		}))

		invoker.AssertExpectations(t)
		acker.AssertExpectations(t)
	})

	t.Run("Should attempt to ack successful invocations up to 3 times", func(t *testing.T) {
		invoker := new(invokerMock)
		invoker.On("Invoke", "Billing", mock.Anything).Return(nil)

		acker := new(acknowledgerMock)
		acker.On("Ack", mock.Anything, false).Return(errors.New("failed"))

		target := Exchange{
			client:     invoker,
			definition: &definition,
		}

		target.StartConsuming("Billing", createDeliveries(amqp.Delivery{
			Acknowledger:    acker,
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			RoutingKey:      "Billing",
			Body:            []byte("Hello World"),
		}))

		time.Sleep(2 * time.Second)

		acker.AssertExpectations(t)
		acker.AssertNumberOfCalls(t, "Ack", 3)
	})

	t.Run("Should invoke function when message is for registered routing key and further send back to queue when error occurred", func(t *testing.T) {
		invoker := new(invokerMock)
		invoker.On("Invoke", "Billing", mock.Anything).Return(errors.New("failed to invoke"))

		acker := new(acknowledgerMock)
		acker.On("Nack", mock.Anything, false, true).Return(nil)

		target := Exchange{
			client:     invoker,
			definition: &definition,
		}

		target.StartConsuming("Billing", createDeliveries(amqp.Delivery{
			Acknowledger:    acker,
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			RoutingKey:      "Billing",
			Body:            []byte("Hello World"),
		}))

		invoker.AssertExpectations(t)
		acker.AssertExpectations(t)
	})

	t.Run("Should attempt to nack unsuccessful invocations up to 3 times", func(t *testing.T) {
		invoker := new(invokerMock)
		invoker.On("Invoke", "Billing", mock.Anything).Return(errors.New("failed to invoke"))

		acker := new(acknowledgerMock)
		acker.On("Nack", mock.Anything, false, true).Return(errors.New("failed"))

		target := Exchange{
			client:     invoker,
			definition: &definition,
		}

		target.StartConsuming("Billing", createDeliveries(amqp.Delivery{
			Acknowledger:    acker,
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			RoutingKey:      "Billing",
			Body:            []byte("Hello World"),
		}))

		time.Sleep(2 * time.Second)

		acker.AssertExpectations(t)
		acker.AssertNumberOfCalls(t, "Nack", 3)
	})

	t.Run("Should not invoke when received message is of no registered topic and further reject message and send it back to queue", func(t *testing.T) {
		invoker := new(invokerMock)
		invoker.On("Invoke", "Billing", mock.Anything).Return(nil)

		acker := new(acknowledgerMock)
		acker.On("Reject", mock.Anything, true).Return(nil)

		target := Exchange{
			client:     invoker,
			definition: &definition,
		}

		target.StartConsuming("Account", createDeliveries(amqp.Delivery{
			Acknowledger:    acker,
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			RoutingKey:      "Billing",
			Body:            []byte("Hello World"),
		}))

		invoker.AssertNotCalled(t, "Invoke", mock.Anything)
		acker.AssertExpectations(t)
	})

	t.Run("Should attempt to reject deliveries for unregistered topics up to 3 times", func(t *testing.T) {
		invoker := new(invokerMock)
		invoker.On("Invoke", "Billing", mock.Anything).Return(nil)

		acker := new(acknowledgerMock)
		acker.On("Reject", mock.Anything, true).Return(errors.New("failed"))

		target := Exchange{
			client:     invoker,
			definition: &definition,
		}

		target.StartConsuming("Account", createDeliveries(amqp.Delivery{
			Acknowledger:    acker,
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			RoutingKey:      "Billing",
			Body:            []byte("Hello World"),
		}))

		acker.AssertExpectations(t)
		acker.AssertNumberOfCalls(t, "Reject", 3)
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
