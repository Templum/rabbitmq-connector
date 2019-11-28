package openfaas

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/openfaas/faas-provider/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTopicMap struct {
	mock.Mock
	lock         sync.RWMutex
	refreshCalls int
}

func (s *MockTopicMap) CalledNTimes() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.refreshCalls
}

func (s *MockTopicMap) GetCachedValues(name string) []string {
	args := s.Called(name)
	return args.Get(0).([]string)
}

func (s *MockTopicMap) Refresh(update map[string][]string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.refreshCalls++
}

type MockOpenFaaSClient struct {
	mock.Mock
	lock       sync.RWMutex
	invocation int
}

func (m *MockOpenFaaSClient) InvokeCalledNTimes() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.invocation
}

func (m *MockOpenFaaSClient) InvokeAsync(ctx context.Context, name string, payload []byte) (bool, error) {
	return true, nil
}

func (m *MockOpenFaaSClient) InvokeSync(ctx context.Context, name string, payload []byte) ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.invocation++
	return nil, nil
}

func (m *MockOpenFaaSClient) HasNamespaceSupport(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockOpenFaaSClient) GetNamespaces(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockOpenFaaSClient) GetFunctions(ctx context.Context, namespace string) ([]types.FunctionStatus, error) {
	args := m.Called(namespace)
	return args.Get(0).([]types.FunctionStatus), args.Error(1)
}

func TestCacher_Start_WithNs(t *testing.T) {
	namespaces := []string{
		"faas",
		"special",
		"test",
	}

	annotations := map[string]string{"topic": "billing,secret,transport"}

	fnFaaSNs := []types.FunctionStatus{
		types.FunctionStatus{
			Name:              "biller",
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
			Name:              "secrter",
			Image:             "docker:image",
			InvocationCount:   0,
			Replicas:          1,
			EnvProcess:        "",
			AvailableReplicas: 1,
			Labels:            nil,
			Annotations:       &annotations,
			Namespace:         "faas",
		},
	}

	fnTestNs := []types.FunctionStatus{
		types.FunctionStatus{
			Name:              "transporter",
			Image:             "docker:image",
			InvocationCount:   0,
			Replicas:          1,
			EnvProcess:        "",
			AvailableReplicas: 1,
			Labels:            nil,
			Annotations:       &annotations,
			Namespace:         "test",
		},
	}

	clientMock := new(MockOpenFaaSClient)
	clientMock.On("HasNamespaceSupport", mock.Anything).Return(true, nil)
	clientMock.On("GetNamespaces", mock.Anything).Return(namespaces, nil)
	clientMock.On("GetFunctions", "faas").Return(fnFaaSNs, nil)
	clientMock.On("GetFunctions", "test").Return(fnTestNs, nil)
	clientMock.On("GetFunctions", "special").Return([]types.FunctionStatus{}, nil)

	conf := &config.Controller{TopicRefreshTime: 3 * time.Second}

	t.Parallel()

	t.Run("Should perform a initial population of the map", func(t *testing.T) {
		cacheMock := new(MockTopicMap)

		cacher := NewController(conf, clientMock, cacheMock)

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an inital sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacheMock := new(MockTopicMap)

		cacher := NewController(conf, clientMock, cacheMock)

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an inital sync")
		time.Sleep(4 * time.Second)
		assert.Equal(t, cacheMock.CalledNTimes(), 2, "Expected a new sync")
	})
}

func TestCacher_Start_Normal(t *testing.T) {
	annotations := map[string]string{"topic": "billing,secret,transport"}

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
			Namespace:         "faas",
		},
	}

	clientMock := new(MockOpenFaaSClient)
	clientMock.On("HasNamespaceSupport", mock.Anything).Return(false, nil)
	clientMock.On("GetFunctions", mock.Anything).Return(functions, nil)

	conf := &config.Controller{TopicRefreshTime: 3 * time.Second}

	t.Parallel()

	t.Run("Should perform a initial population of the map", func(t *testing.T) {
		cacheMock := new(MockTopicMap)

		cacher := NewController(conf, clientMock, cacheMock)

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an inital sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacheMock := new(MockTopicMap)

		cacher := NewController(conf, clientMock, cacheMock)
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an inital sync")
		time.Sleep(4 * time.Second)
		assert.Equal(t, cacheMock.CalledNTimes(), 2, "Expected a new sync")
	})
}

func TestCacher_Start_WithFailures(t *testing.T) {
	conf := &config.Controller{TopicRefreshTime: 3 * time.Second}

	t.Parallel()

	t.Run("Should swallow errors received during get namespace", func(t *testing.T) {
		clientMock := new(MockOpenFaaSClient)
		clientMock.On("HasNamespaceSupport", mock.Anything).Return(true, nil)
		clientMock.On("GetNamespaces", mock.Anything).Return([]string{}, errors.New("Swallow me"))
		cacheMock := new(MockTopicMap)

		cacher := NewController(conf, clientMock, cacheMock)

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an inital sync")
	})

	t.Run("Should swallow errors received during get functions", func(t *testing.T) {
		clientMock := new(MockOpenFaaSClient)
		clientMock.On("HasNamespaceSupport", mock.Anything).Return(false, nil)
		clientMock.On("GetFunctions", mock.Anything).Return([]types.FunctionStatus{}, errors.New("Swallow me"))
		cacheMock := new(MockTopicMap)

		cacher := NewController(conf, clientMock, cacheMock)

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an inital sync")
	})
}

func TestCacher_Invoke(t *testing.T) {
	cacheMock := new(MockTopicMap)
	cacheMock.On("GetCachedValues", "Security").Return([]string{})
	cacheMock.On("GetCachedValues", "Billing").Return([]string{"billing", "secret", "transport"})

	t.Run("Should invoke all functions for specified Topic", func(t *testing.T) {
		clientMock := new(MockOpenFaaSClient)
		clientMock.On("HasNamespaceSupport", mock.Anything).Return(false, nil)

		cacher := NewController(nil, clientMock, cacheMock)

		cacher.Invoke("Billing", nil)
		time.Sleep(2 * time.Second) // Dirty trick as we don't really wait for go routines
		assert.Equal(t, clientMock.InvokeCalledNTimes(), 3)
	})

	t.Run("Should not invoke if there is no function for specified Topic", func(t *testing.T) {
		clientMock := new(MockOpenFaaSClient)
		clientMock.On("HasNamespaceSupport", mock.Anything).Return(false, nil)

		cacher := NewController(nil, clientMock, cacheMock)

		cacher.Invoke("Security", nil)
		assert.Equal(t, clientMock.InvokeCalledNTimes(), 0)
	})
}
