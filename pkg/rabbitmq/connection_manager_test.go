/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"crypto/tls"
	"errors"
	"sync"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type conMock struct {
	mock.Mock
}

func (c *conMock) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	args := c.Called(nil)
	return args.Get(0).(chan *amqp.Error)
}

func (c *conMock) Close() error {
	args := c.Called(nil)
	return args.Error(0)
}

func (c *conMock) Channel() (*amqp.Channel, error) {
	args := c.Called(nil)
	return args.Get(0).(*amqp.Channel), args.Error(1)
}

type brokerMocker struct {
	mock.Mock
}

func (b *brokerMocker) Dial(url string) (RBConnection, error) {
	args := b.Called(url)
	return args.Get(0).(RBConnection), args.Error(1)
}

func (b *brokerMocker) DialTLS(url string, conf *tls.Config) (RBConnection, error) {
	args := b.Called(url, conf)
	return args.Get(0).(RBConnection), args.Error(1)
}

func TestConnectionManager_Connect(t *testing.T) {
	t.Run("Should connect to specified Rabbit MQ host and return close channel", func(t *testing.T) {
		con := new(conMock)
		con.On("NotifyClose", nil).Return(make(chan *amqp.Error))

		broker := new(brokerMocker)
		broker.On("Dial", "amqp://user:pass@localhost:5672/").Return(con, nil)

		target := NewConnectionManager(broker, nil)

		ch, err := target.Connect("amqp://user:pass@localhost:5672/")

		assert.NoError(t, err, "should not throw")
		assert.NotNil(t, ch, "should not be nil")
		broker.AssertExpectations(t)
		con.AssertExpectations(t)
	})

	t.Run("Should perform a TLS connect to the specified Rabbit MQ host if tlsconf is present and return close channel", func(t *testing.T) {
		tlsConf := &tls.Config{}

		con := new(conMock)
		con.On("NotifyClose", nil).Return(make(chan *amqp.Error))

		broker := new(brokerMocker)
		broker.On("DialTLS", "amqps://localhost:5672/", tlsConf).Return(con, nil)

		target := NewConnectionManager(broker, tlsConf)

		ch, err := target.Connect("amqps://localhost:5672/")

		assert.NoError(t, err, "should not throw")
		assert.NotNil(t, ch, "should not be nil")
		broker.AssertExpectations(t)
		con.AssertExpectations(t)
	})

	t.Run("Should try at least 3 times to connect to Rabbit MQ before throwing error", func(t *testing.T) {
		con := new(conMock)

		broker := new(brokerMocker)
		broker.On("Dial", "amqp://user:pass@localhost:5672/").Return(con, errors.New("expected"))

		target := NewConnectionManager(broker, nil)

		ch, err := target.Connect("amqp://user:pass@localhost:5672/")

		assert.Error(t, err, "could not establish connection to Rabbit MQ Cluster")
		assert.Nil(t, ch, "should be nil")
		broker.AssertExpectations(t)
		broker.AssertNumberOfCalls(t, "Dial", 3)
	})
}

func TestConnectionManager_Disconnect(t *testing.T) {
	t.Run("Should disconnect from Rabbit MQ and clear connection", func(t *testing.T) {
		con := new(conMock)
		con.On("Close", nil).Return(nil)

		manager := ConnectionManager{
			con:  con,
			lock: sync.RWMutex{},
		}

		manager.Disconnect()
		assert.Nil(t, manager.con, "should be nil")
		con.AssertExpectations(t)
	})

	t.Run("Should disconnect from Rabbit MQ and clear connection", func(t *testing.T) {
		con := new(conMock)
		con.On("Close", nil).Return(errors.New("some error"))

		manager := ConnectionManager{
			con:  con,
			lock: sync.RWMutex{},
		}

		manager.Disconnect()
		assert.Nil(t, manager.con, "should be nil")
		con.AssertExpectations(t)
	})
}

func TestConnectionManager_Channel(t *testing.T) {
	con := new(conMock)
	con.On("Channel", nil).Return(&amqp.Channel{}, nil)

	manager := ConnectionManager{
		con:  con,
		lock: sync.RWMutex{},
	}

	ch, err := manager.Channel()

	assert.EqualValues(t, ch, &amqp.Channel{})
	assert.NoError(t, err, "should not throw")

	con.AssertExpectations(t)
}
