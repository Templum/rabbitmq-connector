package openfaas

import (
	"context"
	"encoding/json"
	"gotest.tools/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/openfaas/faas-provider/types"
)

type SpyTopicMap struct {
	original    TopicMap
	refreshCall int
}

func (s *SpyTopicMap) GetCachedValues(name string) []string {
	return s.original.GetCachedValues(name)
}

func (s *SpyTopicMap) Refresh(update map[string][]string) {
	s.refreshCall++
	s.original.Refresh(update)
}

func newSpy(original TopicMap) *SpyTopicMap {
	return &SpyTopicMap{
		original:    original,
		refreshCall: 0,
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
		spy := newSpy(cacher.cache)
		cacher.cache = spy
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, spy.refreshCall, 1, "Expected an inital sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacher := NewController(conf, client)
		spy := newSpy(cacher.cache)
		cacher.cache = spy
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, spy.refreshCall, 1, "Expected an inital sync")
		time.Sleep(4 * time.Second)
		assert.Equal(t, spy.refreshCall, 2, "Expected a new sync")
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
		spy := newSpy(cacher.cache)
		cacher.cache = spy
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, spy.refreshCall, 1, "Expected an inital sync")
	})

	t.Run("Should sync every 3 seconds", func(t *testing.T) {
		cacher := NewController(conf, client)
		spy := newSpy(cacher.cache)
		cacher.cache = spy
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cacher.Start(ctx)
		assert.Equal(t, spy.refreshCall, 1, "Expected an inital sync")
		time.Sleep(4 * time.Second)
		assert.Equal(t, spy.refreshCall, 2, "Expected a new sync")
	})
}
