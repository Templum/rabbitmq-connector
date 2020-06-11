/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package types

import (
	"github.com/streadway/amqp"
)

// OpenFaaSInvocation represent an Event Specification used during invocation
type OpenFaaSInvocation struct {
	Topic   string
	Message *[]byte
}

// NewInvocation creates a OpenFaaSInvocation from an amqp.Delivery.
func NewInvocation(delivery amqp.Delivery) OpenFaaSInvocation {
	return OpenFaaSInvocation{
		Topic:   delivery.RoutingKey,
		Message: &delivery.Body,
	}
}
