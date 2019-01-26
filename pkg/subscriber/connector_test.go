package subscriber

import (
	"errors"
	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
	"runtime"
	"testing"
)

//---- QueueConsumer Mock ----//

type minimalQueueConsumer struct {
	IsActive bool
	Output   chan *types.OpenFaaSInvocation
}

func (m *minimalQueueConsumer) Consume() (<-chan *types.OpenFaaSInvocation, error) {
	m.IsActive = true
	m.Output = make(chan *types.OpenFaaSInvocation)
	return m.Output, nil
}

func (m *minimalQueueConsumer) Stop() {
	m.IsActive = false
	close(m.Output)
}

func (m *minimalQueueConsumer) ListenForErrors() <-chan *amqp.Error {
	return make(chan *amqp.Error)
}

//---- Factory Mock ----//

type mockFactory struct {
	Created uint
	faulty  bool
}

func (m *mockFactory) Build(topic string) (rabbitmq.QueueConsumer, error) {
	if m.faulty {
		return nil, errors.New("expected error")
	} else {
		m.Created += 1
		return &minimalQueueConsumer{IsActive: false, Output: nil}, nil
	}
}

//---- Helper ----//
func getSubscribers(f Connector) {

}

func TestConnector_Start(t *testing.T) {
	cfg := config.Controller{Topics: []string{"Hello"}}

	t.Parallel()

	t.Run("Start with no errors", func(t *testing.T) {
		factory := mockFactory{Created: 0, faulty: false}
		target := NewConnector(&cfg, nil, &factory)
		target.Start()

		connector, _ := target.(*connector)

		if len(connector.subscribers) != CalculateWorkerCount(1) {
			t.Errorf("Expected Connector to start %d Worker instead he started %d", CalculateWorkerCount(1), len(connector.subscribers))
		}

		for _, sub := range connector.subscribers {
			subscriber, _ := sub.(*subscriber)

			if !subscriber.IsRunning() {
				t.Error("Subscriber was not running")
			}
		}
	})

	t.Run("Start with errors", func(t *testing.T) {
		factory := mockFactory{Created: 0, faulty: true}
		target := NewConnector(&cfg, nil, &factory)
		target.Start()

		connector, _ := target.(*connector)

		if len(connector.subscribers) != 0 {
			t.Errorf("Expected Connector to start 0 Worker instead he started %d", len(connector.subscribers))
		}
	})
}

func TestConnector_End(t *testing.T) {
	cfg := config.Controller{Topics: []string{"Hello"}}

	factory := mockFactory{Created: 0, faulty: false}
	target := NewConnector(&cfg, nil, &factory)
	target.Start()
	target.End()

	connector, _ := target.(*connector)

	if len(connector.subscribers) != 0 {
		t.Errorf("Expected Connector to cleanup workers. %d are still left", len(connector.subscribers))
	}
}

func TestCalculateWorkerCount(t *testing.T) {
	t.Parallel()

	t.Run("All Workers on one topic", func(t *testing.T) {
		target := runtime.NumCPU() * 2

		calculated := CalculateWorkerCount(1)

		if calculated != target {
			t.Errorf("Expected %d Received %d", target, calculated)
		}
	})

	t.Run("Should Split between two topics", func(t *testing.T) {
		target := runtime.NumCPU()

		calculated := CalculateWorkerCount(2)

		if calculated != target {
			t.Errorf("Expected %d Received %d", target, calculated)
		}
	})

	t.Run("Exactly one worker per topic", func(t *testing.T) {
		target := 1

		calculated := CalculateWorkerCount(runtime.NumCPU() * 2)

		if calculated != target {
			t.Errorf("Expected %d Received %d", target, calculated)
		}
	})

	t.Run("At least one worker", func(t *testing.T) {
		target := 1

		calculated := CalculateWorkerCount(runtime.NumCPU()*2 + 2)

		if calculated != target {
			t.Errorf("Expected %d Received %d", target, calculated)
		}
	})
}
