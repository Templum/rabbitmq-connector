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
	cachedvalues := map[string][]string{"Billing": []string{"SendBill", "Transport", "Notifier"}}
	target.Refresh(cachedvalues)
}

type MockOpenFaaSClient struct {
	mock.Mock
	lock       sync.RWMutex
	invocation int
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

type MockOFClient struct {
	lock       sync.RWMutex
	invocation int
	err        error
	namespace  []string
	functions  []types.FunctionStatus
	hasNS      bool
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
	return m.hasNS, m.err
}

func (m *MockOFClient) GetNamespaces(ctx context.Context) ([]string, error) {
	if len(m.namespace) > 0 {
		return m.namespace, nil
	}
	return m.namespace, m.err
}

func (m *MockOFClient) GetFunctions(ctx context.Context, namespace string) ([]types.FunctionStatus, error) {

	return m.functions, m.err
}

func newOFClientMock(namespace []string, functions []types.FunctionStatus, err error, hasNS bool) *MockOFClient {
	return &MockOFClient{
		lock:       sync.RWMutex{},
		invocation: 0,
		err:        err,
		hasNS:      hasNS,
		namespace:  namespace,
		functions:  functions,
	}
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
		cacher := NewController(conf, clientMock)
		topicMock := newTopicMapMock(cacher.cache)
		cacher.cache = topicMock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, topicMock.CalledNTimes(), 1, "Expected an inital sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacher := NewController(conf, clientMock)
		topicMock := newTopicMapMock(cacher.cache)
		cacher.cache = topicMock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, topicMock.CalledNTimes(), 1, "Expected an inital sync")
		time.Sleep(4 * time.Second)
		assert.Equal(t, topicMock.CalledNTimes(), 2, "Expected a new sync")
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
		cacher := NewController(conf, clientMock)
		topicMock := newTopicMapMock(cacher.cache)
		cacher.cache = topicMock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, topicMock.CalledNTimes(), 1, "Expected an inital sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacher := NewController(conf, clientMock)
		topicMock := newTopicMapMock(cacher.cache)
		cacher.cache = topicMock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, topicMock.CalledNTimes(), 1, "Expected an inital sync")
		time.Sleep(4 * time.Second)
		assert.Equal(t, topicMock.CalledNTimes(), 2, "Expected a new sync")
	})
}

func TestCacher_Start_WithFailures(t *testing.T) {
	conf := &config.Controller{TopicRefreshTime: 3 * time.Second}

	t.Parallel()

	t.Run("Should swallow errors received during get namespace", func(t *testing.T) {
		clientMock := new(MockOpenFaaSClient)
		clientMock.On("HasNamespaceSupport", mock.Anything).Return(true, nil)
		clientMock.On("GetNamespaces", mock.Anything).Return([]string{}, errors.New("Swallow me"))

		cacher := NewController(conf, clientMock)
		topicMock := newTopicMapMock(cacher.cache)
		cacher.cache = topicMock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, topicMock.CalledNTimes(), 1, "Expected an inital sync")
	})

	t.Run("Should swallow errors received during get functions", func(t *testing.T) {
		clientMock := new(MockOpenFaaSClient)
		clientMock.On("HasNamespaceSupport", mock.Anything).Return(false, nil)
		clientMock.On("GetFunctions", mock.Anything).Return([]types.FunctionStatus{}, errors.New("Swallow me"))

		cacher := NewController(conf, clientMock)
		topicMock := newTopicMapMock(cacher.cache)
		cacher.cache = topicMock
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, topicMock.CalledNTimes(), 1, "Expected an inital sync")
	})
}

func TestCacher_Invoke(t *testing.T) {

	t.Run("Should invoke all functions for specified Topic", func(t *testing.T) {
		mock := newOFClientMock(nil, nil, nil, false)
		cacher := NewController(nil, mock)
		populateWithFunctionsForBillingTopic(cacher.cache)

		cacher.Invoke("Billing", nil)
		time.Sleep(2 * time.Second) // Dirty trick as we don't really wait for go routines
		assert.Equal(t, mock.InvokeCalledNTimes(), 3)
	})

	t.Run("Should not invoke if there is no function for specified Topic", func(t *testing.T) {
		mock := newOFClientMock(nil, nil, nil, false)
		cacher := NewController(nil, mock)
		populateWithFunctionsForBillingTopic(cacher.cache)

		cacher.Invoke("Security", nil)
		assert.Equal(t, mock.InvokeCalledNTimes(), 0)
	})
}
