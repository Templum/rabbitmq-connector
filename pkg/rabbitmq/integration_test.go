package rabbitmq

import (
	"flag"
	"os"
	"testing"
	"time"
)

var (
	integration = flag.Bool("integration", false, "run integration test")
)

const EnvMQTopics = "RMQ_TOPICS"
const TOPICS = "account,support"
const TEST_URI = "amqp://user:pass@localhost:5672/"

// Mocks
type MockOpenFaaS struct {
	finishChannel chan bool
}

func (m *MockOpenFaaS) Invoke(topic string, message *[]byte) {

}

func CreateMock() Invoker {
	return &MockOpenFaaS{
		nil,
	}
}

func TestMakeConnector(t *testing.T) {
	flag.Parse()
	if !*integration {
		t.Skip()
	}

	setupEnvironment(t)

	t.Run("Failing Start", func(t *testing.T) {
		target := MakeConnector("amqp://not:correct@localhost:5672/", CreateMock())

		defer func() {
			if r := recover(); r == nil {
				t.Error("Did not panic")
			}
			teardownTarget(t, target)
		}()
		target.Start()
	})

	t.Run("Successfull Start", func(t *testing.T) {
		target := MakeConnector(TEST_URI, CreateMock())
		target.Start()
		time.Sleep(5 * time.Second)
		teardownTarget(t, target)
	})

	t.Run("Quick Stop", func(t *testing.T) {
		target := MakeConnector(TEST_URI, CreateMock())
		target.Close()
		time.Sleep(5 * time.Second)
		teardownTarget(t, target)
	})

	t.Run("Full Stop", func(t *testing.T) {
		target := MakeConnector(TEST_URI, CreateMock())
		target.Start()
		time.Sleep(5 * time.Second)
		target.Close()
		time.Sleep(5 * time.Second)
		teardownTarget(t, target)
	})

	teardown(t)
}

func teardownTarget(t *testing.T, target Connector) {
	t.Helper()
	target.Close()
}

func setupEnvironment(t *testing.T) {
	t.Helper()
	os.Setenv(EnvMQTopics, TOPICS)
}

func teardown(t *testing.T) {
	t.Helper()
	os.Unsetenv(EnvMQTopics)
}
