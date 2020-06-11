/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package subscriber

import (
	"errors"
	"runtime"
	"testing"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//---- QueueConsumer Mock ----//

type minimalQueueConsumer struct {
	IsActive bool
	mock.Mock
}

func (m *minimalQueueConsumer) Consume() (<-chan *types.OpenFaaSInvocation, error) {
	m.IsActive = true

	m.Called()
	return make(<-chan *types.OpenFaaSInvocation), nil
}

func (m *minimalQueueConsumer) Stop() {
	m.IsActive = false

	m.Called()
}

func (m *minimalQueueConsumer) ListenForErrors() <-chan *amqp.Error {
	return make(chan *amqp.Error)
}

//---- Factory Mock ----//

type mockFactory struct {
	mock.Mock
}

func (m *mockFactory) Build(topic string) (rabbitmq.QueueConsumer, error) {
	args := m.Called(topic)

	if c, ok := args.Get(0).(rabbitmq.QueueConsumer); ok {
		return c, nil
	}
	return nil, args.Get(1).(error)
}

func TestConnector_Start(t *testing.T) {
	cfg := config.Controller{Topics: []string{"Hello"}}

	t.Parallel()

	t.Run("Start with no errors", func(t *testing.T) {
		consumerMock := &minimalQueueConsumer{IsActive: false}
		consumerMock.On("Consume").Return()
		factory := new(mockFactory)
		factory.On("Build", "Hello").Return(consumerMock, nil)

		target := NewConnector(&cfg, nil, factory)
		target.Start()

		connector, _ := target.(*connector)
		assert.Equalf(t, len(connector.subscribers), CalculateWorkerCount(1), "Expected Connector to start %d Worker instead he started %d", CalculateWorkerCount(1), len(connector.subscribers))

		for _, sub := range connector.subscribers {
			subscriber, _ := sub.(*subscriber)
			assert.True(t, subscriber.IsRunning(), "Subscriber was not running")
		}
	})

	t.Run("Start with errors", func(t *testing.T) {
		factory := new(mockFactory)
		factory.On("Build", "Hello").Return(nil, errors.New("expected error"))
		target := NewConnector(&cfg, nil, factory)
		target.Start()

		connector, _ := target.(*connector)
		assert.Lenf(t, connector.subscribers, 0, "Expected Connector to start 0 Worker instead he started %d", len(connector.subscribers))
	})
}

func TestConnector_End(t *testing.T) {
	cfg := config.Controller{Topics: []string{"Hello"}}

	consumerMock := &minimalQueueConsumer{IsActive: false}
	consumerMock.On("Consume").Return()
	consumerMock.On("Stop").Return(nil)
	factory := new(mockFactory)
	factory.On("Build", "Hello").Return(consumerMock, nil)
	target := NewConnector(&cfg, nil, factory)
	target.Start()
	target.End()

	connector, _ := target.(*connector)
	assert.Lenf(t, connector.subscribers, 0, "Expected Connector to cleanup workers. %d are still left", len(connector.subscribers))
}

func TestCalculateWorkerCount(t *testing.T) {
	t.Parallel()

	t.Run("All Workers on one topic", func(t *testing.T) {
		target := runtime.NumCPU() * 2

		calculated := CalculateWorkerCount(1)
		assert.Equal(t, target, calculated, "Expected one topic per worker")
	})

	t.Run("Should Split between two topics", func(t *testing.T) {
		target := runtime.NumCPU()

		calculated := CalculateWorkerCount(2)
		assert.Equal(t, target, calculated, "Expected two topics per worker")
	})

	t.Run("Exactly one worker per topic", func(t *testing.T) {
		target := 1

		calculated := CalculateWorkerCount(runtime.NumCPU() * 2)
		assert.Equal(t, target, calculated, "Expected exactly one topic per worker")
	})

	t.Run("At least one worker", func(t *testing.T) {
		target := 1

		calculated := CalculateWorkerCount(runtime.NumCPU()*2 + 2)
		assert.Equal(t, target, calculated, "Expected at least one topic per worker")
	})
}
