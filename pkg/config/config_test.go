// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package config

import (
	"os"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Run("With invalid Gateway Url", func(t *testing.T) {
		os.Setenv("OPEN_FAAS_GW_URL", "gateway:8080")
		defer os.Unsetenv("OPEN_FAAS_GW_URL")

		var err error

		_, err = NewConfig()
		if !strings.Contains(err.Error(), "does not include the protocol http / https") {
			t.Errorf("Did not throw new correct error. Recieved %s", err)
		}

		os.Setenv("OPEN_FAAS_GW_URL", "tcp://gateway:8080")
		_, err = NewConfig()
		if !strings.Contains(err.Error(), "does not include the protocol http / https") {
			t.Errorf("Did not throw new correct error. Recieved %s", err)
		}
	})

	t.Run("With invalid Rabbit MQ Port", func(t *testing.T) {
		os.Setenv("RMQ_PORT", "is_string")
		defer os.Unsetenv("RMQ_PORT")

		var err error

		_, err = NewConfig()
		if !strings.Contains(err.Error(), "is not a valid port") {
			t.Errorf("Did not throw new correct error. Recieved %s", err)
		}

		os.Setenv("RMQ_PORT", "-1")
		_, err = NewConfig()
		if !strings.Contains(err.Error(), "is outside of the allowed port range") {
			t.Errorf("Did not throw new correct error. Recieved %s", err)
		}

		os.Setenv("RMQ_PORT", "65536")
		_, err = NewConfig()
		if !strings.Contains(err.Error(), "is outside of the allowed port range") {
			t.Errorf("Did not throw new correct error. Recieved %s", err)
		}
	})

	t.Run("Empty Topics", func(t *testing.T) {
		os.Setenv("RMQ_TOPICS", "")
		defer os.Unsetenv("RMQ_TOPICS")

		_, err := NewConfig()
		if !strings.Contains(err.Error(), "No Topic was specified. Provide them via Env RMQ_TOPICS=account,billing,support.") {
			t.Errorf("Did not throw new correct error. Recieved %s", err)
		}
	})

	t.Run("Default Config", func(t *testing.T) {
		os.Setenv("RMQ_TOPICS", "test")
		defer os.Unsetenv("RMQ_TOPICS")

		config, err := NewConfig()
		if err != nil {
			t.Error("Should not throw an error")
		}

		if config.GatewayUrl != "http://gateway:8080" {
			t.Errorf("Expected http://gateway:8080 recieved %s", config.GatewayUrl)
		}

		if config.RabbitConnectionUrl != "amqp://guest:guest@localhost:5672/" {
			t.Errorf("Expected amqp://guest:guest@localhost:5672/ recieved %s", config.RabbitConnectionUrl)
		}

		if config.RabbitSanitizedUrl != "amqp://localhost:5672" {
			t.Errorf("Expected amqp://localhost:5672 recieved %s", config.RabbitSanitizedUrl)
		}

		if config.ExchangeName != "OpenFaasEx" {
			t.Errorf("Expected OpenFaasEx recieved %s", config.ExchangeName)
		}

		if config.QueueName != "OpenFaaSQueue" {
			t.Errorf("Expected OpenFaaSQueue recieved %s", config.QueueName)
		}

		if len(config.Topics) != 1 {
			t.Errorf("Expected 1 Topic recieved %d Topic", len(config.Topics))
		}

		if config.TopicRefreshTime.Seconds() != 30 {
			t.Errorf("Expected 30s recieved %fs", config.TopicRefreshTime.Seconds())
		}
	})

	t.Run("Override Config", func(t *testing.T) {
		os.Setenv("RMQ_TOPICS", "test")
		os.Setenv("RMQ_QUEUE", "Queue")
		os.Setenv("RMQ_EXCHANGE", "Ex")
		os.Setenv("RMQ_HOST", "rabbit")
		os.Setenv("RMQ_PORT", "1337")
		os.Setenv("RMQ_USER", "username")
		os.Setenv("RMQ_PASS", "password")
		os.Setenv("OPEN_FAAS_GW_URL", "https://gateway")
		os.Setenv("TOPIC_MAP_REFRESH_TIME", "40s")

		defer os.Unsetenv("RMQ_TOPICS")
		defer os.Unsetenv("RMQ_QUEUE")
		defer os.Unsetenv("RMQ_EXCHANGE")
		defer os.Unsetenv("RMQ_HOST")
		defer os.Unsetenv("RMQ_PORT")
		defer os.Unsetenv("RMQ_USER")
		defer os.Unsetenv("RMQ_PASS")
		defer os.Unsetenv("OPEN_FAAS_GW_URL")
		defer os.Unsetenv("TOPIC_MAP_REFRESH_TIME")

		config, err := NewConfig()
		if err != nil {
			t.Error("Should not throw an error")
		}

		if config.GatewayUrl != "https://gateway" {
			t.Errorf("Expected https://gateway recieved %s", config.GatewayUrl)
		}

		if config.RabbitConnectionUrl != "amqp://username:password@rabbit:1337/" {
			t.Errorf("Expected amqp://username:password@rabbit:1337/ recieved %s", config.RabbitConnectionUrl)
		}

		if config.RabbitSanitizedUrl != "amqp://rabbit:1337" {
			t.Errorf("Expected amqp://rabbit:1337 recieved %s", config.RabbitSanitizedUrl)
		}

		if config.ExchangeName != "Ex" {
			t.Errorf("Expected Ex recieved %s", config.ExchangeName)
		}

		if config.QueueName != "Queue" {
			t.Errorf("Expected Queue recieved %s", config.QueueName)
		}

		if config.TopicRefreshTime.Seconds() != 40 {
			t.Errorf("Expected 40s recieved %fs", config.TopicRefreshTime.Seconds())
		}
	})
}
