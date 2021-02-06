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
	Reconnect(connectionUrl string) (<-chan *amqp.Error, error)
	Disconnect()
}

type Manager interface {
	Connector
	RBChannelCreator
}

type ConnectionManager struct {
	con  *amqp.Connection // TODO: Create Top Level interface as discussed here: https://github.com/streadway/amqp/issues/164#issuecomment-271523006
	lock sync.RWMutex
}

func NewConnectionManager() Manager {
	return &ConnectionManager{
		lock: sync.RWMutex{},
		con:  nil,
	}
}

func (m *ConnectionManager) Connect(connectionUrl string) (<-chan *amqp.Error, error) {
	for attempt := 0; attempt < 3; attempt++ {

		con, err := amqp.Dial(connectionUrl)

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

func (m *ConnectionManager) Reconnect(connectionUrl string) (<-chan *amqp.Error, error) {
	m.lock.Lock()
	// We ignore error here since we anyways create a new connection
	_ = m.con.Close()
	m.con = nil
	m.lock.Unlock()

	return m.Connect(connectionUrl) // TODO: Need to test how this behaves when running life
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

func (m *ConnectionManager) Channel() (*amqp.Channel, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.con.Channel()
}
