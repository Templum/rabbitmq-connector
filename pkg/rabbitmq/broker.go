/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"crypto/tls"

	"github.com/streadway/amqp"
)

func NewBroker() RBDialer {
	return &Broker{}
}

type Broker struct{}

func (b *Broker) Dial(url string) (RBConnection, error) {
	return amqp.Dial(url)
}

func (b *Broker) DialTLS(url string, conf *tls.Config) (RBConnection, error) {
	test, err := amqp.DialTLS(url, conf)
	return test, err
}
