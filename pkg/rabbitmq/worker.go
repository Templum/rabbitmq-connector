// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package rabbitmq

import (
	"log"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/streadway/amqp"
)

type worker struct {
	con          *amqp.Connection
	channel      *amqp.Channel
	errorChannel chan *amqp.Error

	client *types.Controller

	queueName string
	exchange  string
	topic     string

	closed bool
}

func NewWorker(con *amqp.Connection, client *types.Controller, topic string) *worker {
	return &worker{
		con,
		nil,
		nil,

		client,

		topic,
		config.GetExchangeName(),
		topic,

		false,
	}
}

func (w *worker) Start() {
	log.Printf("Initializing Worker for Topic %s", w.topic)
	w.init()
}

func (w *worker) Close() {
	log.Printf("Stopping Worker for Topic %s", w.topic)
	w.closed = true
	w.channel.Close()
}

func (w *worker) init() {
	var err error
	w.channel, err = openChannel(w.con, 3)

	if err != nil {
		log.Printf("Failed to start Worker for Topic %s due to %s", w.topic, err)
		return
	}

	err = w.channel.ExchangeDeclare(
		w.exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Printf("Failed during Exchange Declaration for Topic %s due to %s", w.topic, err)
		return
	}

	_, err = w.channel.QueueDeclare(
		w.queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Printf("Failed during Queue Declaration for Topic %s due to %s", w.topic, err)
		return
	}

	log.Printf("Binding Queue %s to Exchange %s for Topic: %s", w.queueName, w.exchange, w.topic)
	err = w.channel.QueueBind(
		w.queueName,
		w.topic,
		w.exchange,
		false,
		nil,
	)

	if err != nil {
		log.Printf("Failed to Bind Queue %s to Exchange %s due to %s", w.queueName, w.exchange, err)
		return
	}

	_ = w.channel.Qos(100, 0, false)

	// TODO: Self Healing on Channel Level

	deliveries, err := w.channel.Consume(
		w.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Printf("Worker is not able to consume messages for Topic %s due to %s", w.topic, err)
		return
	}

	log.Printf("Successfully started Worker on Queue %s for Topic %s", w.queueName, w.topic)
	go w.handleMessages(deliveries)
}

func (w *worker) handleMessages(deliveries <-chan amqp.Delivery) {
	for message := range deliveries {
		log.Printf("Recieved message on Topic %s of Type %s", w.topic, message.ContentType)
		go w.client.Invoker.Invoke(w.client.TopicMap, w.topic, &message.Body)
	}
	log.Println("Channel was closed")
}

func openChannel(con *amqp.Connection, retries int) (*amqp.Channel, error) {
	channel, err := con.Channel()

	if err != nil && retries > 0 {
		log.Printf("Worker was not able to open a channel. Recieved error %s. Retries left %d", err, retries)
		time.Sleep(2 * time.Second)
		return openChannel(con, retries-1)
	}
	return channel, err
}
