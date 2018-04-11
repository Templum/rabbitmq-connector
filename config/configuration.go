// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const gatewayUrl = "gateway_url"
const rabbitTopics = "topics"
const rabbitHost = "rabbit_mq_host"
const rabbitPort = "rabbit_mq_port"
const rabbitUser = "rabbit_mq_user"
const rabbitPass = "rabbit_mq_pass"

type connectorConfig struct {
	GatewayURL            string
	Topics                []string
	RabbitMQConnectionURI string
}

func BuildConnectorConfig() connectorConfig {
	gatewayURL := "http://gateway:8080"
	if val, exists := os.LookupEnv(gatewayUrl); exists {
		gatewayURL = val
	}

	topics := []string{}
	if val, exists := os.LookupEnv(rabbitTopics); exists {
		for _, topic := range strings.Split(val, ",") {
			if len(topic) > 0 {
				topics = append(topics, topic)
			}
		}
	}

	if len(topics) == 0 {
		log.Fatal(`Provide a list of topics i.e. topics="payment_published,slack_joined"`)
	}

	rabbitURI := generateRabbitConnectionUri()

	return connectorConfig{
		GatewayURL:            gatewayURL,
		Topics:                topics,
		RabbitMQConnectionURI: rabbitURI,
	}
}

// Generates the connection string to RabbitMQ based on
// information provided through environment variables
func generateRabbitConnectionUri() string {
	host := "rabbitqueue"
	if val, exists := os.LookupEnv(rabbitHost); exists {
		host = val
	}

	port := "5672"
	if val, exists := os.LookupEnv(rabbitPort); exists {
		port = val
	}

	user := "user"
	if val, exists := os.LookupEnv(rabbitUser); exists {
		port = val
	}

	pass := "user"
	if val, exists := os.LookupEnv(rabbitPass); exists {
		port = val
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)
}
