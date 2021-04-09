package main

import (
	"flag"
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

var (
	connectionString string
	exchangeName string
	topicName string
	amountOfMessages uint
)


func init() {

	flag.StringVar(&connectionString, "connection", "amqp://user:pass@localhost:5672/", "full RabbitMQ connection url like amqp://user:pass@localhost:5672")
	flag.StringVar(&exchangeName, "exchange", "AEx", "exchange name")
	flag.StringVar(&topicName, "topic", "Foo", "a single topic name/routing key")
	flag.UintVar(&amountOfMessages, "amount", 256, "amount of messages to publish")
	flag.Parse()
}


func main() {
	conn, err := amqp.Dial(connectionString)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	body := "Hello World!"

	for i := uint(0); i < amountOfMessages; i++ {
		err = ch.Publish(
			exchangeName, // exchange
			topicName, // routing key
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s", body)
	}

}
