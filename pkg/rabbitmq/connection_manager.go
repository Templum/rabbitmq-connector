/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"crypto/tls"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Connector is a high level interface for connection related methods
type Connector interface {
	Connect(connectionURL string) (<-chan *amqp.Error, error)
	Disconnect()
}

// Manager is a interface that combines the relevant methods to connect to Rabbit MQ
// And create a new channel on an existing connection.
type Manager interface {
	Connector
	ChannelCreator
}

// ConnectionManager is tasked with managing the connection Rabbit MQ
type ConnectionManager struct {
	con     RBConnection
	lock    sync.RWMutex
	dialer  RBDialer
	tlsConf *tls.Config
}

// NewConnectionManager creates a new instance using the provided dialer
func NewConnectionManager(dialer RBDialer, conf *tls.Config) Manager {
	return &ConnectionManager{
		lock:    sync.RWMutex{},
		con:     nil,
		dialer:  dialer,
		tlsConf: conf,
	}
}

// Connect uses the provided connection urls and tries up to 3 times to establish a connection.
// The retries are performed exponentially starting with 2s. It also creates a listener for close notifications.
func (m *ConnectionManager) Connect(connectionURL string) (<-chan *amqp.Error, error) {
	for attempt := 0; attempt < 3; attempt++ {

		con, err := m.dial(connectionURL)

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

// Disconnect closes the connection and frees up the reference
func (m *ConnectionManager) Disconnect() {
	m.lock.Lock()

	err := m.con.Close()
	if err != nil {
		log.Printf("Received %s during closing connection", err)
	}

	m.con = nil
	m.lock.Unlock()
}

// Channel creates a new Rabbit MQ channel on the existing connection
func (m *ConnectionManager) Channel() (RabbitChannel, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.con.Channel()
}

func (m *ConnectionManager) dial(connectionURL string) (RBConnection, error) {
	if m.tlsConf == nil {
		return m.dialer.Dial(connectionURL)
	}

	return m.dialer.DialTLS(connectionURL, m.tlsConf)
}
