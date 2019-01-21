package types

import (
	"fmt"
	"github.com/streadway/amqp"
)

// OpenFaaSInvocation represent an Event Specification used during invocation
type OpenFaaSInvocation struct {
	Topic   string
	Message *[]byte
	tag     uint64
	done    amqp.Acknowledger
}

func (o *OpenFaaSInvocation) Finished() {
	if o.done != nil {
		err := o.done.Ack(o.tag, false)
		if err != nil {
			fmt.Printf("Recieved %s during ack for tag %d", err, o.tag)
		}
	}
}

func NewInvocation(delivery amqp.Delivery) *OpenFaaSInvocation {
	return &OpenFaaSInvocation{
		Topic:   delivery.RoutingKey,
		Message: &delivery.Body,
		tag:     delivery.DeliveryTag,
		done:    delivery.Acknowledger,
	}
}
