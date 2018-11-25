// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func ensureSuccess(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

type IncomingMessageHandler func(delivery amqp.Delivery)

func MakeConnector(config *Config, handler IncomingMessageHandler) {
	con, err := amqp.Dial(config.RabbitMQConnectionURI)
	ensureSuccess(err, "Failed to create a connection")
	defer con.Close()

	ch, err := con.Channel()
	ensureSuccess(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		config.ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	ensureSuccess(err, "Failed to declare an exchange")

	queue, err := ch.QueueDeclare(
		config.QueueName,
		false,
		false,
		true,
		false,
		nil,
	)
	ensureSuccess(err, "Failed to declare a queue")

	for _, topic := range config.Topics {
		log.Printf("Binding queue %s to exchange %s with Topic: %s", config.QueueName, config.ExchangeName, topic)
		err = ch.QueueBind(
			config.QueueName,
			topic,
			config.ExchangeName,
			false,
			nil,
		)
		ensureSuccess(err, fmt.Sprintf("Failed to bind %s", config.QueueName))
	}

	messages, err := ch.Consume(
		queue.Name,
		"OpenFaaS Consumer",
		true,
		false,
		false,
		false,
		nil,
	)
	ensureSuccess(err, "Failed to register OpenFaaS Consumer")

	forever := make(chan bool)

	go func() {
		for msg := range messages {
			handler(msg)
		}
	}()

	<-forever
}
