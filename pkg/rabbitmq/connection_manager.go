/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

type Connector interface {
	Connect(connectionUrl string) error
	Reconnect()
	Disconnect()
}

type ChannelMaker interface {
	CreateChannel() (*amqp.Channel, error)
}

type Manager interface {
	Connector
	ChannelMaker
}

type ConnectionManager struct {
	url  string
	con  *amqp.Connection
	lock sync.RWMutex
}

func NewConnectionManager() Manager {
	return &ConnectionManager{
		lock: sync.RWMutex{},
		url:  "",
		con:  nil,
	}
}

func (m *ConnectionManager) Connect(connectionUrl string) error {
	for attempt := 0; attempt < 3; attempt++ {
		con, err := amqp.Dial(connectionUrl)

		if err == nil {
			log.Println("Successfully established connection to Rabbit MQ Cluster")
			m.lock.Lock()
			m.con = con
			m.url = connectionUrl
			m.lock.Unlock()

			closeChannel := make(chan *amqp.Error)
			con.NotifyClose(closeChannel)

			// Handling connection closes
			go func(closeChannel chan *amqp.Error) {
				received := <-closeChannel

				if received.Recover {
					log.Printf("Received non critical error %s.", received)
				} else {
					log.Printf("Received critical error %s.", received)
				}

				log.Println("Will now attempt to reconnect to Rabbit MQ Cluster")
				m.Reconnect()
			}(closeChannel)

			return nil
		}

		log.Printf("Failed to establish connection due to %s. Attempt: %d/3", err, attempt)
		time.Sleep(time.Duration(2*attempt+1) * time.Second)
	}

	return errors.New("could not establish connection to Rabbit MQ Cluster")
}

func (m *ConnectionManager) Reconnect() {
	m.lock.RLock()
	connectionUrl := m.url
	m.lock.RUnlock()

	if connectionUrl == "" {
		log.Println("Disconnect was called no need to reconnect")
		return
	}

	err := m.Connect(connectionUrl)

	if err != nil {
		log.Fatalf("Received fatal error %s during reconnecting to Rabbit MQ Cluster", err)
		return
	}
	log.Println("Successfully recovered connection Rabbit MQ Cluster")
}

func (m *ConnectionManager) CreateChannel() (*amqp.Channel, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.con == nil {
		return &amqp.Channel{}, errors.New("missing base connection to Rabbit MQ Cluster")
	}

	for attempt := 0; attempt < 3; attempt++ {
		channel, err := m.con.Channel()

		if err == nil {
			log.Println("Successfully created channel")
			return channel, nil
		}

		log.Printf("Failed to create channel on connection due to %s. Attempt: %d/3", err, attempt)
		time.Sleep(time.Duration(2*attempt+1) * time.Second)
	}

	return &amqp.Channel{}, errors.New("missing base connection to Rabbit MQ Cluster")
}

func (m *ConnectionManager) Disconnect() {
	m.lock.Lock()
	m.url = "" // Will use this as a flag to indicate reconnecting is not wanted

	err := m.con.Close()
	if err != nil {
		log.Printf("Received %s during closing connection", err)
	}

	m.con = nil
	m.lock.Unlock()
}
