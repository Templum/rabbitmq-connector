// +build !race

package main

import (
	"fmt"
	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/subscriber"
	"github.com/streadway/amqp"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

var (
	connector      subscriber.Connector
	mockClient     *invokerMock
	producerClient *Producer
)

type Invocation struct {
	Topic      string
	Message    *[]byte
	ReceivedNo uint
}

func NewInvocation(topic string, message *[]byte, receivedNo uint) *Invocation {
	return &Invocation{Topic: topic, Message: message, ReceivedNo: receivedNo}
}

//---- Producer ----//
type Producer struct {
	channel *amqp.Channel
	counter uint

	exchange string
	topic    string

	mutex sync.Mutex
}

func NewProducer(cfg *config.Controller) (*Producer, error) {
	prod := &Producer{}
	err := prod.setup(cfg)
	return prod, err
}

func (p *Producer) setup(cfg *config.Controller) error {
	var err error
	con, err := amqp.Dial(cfg.RabbitConnectionURL)

	channel, err := con.Channel()

	err = channel.ExchangeDeclare(
		cfg.ExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}
	p.exchange = cfg.ExchangeName
	p.topic = cfg.Topics[0]
	p.channel = channel
	return nil
}

func (p *Producer) SendMessage(body *[]byte) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.channel.Publish(
		p.exchange,
		p.topic,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         *body,
			Timestamp:    time.Now(),
		})

	return err
}

//---- Invoker Mock ----//
type invokerMock struct {
	invocations []*Invocation
	counter     uint
	mutex       sync.Mutex
}

func newInvokerMock() *invokerMock {
	return &invokerMock{counter: 0}
}

func (m *invokerMock) Invoke(topic string, message *[]byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.counter++
	invocation := NewInvocation(topic, message, m.counter)
	m.invocations = append(m.invocations, invocation)
}

func (m *invokerMock) GetInvocations() []*Invocation {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.invocations
}

//---- Invoker Mock ----//

func TestMain(m *testing.M) {
	os.Setenv("RMQ_TOPICS", "unit_test,another_topic")
	os.Setenv("RMQ_EXCHANGE", "Ex")
	os.Setenv("RMQ_HOST", "localhost")
	os.Setenv("RMQ_PORT", "5672")
	os.Setenv("RMQ_USER", "user")
	os.Setenv("RMQ_PASS", "pass")

	defer os.Unsetenv("RMQ_TOPICS")
	defer os.Unsetenv("RMQ_EXCHANGE")
	defer os.Unsetenv("RMQ_HOST")
	defer os.Unsetenv("RMQ_PORT")
	defer os.Unsetenv("RMQ_USER")
	defer os.Unsetenv("RMQ_PASS")

	connectorCfg, err := config.NewConfig()
	if err != nil {
		log.Printf("Recieved %s during config building", err)
		os.Exit(1)
	}
	factory, err := rabbitmq.NewQueueConsumerFactory(connectorCfg)
	if err != nil {
		log.Printf("Recieved %s during factory creation", err)
		os.Exit(1)
	}
	producerClient, err = NewProducer(connectorCfg)
	if err != nil {
		log.Printf("Recieved %s during producer creation", err)
		os.Exit(1)
	}
	mockClient = newInvokerMock()
	connector = subscriber.NewConnector(connectorCfg, mockClient, factory)

	os.Exit(m.Run())
}

func TestSystem(t *testing.T) {
	connector.Start()
	time.Sleep(10 * time.Second)

	t.Run("Should receive 500 messages without losses", func(t *testing.T) {
		for i := 0; i < 500; i++ {
			message := []byte(fmt.Sprintf("I'm Message %d", i))
			err := producerClient.SendMessage(&message)

			if err != nil {
				log.Printf("Recieved error %s for message %d", err, i)
				i--
			}
		}

		time.Sleep(10 * time.Second)
		receivedInvocations := mockClient.GetInvocations()

		if len(receivedInvocations) != 500 {
			t.Errorf("Expected to receive 500 Messages, only received %d", len(receivedInvocations))
		}

		mockClient.invocations = nil
	})

	connector.End()
	time.Sleep(10 * time.Second)

	t.Run("Should process messages send while being inactive", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			message := []byte(fmt.Sprintf("I'm Message %d send while beeing inactive", i))
			err := producerClient.SendMessage(&message)

			if err != nil {
				log.Printf("Recieved error %s for message %d", err, i)
				i--
			}
		}

		connector.Start()
		time.Sleep(10 * time.Second)

		receivedInvocations := mockClient.GetInvocations()
		if len(receivedInvocations) != 100 {
			t.Errorf("Expected to receive 500 Messages, only received %d", len(receivedInvocations))
		}

	})
}
