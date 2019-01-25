package types

import (
	"github.com/streadway/amqp"
)

// OpenFaaSInvocation represent an Event Specification used during invocation
type OpenFaaSInvocation struct {
	Topic   string
	Message *[]byte
}

func NewInvocation(delivery amqp.Delivery) *OpenFaaSInvocation {
	return &OpenFaaSInvocation{
		Topic:   delivery.RoutingKey,
		Message: &delivery.Body,
	}
}
