package subscriber

import (
	"bytes"
	"errors"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"sync"
	"testing"
	"time"
)

//---- QueueConsumer Mock ----//

type fullQueueConsumer struct {
	faulty bool
	Output chan *types.OpenFaaSInvocation
}

func (m *fullQueueConsumer) Consume() (<-chan *types.OpenFaaSInvocation, error)  {
	if m.faulty {
		return nil, errors.New("expected")
	} else {
		m.Output = make(chan *types.OpenFaaSInvocation)
		return m.Output, nil
	}
}

func (m *fullQueueConsumer) Stop() {
	if m.Output != nil{
		close(m.Output)
	}
}

func (m *fullQueueConsumer) ListenForErrors() <-chan error {
	return nil
}

//---- Invoker Mock ----//
type mockInvoker struct {
	receivedTopic string
	receivedMessage *[]byte
}

func (m *mockInvoker) Invoke(topic string, message *[]byte)  {
	m.receivedTopic = topic
	m.receivedMessage = message
}


func TestSubscriber_Start(t *testing.T) {
	t.Run("Start without error", func(t *testing.T) {
		mock := fullQueueConsumer{faulty:false}
		target := NewSubscriber("Unit Test", "Sample", &mock, nil)

		err := target.Start()

		if err != nil {
			t.Errorf("Expected no error, but recieved %s", err)
		}

		if target.IsRunning() != true {
			t.Errorf("Expected true for isRunning recieved %t", target.IsRunning())
		}
	})

	t.Run("Start with error", func(t *testing.T) {
		mock := fullQueueConsumer{faulty: true}
		target := NewSubscriber("Unit Test", "Sample", &mock, nil)

		err := target.Start()

		if err.Error() != "expected" {
			t.Errorf("Expected Error expected recieved %s", err)
		}

		if target.IsRunning() != false {
			t.Errorf("Expected false for isRunning recieved %t", target.IsRunning())
		}
	})
}

func TestSubscriber_IsRunning(t *testing.T) {
	t.Run("When running", func(t *testing.T) {
		mock := fullQueueConsumer{faulty:false}
		target := NewSubscriber("Unit Test", "Sample", &mock, nil)
		_ = target.Start()

		isRunning := target.IsRunning()

		if isRunning != true {
			t.Errorf("Expected true for isRunning recieved %t", isRunning)
		}
	})

	t.Run("When not running", func(t *testing.T) {
		mock := fullQueueConsumer{faulty:false}
		target := NewSubscriber("Unit Test", "Sample", &mock, nil)

		isRunning := target.IsRunning()

		if isRunning != false {
			t.Errorf("Expected false for isRunning recieved %t", isRunning)
		}
	})
}

func TestSubscriber_Stop(t *testing.T) {
	t.Run("End without error", func(t *testing.T) {
		mock := fullQueueConsumer{faulty:false}
		target := NewSubscriber("Unit Test", "Sample", &mock, nil)

		_ = target.Start()
		err := target.Stop()

		if err != nil {
			t.Errorf("Expected no error, but recieved %s", err)
		}
	})

	t.Run("End without being started", func(t *testing.T) {
		mock := fullQueueConsumer{faulty:false}
		target := NewSubscriber("Unit Test", "Sample", &mock, nil)

		err := target.Stop()

		if err != nil {
			t.Errorf("Expected no error, but recieved %s", err)
		}
	})
}

func TestMessageReceived(t *testing.T) {
	t.Parallel()

	t.Run("With correct topic", func(t *testing.T) {
		var wg sync.WaitGroup
		message := []byte("Hello World")
		consumer := fullQueueConsumer{faulty:false}
		invoker := mockInvoker{}
		target := NewSubscriber("Unit Test", "Sample", &consumer, &invoker)

		_ = target.Start()

		wg.Add(1)

		go func() {
			consumer.Output <- &types.OpenFaaSInvocation{
				Topic: "Sample",
				Message: &message,
			}
			time.Sleep(1 * time.Second)
			wg.Done()
		}()
		wg.Wait()

		if invoker.receivedTopic != "Sample" {
			t.Errorf("Invoker was not called with the correct Topic Sample. %s", invoker.receivedTopic)
		}

		if !bytes.Equal(*invoker.receivedMessage, message) {
			t.Error("Invoker was not called with the correct Message.")
		}
	})

	t.Run("Without correct topic", func(t *testing.T) {
		var wg sync.WaitGroup
		message := []byte("Hello World")
		consumer := fullQueueConsumer{faulty:false}
		invoker := mockInvoker{}
		target := NewSubscriber("Unit Test", "Sample", &consumer, &invoker)

		_ = target.Start()

		wg.Add(1)

		go func() {
			consumer.Output <- &types.OpenFaaSInvocation{
				Topic: "Other",
				Message: &message,
			}
			time.Sleep(1 * time.Second)
			wg.Done()
		}()
		wg.Wait()

		if invoker.receivedTopic == "Sample" {
			t.Error("Invoker should not be called for Other")
		}
	})
}