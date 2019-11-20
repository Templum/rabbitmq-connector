package openfaas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/openfaas/faas-provider/types"
)

type MockTopicMap struct {
	lock        sync.RWMutex
	original    TopicMap
	refreshCall int
}

func (s *MockTopicMap) CalledNTimes() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.refreshCall
}

func (s *MockTopicMap) GetCachedValues(name string) []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.original.GetCachedValues(name)
}

func (s *MockTopicMap) Refresh(update map[string][]string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.refreshCall++
	s.original.Refresh(update)
}

func newTopicMapMock(original TopicMap) *MockTopicMap {
	return &MockTopicMap{
		lock:        sync.RWMutex{},
		original:    original,
		refreshCall: 0,
	}
}

func populateWithFunctionsForBillingTopic(target TopicMap) {
	cachedvalues := make(map[string][]string)
	cachedvalues["Billing"] = []string{"SendBill", "Transport", "Notifier"}
	target.Refresh(cachedvalues)
}

type MockOFClient struct {
	lock       sync.RWMutex
	invocation int
}

func (m *MockOFClient) InvokeCalledNTimes() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.invocation
}

func (m *MockOFClient) InvokeAsync(ctx context.Context, name string, payload []byte) (bool, error) {
	return true, nil
}

func (m *MockOFClient) InvokeSync(ctx context.Context, name string, payload []byte) ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.invocation++
	return nil, nil
}

func (m *MockOFClient) HasNamespaceSupport(ctx context.Context) (bool, error) {
	return true, nil
}

func (m *MockOFClient) GetNamespaces(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (m *MockOFClient) GetFunctions(ctx context.Context, namespace string) ([]types.FunctionStatus, error) {
	return []types.FunctionStatus{}, nil
}

func newOFClientMock() *MockOFClient {
	return &MockOFClient{
		lock:       sync.RWMutex{},
		invocation: 0,
	}
}

func TestCacher_Start_WithNs(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/system/namespaces" {
			namespaces := []string{
				"faas",
				"special",
				"namespace",
			}

			w.WriteHeader(200)
			out, _ := json.Marshal(namespaces)
			w.Write(out)
			return
		}
		if r.URL.Path == "/system/functions" {
			annotations := map[string]string{}
			annotations["topic"] = "billing,secret,transport"
			namespace := r.URL.Query().Get("namespace")

			if len(namespace) > 0 {
				functions := []types.FunctionStatus{
					types.FunctionStatus{
						Name:              "wrencher",
						Image:             "docker:image",
						InvocationCount:   0,
						Replicas:          1,
						EnvProcess:        "",
						AvailableReplicas: 1,
						Labels:            nil,
						Annotations:       &annotations,
						Namespace:         namespace,
					},
				}

				w.WriteHeader(200)
				out, _ := json.Marshal(functions)
				w.Write(out)
				return
			}
			functions := []types.FunctionStatus{
				types.FunctionStatus{
					Name:              "function-name",
					Image:             "docker:image",
					InvocationCount:   0,
					Replicas:          1,
					EnvProcess:        "",
					AvailableReplicas: 1,
					Labels:            nil,
					Annotations:       &annotations,
					Namespace:         "faas",
				},
				types.FunctionStatus{
					Name:              "wrencher",
					Image:             "docker:image",
					InvocationCount:   0,
					Replicas:          1,
					EnvProcess:        "",
					AvailableReplicas: 1,
					Labels:            nil,
					Annotations:       &annotations,
					Namespace:         "special",
				},
			}
			w.WriteHeader(200)
			out, _ := json.Marshal(functions)
			w.Write(out)
			return
		}
	}))
	defer server.Close()

	conf := &config.Controller{TopicRefreshTime: 3 * time.Second}
	client := NewClient(server.Client(), nil, server.URL)

	t.Parallel()

	t.Run("Should perform a initial population of the map", func(t *testing.T) {
		cacher := NewController(conf, client)
		mock := newTopicMapMock(cacher.cache)
		cacher.cache = mock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, mock.CalledNTimes(), 1, "Expected an inital sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacher := NewController(conf, client)
		mock := newTopicMapMock(cacher.cache)
		cacher.cache = mock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, mock.CalledNTimes(), 1, "Expected an inital sync")
		time.Sleep(4 * time.Second)
		assert.Equal(t, mock.CalledNTimes(), 2, "Expected a new sync")
	})
}

func TestCacher_Start_Normal(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/system/namespaces" {
			w.WriteHeader(502)
			w.Write(nil)
			return
		}
		if r.URL.Path == "/system/functions" {
			annotations := map[string]string{}
			annotations["topic"] = "billing,secret,transport"

			functions := []types.FunctionStatus{
				types.FunctionStatus{
					Name:              "function-name",
					Image:             "docker:image",
					InvocationCount:   0,
					Replicas:          1,
					EnvProcess:        "",
					AvailableReplicas: 1,
					Labels:            nil,
					Annotations:       &annotations,
					Namespace:         "faas",
				},
				types.FunctionStatus{
					Name:              "wrencher",
					Image:             "docker:image",
					InvocationCount:   0,
					Replicas:          1,
					EnvProcess:        "",
					AvailableReplicas: 1,
					Labels:            nil,
					Annotations:       &annotations,
					Namespace:         "special",
				},
			}
			w.WriteHeader(200)
			out, _ := json.Marshal(functions)
			w.Write(out)
			return
		}
	}))
	defer server.Close()

	conf := &config.Controller{TopicRefreshTime: 3 * time.Second}
	client := NewClient(server.Client(), nil, server.URL)

	t.Parallel()

	t.Run("Should perform a initial population of the map", func(t *testing.T) {
		cacher := NewController(conf, client)
		mock := newTopicMapMock(cacher.cache)
		cacher.cache = mock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, mock.CalledNTimes(), 1, "Expected an inital sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacher := NewController(conf, client)
		mock := newTopicMapMock(cacher.cache)
		cacher.cache = mock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, mock.CalledNTimes(), 1, "Expected an inital sync")
		time.Sleep(4 * time.Second)
		assert.Equal(t, mock.CalledNTimes(), 2, "Expected a new sync")
	})
}

func TestCacher_Invoke(t *testing.T) {

	t.Run("Should invoke all functions for specified Topic", func(t *testing.T) {
		mock := newOFClientMock()
		cacher := NewController(nil, mock)
		populateWithFunctionsForBillingTopic(cacher.cache)

		cacher.Invoke("Billing", nil)
		time.Sleep(2 * time.Second) // Dirty trick as we don't really wait for go routines
		assert.Equal(t, mock.InvokeCalledNTimes(), 3)
	})

	t.Run("Should not invoke if there is no function for specified Topic", func(t *testing.T) {
		mock := newOFClientMock()
		cacher := NewController(nil, mock)
		populateWithFunctionsForBillingTopic(cacher.cache)

		cacher.Invoke("Security", nil)
		assert.Equal(t, mock.InvokeCalledNTimes(), 0)
	})
}
