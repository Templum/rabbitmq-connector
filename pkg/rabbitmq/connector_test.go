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
	Called bool
}

func (m *MockRecoverer) recover(receivedError *amqp.Error){
	m.Called = true
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
	mock := MockRecoverer{Called:false}
	mockStream :=  make(chan *amqp.Error)

	go Healer(&mock, mockStream)

	// Simulate Error
	mockStream <- &amqp.Error{Recover:false, Reason: "Expected Error", Server: false}
	close(mockStream)

	time.Sleep(31 * time.Second)

	if !mock.Called {
		t.Errorf("Healer should call recover(), but it did not after %d Seconds", 31)
	}
}