package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetOpenFaaSUrl(t *testing.T) {
	t.Run("Read Valid URL", func(t *testing.T) {
		const VALID_URL = "https://openfaas-gw:8080"
		os.Setenv(EnvFaaSGWUrl, VALID_URL)
		defer os.Unsetenv(EnvFaaSGWUrl)

		result := GetOpenFaaSUrl()

		if result != VALID_URL {
			t.Errorf("Expected %s recieved %s", VALID_URL, result)
		}
	})

	t.Run("Read Invalid URL", func(t *testing.T) {
		const INVALID_URL = "openfaas-gw:8080"
		os.Setenv(EnvFaaSGWUrl, INVALID_URL)
		defer os.Unsetenv(EnvFaaSGWUrl)
		defer func() {
			if r := recover(); r == nil {
				t.Error("Did not panic")
			}
		}()
		GetOpenFaaSUrl()
	})

	t.Run("Read Default URL", func(t *testing.T) {
		const DEFAULT_URL = "http://gateway:8080"
		os.Unsetenv(EnvFaaSGWUrl)

		result := GetOpenFaaSUrl()

		if result != DEFAULT_URL {
			t.Errorf("Expected %s recieved %s", DEFAULT_URL, result)
		}
	})
}

func TestGetExchangeName(t *testing.T) {
	t.Run("Read Exchange", func(t *testing.T) {
		const EXCHANGE_NAME = "AwesomeExchange"
		os.Setenv(EnvMQExchange, EXCHANGE_NAME)
		defer os.Unsetenv(EnvMQExchange)

		result := GetExchangeName()

		if result != EXCHANGE_NAME {
			t.Errorf("Expected %s recieved %s", EXCHANGE_NAME, result)
		}
	})

	t.Run("Read Default", func(t *testing.T) {
		const DEFAULT_NAME = "OpenFaasEx"
		os.Unsetenv(EnvMQExchange)

		result := GetExchangeName()

		if result != DEFAULT_NAME {
			t.Errorf("Expected %s recieved %s", DEFAULT_NAME, result)
		}
	})
}

func TestGetQueueName(t *testing.T) {
	t.Run("Read Queue", func(t *testing.T) {
		const QUEUE_NAME = "MyAwesomeQueue"
		os.Setenv(EnvMQQueue, QUEUE_NAME)
		defer os.Unsetenv(EnvMQQueue)

		result := GetQueueName()

		if result != QUEUE_NAME {
			t.Errorf("Expected %s recieved %s", QUEUE_NAME, result)
		}
	})

	t.Run("Read Default", func(t *testing.T) {
		const DEFAULT_NAME = "OpenFaaSQueue"
		os.Unsetenv(EnvMQQueue)

		result := GetQueueName()

		if result != DEFAULT_NAME {
			t.Errorf("Expected %s recieved %s", DEFAULT_NAME, result)
		}
	})
}

func TestGetRefreshTime(t *testing.T) {
	t.Run("Read Valid Time", func(t *testing.T) {
		const VALID_TIME = "1337s"
		os.Setenv(EnvTopicRefreshTime, VALID_TIME)
		defer os.Unsetenv(EnvTopicRefreshTime)

		result := GetRefreshTime()

		if expected, _ := time.ParseDuration(VALID_TIME); result != expected {
			t.Errorf("Expected %s recieved %s", expected, result)
		}
	})

	t.Run("Read Invalid Time", func(t *testing.T) {
		const INVALID_TIME = "asdasdasd"
		os.Setenv(EnvTopicRefreshTime, INVALID_TIME)
		defer os.Unsetenv(EnvTopicRefreshTime)

		result := GetRefreshTime()

		if expected, _ := time.ParseDuration("30s"); result != expected {
			t.Errorf("Expected %s recieved %s", expected, result)
		}
	})

	t.Run("Read Default", func(t *testing.T) {
		const DEFAULT_TIME = "30s"
		os.Unsetenv(EnvTopicRefreshTime)

		result := GetRefreshTime()

		if expected, _ := time.ParseDuration(DEFAULT_TIME); result != expected {
			t.Errorf("Expected %s recieved %s", expected, result)
		}
	})
}

func TestGenerateRabbitMQUrl(t *testing.T) {
	t.Run("Read Informations", func(t *testing.T) {
		const RMQ_USER = "user"
		const RMQ_PASS = "pass"
		const RMQ_HOST = "rabbitmq-master"
		const RMQ_PORT = "1337"

		os.Setenv(EnvMQUser, RMQ_USER)
		os.Setenv(EnvMQPass, RMQ_PASS)
		os.Setenv(EnvMQHost, RMQ_HOST)
		os.Setenv(EnvMQPort, RMQ_PORT)

		defer os.Unsetenv(EnvMQUser)
		defer os.Unsetenv(EnvMQPass)
		defer os.Unsetenv(EnvMQHost)
		defer os.Unsetenv(EnvMQPort)

		result := GenerateRabbitMQUrl()
		expected := fmt.Sprintf("amqp://%s:%s@%s:%s/", RMQ_USER, RMQ_PASS, RMQ_HOST, RMQ_PORT)

		if result != expected {
			t.Errorf("Expected %s recieved %s", expected, result)
		}
	})

	t.Run("Read Defaults", func(t *testing.T) {
		const DEFAULT_USER = "guest"
		const DEFAULT_PASS = "guest"
		const DEFAULT_HOST = "localhost"
		const DEFAULT_PORT = "5672"

		os.Unsetenv(EnvMQUser)
		os.Unsetenv(EnvMQPass)
		os.Unsetenv(EnvMQHost)
		os.Unsetenv(EnvMQPort)

		result := GenerateRabbitMQUrl()
		expected := fmt.Sprintf("amqp://%s:%s@%s:%s/", DEFAULT_USER, DEFAULT_PASS, DEFAULT_HOST, DEFAULT_PORT)

		if result != expected {
			t.Errorf("Expected %s recieved %s", expected, result)
		}
	})
}

func TestGetTopics(t *testing.T) {
	t.Run("Read Valid Topics", func(t *testing.T) {
		const TOPICS = "awesome,sales,support"
		os.Setenv(EnvMQTopics, TOPICS)
		defer os.Unsetenv(EnvMQTopics)

		result := GetTopics()

		if strings.Join(result, ",") != TOPICS {
			t.Errorf("Expected %s recieved %s", TOPICS, result)
		}
	})

	t.Run("Empty Topic", func(t *testing.T) {
		os.Setenv(EnvMQTopics, "")
		defer os.Unsetenv(EnvMQTopics)
		defer func() {
			if r := recover(); r == nil {
				t.Error("Did not panic")
			}
		}()

		GetTopics()
	})

	t.Run("No Topic", func(t *testing.T) {
		os.Unsetenv(EnvMQTopics)
		defer func() {
			if r := recover(); r == nil {
				t.Error("Did not panic")
			}
		}()

		GetTopics()
	})
}
