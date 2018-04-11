// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"github.com/Templum/rabbitmq-connector/config"
	"os"
)

func failOnError(err error, msg string){
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func receiveMessagesFromTopic(){
	conf := config.BuildConnectorConfig()
	con, err := amqp.Dial(conf.RabbitMQConnectionURI)
	failOnError(err, "Failed to open a channel")
	defer con.Close()

	ch, err := con.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	//TODO: Build up connection pool for provided topics
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

	queue, err := ch.QueueDeclare(
		"OpenFaaS-Queue",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for _, s := range conf.Topics{
		log.Printf("Binding queue %s to exchange %s with routing key %s",
			queue.Name, "logs_topic", s)
		err = ch.QueueBind(
			queue.Name,       // queue name
			s,            // routing key
			"OpenFaaS-Topic", // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
	}

	msgs, err := ch.Consume(
		queue.Name, // queue
		"FaaS Worker",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}

func emitMessagesOnTopic(){
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	err = ch.Publish(
		"logs_topic",          // exchange
		"", // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", body)
}




func main() {
	receiveMessagesFromTopic()
}
