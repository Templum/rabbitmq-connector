package rabbitmq

import (
	"github.com/streadway/amqp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Mocks
type MockRecoverer struct {
	finishChannel  chan bool
}

func CreateRecoverer(finished chan bool) *MockRecoverer{
	return & MockRecoverer{finished}
}

func (m *MockRecoverer) recover(receivedError *amqp.Error){
	m.finishChannel <- true
	close(m.finishChannel)
}


// Tests

func TestSanitizeConnectionUri(t *testing.T) {
	t.Run("Sanitize Credentials", func(t *testing.T) {
		const CONTAINS_CREDS = "amqp://user:pass@localhost:5672/"
		sanitized := SanitizeConnectionUri(CONTAINS_CREDS)

		if idx := strings.Index(sanitized, "user:pass"); idx != -1 {
			t.Errorf("Expected %s recieved %s", "localhost:5672/", sanitized)
		}
	})

	t.Run("Not Modify", func(t *testing.T) {
		const CONTAINS_NO_CREDS = "amqp://localhost:5672"
		sanitized := SanitizeConnectionUri(CONTAINS_NO_CREDS)

		if sanitized != CONTAINS_NO_CREDS {
			t.Errorf("Expected %s recieved %s", CONTAINS_NO_CREDS, sanitized)
		}
	})
}

func TestCalculateWorkerCount(t *testing.T) {
	t.Run("All Workers on one topic", func(t *testing.T) {
		target := runtime.NumCPU() * 2

		calculated := CalculateWorkerCount(1)

		if calculated != target {
			t.Errorf("Expected %d recieved %d", target, calculated)
		}
	})

	t.Run("Should Split between two topics", func(t *testing.T) {
		target := runtime.NumCPU()

		calculated := CalculateWorkerCount(2)

		if calculated != target {
			t.Errorf("Expected %d recieved %d", target, calculated)
		}
	})

	t.Run("Exactly one worker per topic", func(t *testing.T) {
		target := 1

		calculated := CalculateWorkerCount(runtime.NumCPU() * 2)

		if calculated != target {
			t.Errorf("Expected %d recieved %d", target, calculated)
		}
	})

	t.Run("At least one worker", func(t *testing.T) {
		target := 1

		calculated := CalculateWorkerCount(runtime.NumCPU() * 2 + 2)

		if calculated != target {
			t.Errorf("Expected %d recieved %d", target, calculated)
		}
	})
}

func TestHealer(t *testing.T) {
	mockStream :=  make(chan *amqp.Error)
	calledStream := make(chan bool)

	go Healer(CreateRecoverer(calledStream), mockStream)

	// Simulate Error
	mockStream <- &amqp.Error{Recover:false, Reason: "Expected Error", Server: false}
	close(mockStream)

	select {
	case called := <-calledStream:
		if called {}
	case <-time.After(31 * time.Second):
		t.Errorf("Healer did not call recover() after %d Seconds", 31)
	}
}