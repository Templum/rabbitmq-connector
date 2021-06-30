/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package types

import (
	"github.com/streadway/amqp"
)

// OpenFaaSInvocation represent an Event Specification used during invocation
type OpenFaaSInvocation struct {
	ContentType     string
	ContentEncoding string
	Topic           string
	Message         *[]byte
}

// NewInvocation creates a OpenFaaSInvocation from an amqp.Delivery.
func NewInvocation(delivery amqp.Delivery) *OpenFaaSInvocation {
	return &OpenFaaSInvocation{
		ContentType:     delivery.ContentType,
		ContentEncoding: delivery.ContentEncoding,
		Topic:           delivery.RoutingKey,
		Message:         &delivery.Body,
	}
}
