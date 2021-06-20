/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package openfaas

import (
	"context"
	"sync"
	"testing"
	"time"

	types2 "github.com/Templum/rabbitmq-connector/pkg/types"

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
}

func (m *MockOpenFaaSClient) InvokeAsync(ctx context.Context, name string, invocation *types2.OpenFaaSInvocation) (bool, error) {
	args := m.Called(ctx, name, invocation)
	return args.Bool(0), args.Error(1)
}

func (m *MockOpenFaaSClient) InvokeSync(ctx context.Context, name string, invocation *types2.OpenFaaSInvocation) ([]byte, error) {
	args := m.Called(ctx, name, invocation)
	return args.Get(0).([]byte), args.Error(1)
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
		{
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
		{
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
		{
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
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an initial sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacheMock := new(MockTopicMap)

		cacher := NewController(conf, clientMock, cacheMock)

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an initial sync")
		time.Sleep(4 * time.Second)
		assert.Equal(t, cacheMock.CalledNTimes(), 2, "Expected a new sync")
	})
}

func TestCacher_Start_Normal(t *testing.T) {
	annotations := map[string]string{"topic": "billing,secret,transport"}

	functions := []types.FunctionStatus{
		{
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
		{
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
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an initial sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacheMock := new(MockTopicMap)

		cacher := NewController(conf, clientMock, cacheMock)
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an initial sync")
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
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an initial sync")
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
		assert.Equal(t, cacheMock.CalledNTimes(), 1, "Expected an initial sync")
	})
}

func TestCacher_Invoke(t *testing.T) {
	cacheMock := new(MockTopicMap)
	cacheMock.On("GetCachedValues", "Security").Return([]string{})
	cacheMock.On("GetCachedValues", "Billing").Return([]string{"billing", "secret", "transport"})

	const TOPIC = "Billing"

	t.Run("Should invoke all functions for specified Topic", func(t *testing.T) {
		clientMock := new(MockOpenFaaSClient)
		clientMock.On("InvokeSync", mock.Anything, mock.Anything, mock.Anything).Return([]byte{}, nil)

		cacher := NewController(nil, clientMock, cacheMock)

		err := cacher.Invoke(TOPIC, nil)

		assert.NoError(t, err, "should not throw")
		clientMock.AssertNumberOfCalls(t, "InvokeSync", 3)
		clientMock.AssertExpectations(t)
	})

	t.Run("Should abort invocation of functions on receiving first error further returning it", func(t *testing.T) {
		clientMock := new(MockOpenFaaSClient)
		clientMock.On("InvokeSync", mock.Anything, mock.Anything, mock.Anything).Return([]byte{}, errors.New("failed"))

		cacher := NewController(nil, clientMock, cacheMock)

		err := cacher.Invoke(TOPIC, nil)

		assert.Error(t, err, "failed")
		clientMock.AssertNumberOfCalls(t, "InvokeSync", 1)
		clientMock.AssertExpectations(t)
	})

	t.Run("Should not invoke if there is no function for specified Topic", func(t *testing.T) {
		clientMock := new(MockOpenFaaSClient)
		clientMock.On("InvokeSync", mock.Anything, mock.Anything, mock.Anything).Return([]byte{}, nil)

		cacher := NewController(nil, clientMock, cacheMock)

		err := cacher.Invoke("Security", nil)

		assert.NoError(t, err, "should not throw")
		clientMock.AssertNotCalled(t, "InvokeSync")
	})
}
