// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package config

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
const EnvMQExchange = "RMQ_EXCHANGE"
const EnvMQQueue = "RMQ_QUEUE"
const EnvTopicRefreshTime = "TOPIC_MAP_REFRESH_TIME"

func GetOpenFaaSUrl() string {
	return readFromEnv(EnvFaaSGWUrl, "http://gateway:8080")
}

func GetExchangeName() string {
	return readFromEnv(EnvMQExchange, "OpenFaasEx")
}

func GetQueueName() string {
	return readFromEnv(EnvMQQueue, "OpenFaaSQueue")
}

func GetRefreshTime() time.Duration {
	refreshTime, err := time.ParseDuration(readFromEnv(EnvTopicRefreshTime, "30s"))
	if err != nil {
		log.Println("Provided Topicmap Refresh Time was not a valid Duration, like 30s or 60ms. Falling back to 30s")
		refreshTime, _ = time.ParseDuration("30s")
	}

	return refreshTime
}

func GenerateRabbitMQUrl() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		readFromEnv(EnvMQUser, "guest"),
		readFromEnv(EnvMQPass, "guest"),
		readFromEnv(EnvMQHost, "localhost"),
		readFromEnv(EnvMQPort, "5672"),
	)
}

func GetTopics() []string {
	topicsString := readFromEnv(EnvMQTopics, "")
	topics := strings.Split(topicsString, ",")

	if len(topics) == 0 {
		log.Panicf("No Topic was specified. Provide them via Env RMQ_TOPICS=account,billing,support")
	}

	return topics
}

// Helper Functions
func readFromEnv(env string, fallback string) string {
	if val, exists := os.LookupEnv(env); exists {
		return val
	} else {
		return fallback
	}
}
