// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"log"

	"github.com/Templum/rabbitmq-connector/config"
	"github.com/streadway/amqp"
	"time"
	"github.com/Templum/rabbitmq-connector/openfaas"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func makeConsumer(conf config.ConnectorConfig) {
	con, err := amqp.Dial(conf.RabbitMQConnectionURI)
	failOnError(err, "Failed to create a connection")
	defer con.Close()

	ch, err := con.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		conf.ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare an exchange")

	queue, err := ch.QueueDeclare(
		conf.QueueName,
		false,
		false,
		true,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	for _, topic := range conf.Topics {
		log.Printf("Binding queue %s to exchange %s with Topic: %s", conf.QueueName, conf.ExchangeName, topic)
		err = ch.QueueBind(
			conf.QueueName,
			topic,
			conf.ExchangeName,
			false,
			nil,
		)
		failOnError(err, fmt.Sprintf("Failed to bind %s", conf.QueueName))
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
	failOnError(err, "Failed to register OpenFaaS Consumer")

	forever := make(chan bool)

	go func() {
		for msg := range messages {
			log.Printf("Received Message [%s] on Topic [%s] of Type [%s]", msg.Body, msg.RoutingKey, msg.ContentType)
		}
	}()

	<-forever
}

func emitMessagesOnTopic() {
	conf := config.BuildConnectorConfig()
	con, err := amqp.Dial(conf.RabbitMQConnectionURI)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer con.Close()

	ch, err := con.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"OpenFaaS-Topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare an exchange")
	for {

		body := "Awesome Message"

		err = ch.Publish(
			conf.ExchangeName, // exchange
			"not",             // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")

		log.Printf(" [x] Sent %s", body)

		time.Sleep(2 * time.Second)
	}
}

func synchronizeLookupTable(ticker *time.Ticker,
	lookupBuilder *openfaas.FunctionLookupBuilder,
	topicMap *openfaas.TopicLookupTable) {

	for {
		<-ticker.C
		topics, err := lookupBuilder.Build()
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Syncing topic map")
		topicMap.Sync(&topics)
	}
}


func main() {
	conf := config.BuildConnectorConfig()

	topicMap := openfaas.TopicLookupTable{}

	lookupBuilder  := openfaas.FunctionLookupBuilder{
		GatewayURL:conf.GatewayURL,
		Client: openfaas.MakeClient(30 * time.Second ), // TODO: Read in from conf
	}

	ticker := time.NewTicker(5 * time.Second) // TODO: Read in from conf
	go synchronizeLookupTable(ticker, &lookupBuilder, &topicMap)

	go makeConsumer(conf)

	go emitMessagesOnTopic()
	forever := make(chan bool)
	<-forever
}
