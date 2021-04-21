/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"crypto/tls"

	"github.com/streadway/amqp"
)

// NewBroker generates a new wrapper around the RabbitMQ Client lib
func NewBroker() RBDialer {
	return &Broker{}
}

// Broker is a wrapper around the RabbitMQ Client lib, which allows better
// unit testing. By abstracting away the RabbitMQ raw types, which are struct based.
type Broker struct{}

// Dial tries to connect to the providing url, returning either a RBConnection or
// the received connection error.
func (b *Broker) Dial(url string) (RBConnection, error) {
	return amqp.Dial(url)
}

// DialTLS tries to connect to the providing url using TLS, returning either a RBConnection or
// the received connection error.
func (b *Broker) DialTLS(url string, conf *tls.Config) (RBConnection, error) {
	test, err := amqp.DialTLS(url, conf)
	return test, err
}
