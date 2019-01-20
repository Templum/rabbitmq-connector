// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Controller struct {
	GatewayUrl          string
	RabbitConnectionUrl string
	RabbitSanitizedUrl  string

	ExchangeName string
	QueueName    string
	Topics       []string

	TopicRefreshTime time.Duration
}

func NewConfig() (*Controller, error) {
	gatewayUrl, err := getOpenFaaSUrl()
	if err != nil {
		return nil, err
	}

	rabbitUrl, err := getRabbitMQConnectionUrl()
	sanitizedUrl := getSanitizedRabbitMQUrl()
	if err != nil {
		return nil, err
	}

	queue := readFromEnv(envRabbitQueue, "OpenFaaSQueue")
	exchange := readFromEnv(envRabbitExchange, "OpenFaasEx")

	topics, err := getTopics()
	if err != nil {
		return nil, err
	}

	return &Controller{
		GatewayUrl:          gatewayUrl,
		RabbitConnectionUrl: rabbitUrl,
		RabbitSanitizedUrl:  sanitizedUrl,

		ExchangeName: exchange,
		QueueName:    queue,
		Topics:       topics,

		TopicRefreshTime: getRefreshTime(),
	}, nil
}

const envFaaSGwUrl = "OPEN_FAAS_GW_URL"
const envRabbitUser = "RMQ_USER"
const envRabbitPass = "RMQ_PASS"
const envRabbitHost = "RMQ_HOST"
const envRabbitPort = "RMQ_PORT"

const envRabbitTopics = "RMQ_TOPICS"
const envRabbitExchange = "RMQ_EXCHANGE"
const envRabbitQueue = "RMQ_QUEUE"

const envRefreshTime = "TOPIC_MAP_REFRESH_TIME"

func getOpenFaaSUrl() (string, error) {
	url := readFromEnv(envFaaSGwUrl, "http://gateway:8080")
	if !(strings.HasPrefix(url, "http://")) && !(strings.HasPrefix(url, "https://")) {
		message := fmt.Sprintf("Provided url %s does not include the protocoll http / https", url)
		return "", errors.New(message)
	}
	return url, nil
}

func getRabbitMQConnectionUrl() (string, error) {
	user := readFromEnv(envRabbitUser, "guest")
	pass := readFromEnv(envRabbitPass, "guest")
	host := readFromEnv(envRabbitHost, "localhost")
	port := readFromEnv(envRabbitPort, "5672")

	parsedPort, err := strconv.Atoi(port)

	if err != nil {
		message := fmt.Sprintf("Provided port %s is not a valid port", port)
		return "", errors.New(message)
	}

	if parsedPort <= 0 || parsedPort > 65535 {
		message := fmt.Sprintf("Provided port %s is outside of the allowed port range", port)
		return "", errors.New(message)
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port), nil
}

func getSanitizedRabbitMQUrl() string {
	host := readFromEnv(envRabbitHost, "localhost")
	port := readFromEnv(envRabbitPort, "5672")
	return fmt.Sprintf("amqp://%s:%s", host, port)
}

func getTopics() ([]string, error) {
	topicsString := readFromEnv(envRabbitTopics, "")
	topics := strings.Split(topicsString, ",")

	if topicsString == "" || len(topics) < 1 {
		return nil, errors.New("No Topic was specified. Provide them via Env RMQ_TOPICS=account,billing,support")
	} else {
		return topics, nil
	}
}

func getRefreshTime() time.Duration {
	refreshTime, err := time.ParseDuration(readFromEnv(envRefreshTime, "30s"))
	if err != nil {
		log.Println("Provided Topicmap Refresh Time was not a valid Duration, like 30s or 60ms. Falling back to 30s")
		refreshTime, _ = time.ParseDuration("30s")
	}

	return refreshTime
}

// Helper Functions
func readFromEnv(env string, fallback string) string {
	if val, exists := os.LookupEnv(env); exists {
		return val
	} else {
		return fallback
	}
}
