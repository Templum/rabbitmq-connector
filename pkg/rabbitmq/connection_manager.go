/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Connector interface {
	Connect(connectionUrl string) (<-chan *amqp.Error, error)
	Disconnect()
}

type Manager interface {
	Connector
	ChannelCreator
}

type ConnectionManager struct {
	con    RBConnection
	lock   sync.RWMutex
	dialer RBDialer
}

func NewConnectionManager(dialer RBDialer) Manager {
	return &ConnectionManager{
		lock:   sync.RWMutex{},
		con:    nil,
		dialer: dialer,
	}
}

func (m *ConnectionManager) Connect(connectionUrl string) (<-chan *amqp.Error, error) {
	for attempt := 0; attempt < 3; attempt++ {

		con, err := m.dialer.Dial(connectionUrl)

		if err == nil {
			log.Println("Successfully established connection to Rabbit MQ Cluster")
			m.lock.Lock()
			m.con = con
			m.lock.Unlock()

			closeChannel := make(chan *amqp.Error)
			con.NotifyClose(closeChannel)
			return closeChannel, nil
		}

		log.Printf("Failed to establish connection due to %s. Attempt: %d/3", err, attempt)
		time.Sleep(time.Duration(2*attempt+1) * time.Second)
	}

	return nil, errors.New("could not establish connection to Rabbit MQ Cluster")
}

func (m *ConnectionManager) Disconnect() {
	m.lock.Lock()

	err := m.con.Close()
	if err != nil {
		log.Printf("Received %s during closing connection", err)
	}

	m.con = nil
	m.lock.Unlock()
}

func (m *ConnectionManager) Channel() (RabbitChannel, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.con.Channel()
}
