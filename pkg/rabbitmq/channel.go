/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"crypto/tls"

	"github.com/streadway/amqp"
)

// ChannelCreator interface allows the generations of channels
type ChannelCreator interface {
	Channel() (RabbitChannel, error)
}

// ChannelConsumer are interacting on channels
type ChannelConsumer interface {
	Consume(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	NotifyClose(c chan *amqp.Error) chan *amqp.Error
	Close() error
}

// ExchangeHandler offers a interface for the decleration of an exchange or the validation against existing exchanges
// on the RabbitMQ cluster
type ExchangeHandler interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
}

// QueueHandler offers a interface for the decleration & binding of an queues. Further it allows the validation against existing queues
// on the RabbitMQ cluster
type QueueHandler interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
}

// RBDialer is a abstraction of the RabbitMQ Dial methods
type RBDialer interface {
	Dial(url string) (RBConnection, error)
	DialTLS(url string, conf *tls.Config) (RBConnection, error)
}

// RabbitChannel is a abstraction of a RabbitMQ Channel
type RabbitChannel interface {
	ExchangeHandler
	QueueHandler
	ChannelConsumer
}

// RBConnection is a abstraction of a RabbitMQ Connection
type RBConnection interface {
	NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
	Close() error
	Channel() (*amqp.Channel, error)
}
