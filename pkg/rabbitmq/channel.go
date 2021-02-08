/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"crypto/tls"

	"github.com/streadway/amqp"
)

type ChannelCreator interface {
	Channel() (RabbitChannel, error)
}

type ChannelConsumer interface {
	Consume(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	NotifyClose(c chan *amqp.Error) chan *amqp.Error
	Close() error
}

type ExchangeHandler interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
}

type QueueHandler interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
}

type RBDialer interface {
	Dial(url string) (*amqp.Connection, error)
	DialTLS(url string, amqps *tls.Config) (*amqp.Connection, error)
}

type RabbitChannel interface {
	ExchangeHandler
	QueueHandler
	ChannelConsumer
}
