# Example Producer

The following code is a minimal producer please see below for usage instructions.

## Setup

1. Navigate into `producer` folder and get dependencies using `$ go mod download`
2. (Optional) Build executable using `$ go build`

## Usage

Following parameters are available:

* connection := full RabbitMQ connection url like `amqp://user:pass@localhost:5672`
* exchange := name of the exchange
* topic := a single routing key/topic
* amount := amount of messages to publish

`$ ./rabbitmq-producer(.exe) --topic your_topic --exchange Ex --amount 5`
