/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"crypto/tls"
	"errors"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type brokerMocker struct {
	mock.Mock
}

func (b *brokerMocker) Dial(url string) (*amqp.Connection, error) {
	args := b.Called(url)
	return args.Get(0).(*amqp.Connection), args.Error(1)
}

func (b *brokerMocker) DialTLS(url string, conf *tls.Config) (*amqp.Connection, error) {
	args := b.Called(url, conf)
	return args.Get(0).(*amqp.Connection), args.Error(1)
}

func TestConnectionManager_Connect(t *testing.T) {
	t.Run("Should connect to specified Rabbit MQ host and return close channel", func(t *testing.T) {
		broker := new(brokerMocker)
		broker.On("Dial", "amqp://user:pass@localhost:5672/").Return(&amqp.Connection{}, nil)

		target := NewConnectionManager(broker)

		ch, err := target.Connect("amqp://user:pass@localhost:5672/")

		assert.NoError(t, err, "should not throw")
		assert.NotNil(t, ch, "should not be nil")
		broker.AssertExpectations(t)
	})

	t.Run("Should try at least 3 times to connect to Rabbit MQ before throwing error", func(t *testing.T) {
		broker := new(brokerMocker)
		broker.On("Dial", "amqp://user:pass@localhost:5672/").Return(&amqp.Connection{}, errors.New("expected"))

		target := NewConnectionManager(broker)

		ch, err := target.Connect("amqp://user:pass@localhost:5672/")

		assert.Error(t, err, "could not establish connection to Rabbit MQ Cluster")
		assert.Nil(t, ch, "should be nil")
		broker.AssertExpectations(t)
		broker.AssertNumberOfCalls(t, "Dial", 3)
	})
}
