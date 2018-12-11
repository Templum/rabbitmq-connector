// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package rabbitmq

import (
	"log"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	Generator "github.com/docker/docker/pkg/namesgenerator"
	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/streadway/amqp"
)

type worker struct {
	con     *amqp.Connection
	channel *amqp.Channel

	client *types.Controller

	name      string // Only for better debugging
	queueName string
	exchange  string
	topic     string

	closed bool
}

func NewWorker(con *amqp.Connection, client *types.Controller, topic string) *worker {
	return &worker{
		con,
		nil,

		client,

		Generator.GetRandomName(2),
		config.GetQueueName(),
		config.GetExchangeName(),
		topic,

		false,
	}
}

type Worker interface {
	Start()
	Close()
	processMessage(deliveries <-chan amqp.Delivery)
}

func (w *worker) Start() {
	log.Printf("Starting Init Process of Worker %s for topic %s", w.name, w.topic)
	err := w.initializeChannel()

	if err != nil {
		log.Printf("Init Process for Topic: %s failed with %s. Aborting Init Process for Worker %s", w.topic, err, w.name)
		return
	}

	// Related to Self Healing
	errorChannel := make(chan *amqp.Error)
	w.channel.NotifyClose(errorChannel)
	go handleError(w, errorChannel)

	deliveries, err := w.startConsuming()

	if err != nil {
		log.Printf("Worker %s not able to consume messages", w.name)
		return
	} else {
		log.Printf("Worker %s started consuming messages for %s", w.name, w.topic)
	}
	go w.processMessage(deliveries)
}

func (w *worker) Close() {
	log.Printf("Stopping Worker for Topic %s", w.topic)
	w.closed = true
	err := w.channel.Close()

	if err != nil {
		log.Printf("Recieved %s during channel closing", err)
	}
}

// initializeChannel will create new channel using the stored connection. Further it performs
// a variety of initialization tasks, including Exchange & Queue Declaration and also the binding.
// It will return an error if the initialization failed.
func (w *worker) initializeChannel() error {
	var err error
	w.channel, err = openChannel(w.con, 3)

	if err != nil {
		log.Printf("Failed to start Worker for Topic %s due to %s", w.topic, err)
		return err
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
	}

	log.Printf("Binding Queue %s to Exchange %s for Topic: %s", w.queueName, w.exchange, w.topic)
	err = w.channel.QueueBind(
		w.queueName,
		w.topic,
		w.exchange,
		false,
		nil,
	)
	return err
}

// startConsuming will start consuming messages on the specified queue. It returns a channel containing
// the received messages or an error
func (w *worker) startConsuming() (<-chan amqp.Delivery, error) {
	return w.channel.Consume(
		w.queueName,
		"", // Let Rabbit MQ Generate a Tag
		true,
		false,
		false,
		false,
		nil,
	)
}

// processMessage will listen on the provide channel indefinitely until the channel is closed.
// Received messages will be filtered by RoutingKey and then forwarded to OpenFaaS
func (w *worker) processMessage(deliveries <-chan amqp.Delivery) {
	for message := range deliveries {
		if message.RoutingKey == w.topic {
			log.Printf("Recieved message for Topic %s on Type %s", w.topic, w.name)
			go w.client.Invoker.Invoke(w.client.TopicMap, w.topic, &message.Body)
		}
	}
	log.Printf("Message Channel of Worker %s was closed", w.name)
}

// handleError currently only prints the error if it is not nil.
func handleError(_ Worker, errorStream chan *amqp.Error) {
	for {
		err := <-errorStream
		if err != nil {
			log.Printf("Worker recieved the following error %s", err)
		}
	}
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
