// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

const EnvMQHost = "MQ_HOST"
const EnvMQPort = "MQ_PORT"
const EnvMQUser = "MQ_USER"
const EnvMQPass = "MQ_PASS"

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func getConnectionUri() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		readFromEnv(EnvMQUser, "user"),
		readFromEnv(EnvMQPass, "pass"),
		readFromEnv(EnvMQHost, "localhost"),
		readFromEnv(EnvMQPort, "5672"),
	)
}

func readFromEnv(env string, fallback string) string {
	if val, exists := os.LookupEnv(env); exists {
		return val
	} else {
		return fallback
	}
}

func main() {
	topic := "account"

	con, err := amqp.Dial(getConnectionUri())
	failOnError(err, "Failed to connect to RabbitMQ")
	defer con.Close()

	ch, err := con.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"OpenFaasEx",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare an exchange")
	for {

		body := "Account Related Info"

		err = ch.Publish(
			"OpenFaasEx", // exchange
			topic,        // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")

		log.Printf(" Sent [%s] on Topic [%s]", body, topic)

		time.Sleep(2 * time.Second)
	}
}
