// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package rabbitmq

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const EnvFaaSGWUrl = "OPEN_FAAS_GW_URL"
const EnvMQTopics = "RMQ_TOPICS"
const EnvMQHost = "RMQ_HOST"
const EnvMQPort = "RMQ_PORT"
const EnvMQUser = "RMQ_USER"
const EnvMQPass = "RMQ_PASS"
const EnvMQQueue = "RMQ_QUEUE"
const EnvMQExchange = "RMQ_EXCHANGE"
const EnvReqTimeout = "REQ_TIMEOUT"
const EnvTopicRefreshTime = "TOPIC_MAP_REFRESH_TIME"

type Config struct {
	GatewayURL            string
	Topics                []string
	RabbitMQConnectionURI string
	ExchangeName          string
	QueueName             string
	RequestTimeout        time.Duration
	TopicMapRefreshTime   time.Duration
}

func readFromEnv(env string, fallback string) string {
	if val, exists := os.LookupEnv(env); exists {
		return val
	} else {
		return fallback
	}
}

// Reads in the configuration from environment variables
// If nothing is provided it will fallback to default values
func BuildConfig() Config {
	var topics []string

	for _, topic := range strings.Split(readFromEnv(EnvMQTopics, ""), ",") {
		if len(topic) > 0 {
			topics = append(topics, topic)
		}
	}

	if len(topics) == 0 {
		log.Fatal(`Provide a list of topics i.e. topics="payment_published,slack_joined"`)
	}

	requestTimeout, err := time.ParseDuration(readFromEnv(EnvReqTimeout, "30s"))
	if err != nil {
		log.Println("Provided Request Timeout was not a valid Duration, like 30s or 60ms. Falling back to 30s")
		requestTimeout, _ = time.ParseDuration("30s")
	}

	refreshTime, err := time.ParseDuration(readFromEnv(EnvTopicRefreshTime, "60s"))
	if err != nil {
		log.Println("Provided Topicmap Refresh Time was not a valid Duration, like 30s or 60ms. Falling back to 60s")
		refreshTime, _ = time.ParseDuration("60s")
	}

	return Config{
		GatewayURL:            readFromEnv(EnvFaaSGWUrl, "http://gateway:8080"),
		Topics:                topics,
		RabbitMQConnectionURI: generateRabbitConnectionUri(),
		ExchangeName:          readFromEnv(EnvMQExchange, "OpenFaasEx"),
		QueueName:             readFromEnv(EnvMQQueue, "OpenFaasQueue"),
		RequestTimeout:        requestTimeout,
		TopicMapRefreshTime:   refreshTime,
	}
}

// Generates the connection string to RabbitMQ based on
// information provided through environment variables
func generateRabbitConnectionUri() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		readFromEnv(EnvMQUser, "guest"),
		readFromEnv(EnvMQPass, "guest"),
		readFromEnv(EnvMQHost, "localhost"),
		readFromEnv(EnvMQPort, "5672"),
	)
}
