package subscriber

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//---- QueueConsumer Mock ----//

type fullQueueConsumer struct {
	mock.Mock
}

func (m *fullQueueConsumer) Consume() (<-chan *types.OpenFaaSInvocation, error) {
	args := m.Called(nil)

	if c, ok := args.Get(0).(chan *types.OpenFaaSInvocation); ok {
		return c, nil
	}
	return nil, errors.New("expected")
}

func (m *fullQueueConsumer) Stop() {
	m.Called(nil)
}

func (m *fullQueueConsumer) ListenForErrors() <-chan *amqp.Error {

	args := m.Called(nil)
	return args.Get(0).(chan *amqp.Error)
}

//---- Invoker Mock ----//
type mockInvoker struct {
	mock.Mock
	mutex sync.RWMutex
}

func (m *mockInvoker) Invoke(topic string, message []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Called(topic, message)
}

func TestSubscriber_Start(t *testing.T) {
	t.Parallel()

	t.Run("Start without error", func(t *testing.T) {
		mock := new(fullQueueConsumer)
		mock.On("Consume", nil).Return(make(chan *types.OpenFaaSInvocation))
		mock.On("ListenForErrors", nil).Return(make(chan *amqp.Error))
		target := NewSubscriber("Unit Test", "Sample", mock, nil)

		err := target.Start()
		assert.Nil(t, err, "Expected start not to fail")
		assert.True(t, target.IsRunning(), "Expected consumer to be running")
	})

	t.Run("Start with error", func(t *testing.T) {
		mock := new(fullQueueConsumer)
		mock.On("Consume", nil).Return(nil)
		target := NewSubscriber("Unit Test", "Sample", mock, nil)

		err := target.Start()
		assert.Error(t, err, "Expected start to fail")
		assert.False(t, target.IsRunning(), "Expected consumer not to be running")
	})
}

func TestSubscriber_IsRunning(t *testing.T) {
	t.Parallel()

	t.Run("When running", func(t *testing.T) {
		mock := new(fullQueueConsumer)
		mock.On("Consume", nil).Return(make(chan *types.OpenFaaSInvocation))
		mock.On("ListenForErrors", nil).Return(make(chan *amqp.Error))
		target := NewSubscriber("Unit Test", "Sample", mock, nil)
		_ = target.Start()

		isRunning := target.IsRunning()
		assert.True(t, isRunning, "Expected consumer to be running")
	})

	t.Run("When not running", func(t *testing.T) {
		mock := new(fullQueueConsumer)
		target := NewSubscriber("Unit Test", "Sample", mock, nil)

		isRunning := target.IsRunning()
		assert.False(t, isRunning, "Expected consumer not to be running")
	})
}

func TestSubscriber_Stop(t *testing.T) {
	t.Parallel()

	t.Run("End without error", func(t *testing.T) {
		mock := new(fullQueueConsumer)
		mock.On("Consume", nil).Return(make(chan *types.OpenFaaSInvocation))
		mock.On("ListenForErrors", nil).Return(make(chan *amqp.Error))
		mock.On("Stop", nil).Return()
		target := NewSubscriber("Unit Test", "Sample", mock, nil)

		_ = target.Start()
		err := target.Stop()
		assert.Nil(t, err, "Expected stop not to fail")
	})

	t.Run("End without being started", func(t *testing.T) {
		mock := new(fullQueueConsumer)
		mock.On("Consume", nil).Return(make(chan *types.OpenFaaSInvocation))
		mock.On("ListenForErrors", nil).Return(make(chan *amqp.Error))
		mock.On("Stop", nil).Return()
		target := NewSubscriber("Unit Test", "Sample", mock, nil)

		err := target.Stop()
		assert.Nil(t, err, "Expected stop not to fail")
	})
}

func TestConsumerError(t *testing.T) {
	t.Parallel()

	t.Run("Should stop consumer if critical error was received", func(t *testing.T) {
		errChannel := make(chan *amqp.Error)
		mock := new(fullQueueConsumer)
		mock.On("Consume", nil).Return(make(chan *types.OpenFaaSInvocation))
		mock.On("ListenForErrors", nil).Return(errChannel)
		mock.On("Stop", nil).Return()
		target := NewSubscriber("Unit Test", "Sample", mock, nil)

		err := target.Start()
		assert.Nil(t, err, "Expected start not to fail")
		assert.True(t, target.IsRunning(), "Expected consumer to be running")

		errChannel <- &amqp.Error{Recover: false, Reason: "Critical Error Occured"}
		time.Sleep(300 * time.Millisecond)

		mock.AssertNumberOfCalls(t, "Consume", 1)
		mock.AssertNumberOfCalls(t, "Stop", 1)
		assert.False(t, target.IsRunning(), "Expected consumer to be stopped")
	})

	t.Run("Should restart consumer if recoverable error was received", func(t *testing.T) {
		errChannel := make(chan *amqp.Error)
		mock := new(fullQueueConsumer)
		mock.On("Consume", nil).Return(make(chan *types.OpenFaaSInvocation))
		mock.On("ListenForErrors", nil).Return(errChannel)
		mock.On("Stop", nil).Return()
		target := NewSubscriber("Unit Test", "Sample", mock, nil)

		err := target.Start()
		assert.Nil(t, err, "Expected start not to fail")
		assert.True(t, target.IsRunning(), "Expected consumer to be running")

		errChannel <- &amqp.Error{Recover: true, Reason: "Recoverable by reconnecting"}
		time.Sleep(300 * time.Millisecond)

		mock.AssertNumberOfCalls(t, "Consume", 2)
		mock.AssertNumberOfCalls(t, "Stop", 1)
		assert.True(t, target.IsRunning(), "Expected consumer to be running")
	})
}

func TestMessageReceived(t *testing.T) {
	t.Parallel()

	t.Run("With correct topic", func(t *testing.T) {
		msgStream := make(chan *types.OpenFaaSInvocation)
		message := []byte("Hello World")
		targetTopic := "Sample"

		consumerMock := new(fullQueueConsumer)
		consumerMock.On("Consume", nil).Return(msgStream)
		consumerMock.On("ListenForErrors", nil).Return(make(chan *amqp.Error))
		consumerMock.On("Stop", nil).Return()

		invokerMock := new(mockInvoker)
		invokerMock.On("Invoke", targetTopic, message).Return()

		target := NewSubscriber("Unit Test", "Sample", consumerMock, invokerMock)

		_ = target.Start()

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			msgStream <- &types.OpenFaaSInvocation{
				Topic:   "Sample",
				Message: &message,
			}
			time.Sleep(300 * time.Millisecond)
			wg.Done()
		}()
		wg.Wait()

		invokerMock.AssertCalled(t, "Invoke", targetTopic, message)
		invokerMock.AssertNumberOfCalls(t, "Invoke", 1)
	})

	t.Run("Without correct topic", func(t *testing.T) {
		msgStream := make(chan *types.OpenFaaSInvocation)
		message := []byte("Hello World")
		targetTopic := "Other"

		consumerMock := new(fullQueueConsumer)
		consumerMock.On("Consume", nil).Return(msgStream)
		consumerMock.On("ListenForErrors", nil).Return(make(chan *amqp.Error))
		consumerMock.On("Stop", nil).Return()

		invokerMock := new(mockInvoker)
		invokerMock.On("Invoke", targetTopic, message).Return()

		target := NewSubscriber("Unit Test", "Sample", consumerMock, invokerMock)

		_ = target.Start()

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			msgStream <- &types.OpenFaaSInvocation{
				Topic:   targetTopic,
				Message: &message,
			}
			time.Sleep(500 * time.Millisecond)
			wg.Done()
		}()
		wg.Wait()

		invokerMock.AssertNotCalled(t, "Invoke", "Sample", message)
		invokerMock.AssertNumberOfCalls(t, "Invoke", 0)
	})
}
