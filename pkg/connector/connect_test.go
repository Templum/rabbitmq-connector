/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package connector

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type managerMock struct {
	mock.Mock
}

func (m *managerMock) Connect(connectionUrl string) (<-chan *amqp.Error, error) {
	args := m.Called(connectionUrl)
	return args.Get(0).(<-chan *amqp.Error), args.Error(1)
}

func (m *managerMock) Reconnect(connectionUrl string) (<-chan *amqp.Error, error) {
	args := m.Called(connectionUrl)
	return args.Get(0).(<-chan *amqp.Error), args.Error(1)
}

func (m *managerMock) Disconnect() {
	_ = m.Called(nil)
}

func (m *managerMock) Channel() (*amqp.Channel, error) {
	args := m.Called(nil)
	return args.Get(0).(*amqp.Channel), args.Error(1)
}

type factoryMock struct {
	mock.Mock
}

func (f *factoryMock) WithInvoker(client types.Invoker) rabbitmq.Factory {
	f.Called(nil)
	return f
}

func (f *factoryMock) WithChanCreator(creator rabbitmq.RBChannelCreator) rabbitmq.Factory {
	f.Called(nil)
	return f
}

func (f *factoryMock) WithExchange(ex *types.Exchange) rabbitmq.Factory {
	f.Called(nil)
	return f
}

func (f *factoryMock) Build() (rabbitmq.ExchangeOrganizer, error) {
	args := f.Called(nil)
	tmp := args.Get(0)

	if tmp == nil {
		return nil, args.Error(1)
	} else {
		return tmp.(rabbitmq.ExchangeOrganizer), args.Error(1)
	}
}

type exchangeMock struct {
	mock.Mock
}

func (e *exchangeMock) Start() error {
	arg := e.Called(nil)
	return arg.Error(0)
}

func (e *exchangeMock) Stop() {
	e.Called(nil)
}

func TestConnector_Run(t *testing.T) {
	conf := config.Controller{
		RabbitSanitizedURL:  "amqp://localhost:5672/",
		RabbitConnectionURL: "amqp://user:pass@localhost:5672/",
		Topology: types.Topology{struct {
			Name        string   "json:\"name\""
			Topics      []string "json:\"topics\""
			Declare     bool     "json:\"declare\""
			Type        string   "json:\"type,omitempty\""
			Durable     bool     "json:\"durable,omitempty\""
			AutoDeleted bool     "json:\"auto-deleted,omitempty\""
		}{
			Name:        "Nasdaq",
			Topics:      []string{"Transport", "Billing"},
			Declare:     false,
			Type:        "direct",
			Durable:     false,
			AutoDeleted: false,
		}},
	}

	t.Run("Should start every exchange after building topology", func(t *testing.T) {
		manager := new(managerMock)
		manager.On("Connect", conf.RabbitConnectionURL).Return(make(<-chan *amqp.Error), nil)

		exchange := new(exchangeMock)
		exchange.On("Start", nil).Return(nil)

		factory := new(factoryMock)
		factory.On("WithInvoker", nil)
		factory.On("WithChanCreator", nil)
		factory.On("WithExchange", nil)
		factory.On("Build", nil).Return(exchange, nil)

		target := New(manager, factory, nil, &conf)

		err := target.Run()
		assert.NoError(t, err, "should not throw")
		manager.AssertExpectations(t)
		factory.AssertExpectations(t)
		exchange.AssertExpectations(t)
	})

	t.Run("Should return error encountered during connecting to RabbitMQ", func(t *testing.T) {
		manager := new(managerMock)
		manager.On("Connect", conf.RabbitConnectionURL).Return(make(<-chan *amqp.Error), errors.New("could not establish connection to Rabbit MQ Cluster"))
		factory := new(factoryMock)

		target := New(manager, factory, nil, &conf)

		err := target.Run()
		assert.Error(t, err, "could not establish connection to Rabbit MQ Cluster")
		manager.AssertExpectations(t)
	})

	t.Run("Should return error encountered during topology building", func(t *testing.T) {
		manager := new(managerMock)
		manager.On("Connect", conf.RabbitConnectionURL).Return(make(<-chan *amqp.Error), nil)

		factory := new(factoryMock)
		factory.On("WithInvoker", nil)
		factory.On("WithChanCreator", nil)
		factory.On("WithExchange", nil)
		factory.On("Build", nil).Return(nil, errors.New("build error"))

		target := New(manager, factory, nil, &conf)

		err := target.Run()
		assert.Error(t, err, "build error")
		manager.AssertExpectations(t)
		factory.AssertExpectations(t)
	})

	t.Run("Should return error encountered during starting exchange", func(t *testing.T) {
		manager := new(managerMock)
		manager.On("Connect", conf.RabbitConnectionURL).Return(make(<-chan *amqp.Error), nil)

		exchange := new(exchangeMock)
		exchange.On("Start", nil).Return(errors.New("exchange start error"))

		factory := new(factoryMock)
		factory.On("WithInvoker", nil)
		factory.On("WithChanCreator", nil)
		factory.On("WithExchange", nil)
		factory.On("Build", nil).Return(exchange, nil)

		target := New(manager, factory, nil, &conf)

		err := target.Run()
		assert.Error(t, err, "exchange start error")
		manager.AssertExpectations(t)
		factory.AssertExpectations(t)
		exchange.AssertExpectations(t)
	})
}

func TestConnector_Stop(t *testing.T) {
	conf := config.Controller{
		RabbitSanitizedURL:  "amqp://localhost:5672/",
		RabbitConnectionURL: "amqp://user:pass@localhost:5672/",
		Topology: types.Topology{struct {
			Name        string   "json:\"name\""
			Topics      []string "json:\"topics\""
			Declare     bool     "json:\"declare\""
			Type        string   "json:\"type,omitempty\""
			Durable     bool     "json:\"durable,omitempty\""
			AutoDeleted bool     "json:\"auto-deleted,omitempty\""
		}{
			Name:        "Nasdaq",
			Topics:      []string{"Transport", "Billing"},
			Declare:     false,
			Type:        "direct",
			Durable:     false,
			AutoDeleted: false,
		}},
	}

	manager := new(managerMock)
	manager.On("Connect", conf.RabbitConnectionURL).Return(make(<-chan *amqp.Error), nil)
	manager.On("Disconnect", nil)

	exchange := new(exchangeMock)
	exchange.On("Start", nil).Return(nil)
	exchange.On("Stop", nil)

	factory := new(factoryMock)
	factory.On("WithInvoker", nil)
	factory.On("WithChanCreator", nil)
	factory.On("WithExchange", nil)
	factory.On("Build", nil).Return(exchange, nil)

	target := New(manager, factory, nil, &conf)

	_ = target.Run()

	t.Run("Should stop all registered exchanges during shutdown", func(t *testing.T) {
		target.Shutdown()

		manager.AssertExpectations(t)
		exchange.AssertExpectations(t)
	})
}

func makeErrorStream(wg *sync.WaitGroup, err *amqp.Error) <-chan *amqp.Error {
	errorStream := make(chan *amqp.Error, 1)

	// This is a bit hacky, but should ensure correct checking of background behavior
	go func(wg *sync.WaitGroup, err *amqp.Error, c chan *amqp.Error) {
		time.Sleep(50 * time.Millisecond)
		c <- err
		time.Sleep(50 * time.Millisecond)
		wg.Done()
	}(wg, err, errorStream)

	return errorStream
}

func TestConnector_handleConnectionError(t *testing.T) {
	conf := config.Controller{
		RabbitSanitizedURL:  "amqp://localhost:5672/",
		RabbitConnectionURL: "amqp://user:pass@localhost:5672/",
		Topology: types.Topology{struct {
			Name        string   "json:\"name\""
			Topics      []string "json:\"topics\""
			Declare     bool     "json:\"declare\""
			Type        string   "json:\"type,omitempty\""
			Durable     bool     "json:\"durable,omitempty\""
			AutoDeleted bool     "json:\"auto-deleted,omitempty\""
		}{
			Name:        "Nasdaq",
			Topics:      []string{"Transport", "Billing"},
			Declare:     false,
			Type:        "direct",
			Durable:     false,
			AutoDeleted: false,
		}},
	}

	t.Run("Should attempt recovery if observed error is recoverable", func(t *testing.T) {
		var waitgroup sync.WaitGroup
		waitgroup.Add(1)

		manager := new(managerMock)
		manager.On("Connect", conf.RabbitConnectionURL).Return(makeErrorStream(&waitgroup, &amqp.Error{
			Code:    200,
			Reason:  "Expected",
			Server:  true,
			Recover: true,
		}), nil)

		exchange := new(exchangeMock)
		exchange.On("Start", nil).Return(nil)
		exchange.On("Stop", nil)

		factory := new(factoryMock)
		factory.On("WithInvoker", nil)
		factory.On("WithChanCreator", nil)
		factory.On("WithExchange", nil)
		factory.On("Build", nil).Return(exchange, nil)

		target := New(manager, factory, nil, &conf)

		err := target.Run()
		assert.NoError(t, err, "")

		waitgroup.Wait()

		manager.AssertExpectations(t)
		factory.AssertExpectations(t)
		exchange.AssertExpectations(t)

		manager.AssertNumberOfCalls(t, "Connect", 2)
		exchange.AssertNumberOfCalls(t, "Start", 2)
		factory.AssertNumberOfCalls(t, "Build", 2)
	})
}
