/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package openfaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	types2 "github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/valyala/fasthttp"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
)

func CreateClient(server *httptest.Server) *fasthttp.Client {
	client := types2.MakeHTTPClient(true, 256, 30*time.Second)
	// TODO: For the future configure client with cert pool from the server
	return client
}

func TestClient_InvokeSync(t *testing.T) {
	expectedResponse := "Hello World"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, pass, ok := r.BasicAuth(); ok {
			if user == "User" && pass == "Pass" {
				w.WriteHeader(202)
				fmt.Fprint(w, "Hello World")
				return
			}
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return
		}

		if r.Method != fasthttp.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Method not supported")
			return
		}

		switch r.URL.Path {
		case "/function/exists":
			w.WriteHeader(200)
			fmt.Fprint(w, expectedResponse)
		case "/function/nonexisting":
			w.WriteHeader(404)
			fmt.Fprint(w, "Not Found")
		default:
			w.WriteHeader(500)
			fmt.Fprint(w, "Internal Server Error")
		}
	}))
	defer server.Close()

	openfaasClient := NewClient(CreateClient(server), nil, server.URL)

	authenticatedOpenFaaSClient := NewClient(CreateClient(server), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Invalid",
	}, server.URL)

	message := []byte("Test")
	payload := types2.OpenFaaSInvocation{
		Topic:           "",
		Message:         &message,
		ContentEncoding: "gzip",
		ContentType:     "text/plain",
	}
	nilPayload := types2.OpenFaaSInvocation{
		Topic:           "",
		Message:         nil,
		ContentEncoding: "gzip",
		ContentType:     "text/plain",
	}

	t.Parallel()

	t.Run("Should invoke the specified function", func(t *testing.T) {
		resp, err := openfaasClient.InvokeSync(context.Background(), "exists", &payload)

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, string(resp), expectedResponse, "Did not receive expected response")
	})

	t.Run("Should except nil as body", func(t *testing.T) {
		resp, err := openfaasClient.InvokeSync(context.Background(), "exists", &nilPayload)

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, string(resp), expectedResponse, "Did not receive expected response")
	})

	t.Run("Should throw error if function does not exist", func(t *testing.T) {
		_, err := openfaasClient.InvokeSync(context.Background(), "nonexisting", &payload)

		assert.Error(t, err, "Function nonexisting is not deployed", "Did receive unexpected error")
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.InvokeSync(context.Background(), "exists", &nilPayload)

		assert.Error(t, err, "OpenFaaS Credentials are invalid", "Did receive unexpected error")
	})

	t.Run("Should throw error on unexpected status code", func(t *testing.T) {
		_, err := openfaasClient.InvokeSync(context.Background(), "internal", &payload)

		assert.Error(t, err, "Received unexpected Status Code 500", "Did receive unexpected error")
	})
}

func TestClient_InvokeAsync(t *testing.T) {

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, pass, ok := r.BasicAuth(); ok {
			if user == "User" && pass == "Pass" {
				w.WriteHeader(202)
				fmt.Fprint(w, "Hello World")
				return
			}
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return
		}

		if r.Method != fasthttp.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Method not supported")
			return
		}

		switch r.URL.Path {
		case "/async-function/exists":
			w.WriteHeader(202)
			fmt.Fprint(w, "Hello World")
		case "/async-function/nonexisting":
			w.WriteHeader(404)
			fmt.Fprint(w, "Not Found")
		default:
			w.WriteHeader(500)
			fmt.Fprint(w, "Internal Server Error")
		}
	}))
	defer server.Close()

	openfaasClient := NewClient(CreateClient(server), nil, server.URL)

	authenticatedOpenFaaSClient := NewClient(CreateClient(server), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Invalid",
	}, server.URL)

	message := []byte("Test")
	payload := types2.OpenFaaSInvocation{
		Topic:           "",
		Message:         &message,
		ContentEncoding: "gzip",
		ContentType:     "text/plain",
	}
	nilPayload := types2.OpenFaaSInvocation{
		Topic:           "",
		Message:         nil,
		ContentEncoding: "gzip",
		ContentType:     "text/plain",
	}

	t.Parallel()

	t.Run("Should invoke the specified function", func(t *testing.T) {
		ok, err := openfaasClient.InvokeAsync(context.Background(), "exists", &payload)

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, ok, true, "Did not receive expected response")
	})

	t.Run("Should except nil as body", func(t *testing.T) {
		ok, err := openfaasClient.InvokeAsync(context.Background(), "exists", &nilPayload)

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, ok, true, "Did not receive expected response")
	})

	t.Run("Should throw error if function does not exist", func(t *testing.T) {
		_, err := openfaasClient.InvokeAsync(context.Background(), "nonexisting", &payload)

		assert.Error(t, err, "Function nonexisting is not deployed", "Did receive unexpected error")
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.InvokeAsync(context.Background(), "exists", &nilPayload)

		assert.Error(t, err, "OpenFaaS Credentials are invalid", "Did receive unexpected error")
	})

	t.Run("Should throw error on unexpected status code", func(t *testing.T) {
		_, err := openfaasClient.InvokeAsync(context.Background(), "internal", &payload)

		assert.Error(t, err, "Received unexpected Status Code 500", "Did receive unexpected error")
	})
}

func TestClient_HasNamespaceSupport(t *testing.T) {
	k8sOF := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, pass, ok := r.BasicAuth(); ok {
			if user == "User" && pass == "Pass" {
				w.WriteHeader(502)
				_, _ = w.Write(nil)
				return
			}
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return
		}

		if r.Method != fasthttp.MethodGet {
			w.WriteHeader(400)
			fmt.Fprint(w, "Method not supported")
			return
		}

		namespaces := []string{"one", "two"}
		out, _ := json.Marshal(namespaces)

		w.WriteHeader(200)
		_, _ = w.Write(out)
	}))
	defer k8sOF.Close()

	swarmOF := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		namespaces := []string{}
		out, _ := json.Marshal(namespaces)

		w.WriteHeader(200)
		_, _ = w.Write(out)
	}))

	ofK8SClient := NewClient(CreateClient(k8sOF), nil, k8sOF.URL)

	ofSwarmClient := NewClient(CreateClient(swarmOF), nil, swarmOF.URL)

	authenticatedOpenFaaSClient := NewClient(CreateClient(k8sOF), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Invalid",
	}, k8sOF.URL)

	failingOpenFaaSClient := NewClient(CreateClient(k8sOF), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Pass",
	}, k8sOF.URL)

	t.Parallel()

	t.Run("Should return true if namespaces endpoint available", func(t *testing.T) {
		ok, err := ofK8SClient.HasNamespaceSupport(context.Background())

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, ok, true, "Did not receive expected response")
	})

	t.Run("Should return false if namespace endpoint does not return empty list", func(t *testing.T) {
		ok, err := ofSwarmClient.HasNamespaceSupport(context.Background())

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, ok, false, "Did not receive expected response")
	})

	t.Run("Should return false if namespace endpoint is not available", func(t *testing.T) {
		ok, err := failingOpenFaaSClient.HasNamespaceSupport(context.Background())

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, ok, false, "Did not receive expected response")
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.HasNamespaceSupport(context.Background())

		assert.Error(t, err, "OpenFaaS Credentials are invalid", "Did receive unexpected error")
	})
}

func TestClient_GetFunctions(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allFn := []types.FunctionStatus{
			{
				Name:              "function-name",
				Image:             "docker:image",
				InvocationCount:   0,
				Replicas:          1,
				EnvProcess:        "",
				AvailableReplicas: 1,
				Labels:            nil,
				Annotations:       nil,
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
				Annotations:       nil,
				Namespace:         "special",
			},
		}

		namespacedFn := []types.FunctionStatus{
			{
				Name:              "wrencher",
				Image:             "docker:image",
				InvocationCount:   0,
				Replicas:          1,
				EnvProcess:        "",
				AvailableReplicas: 1,
				Labels:            nil,
				Annotations:       nil,
				Namespace:         "special",
			},
		}

		if user, pass, ok := r.BasicAuth(); ok {
			if user == "User" && pass == "Pass" {
				w.WriteHeader(200)
				out, _ := json.Marshal(allFn)
				_, _ = w.Write(out)
				return
			}
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return
		}

		if r.Method != fasthttp.MethodGet {
			w.WriteHeader(400)
			fmt.Fprint(w, "Method not supported")
			return
		}

		namespace := r.URL.Query().Get("namespace")
		if len(namespace) > 0 {
			if namespace == "special" {
				w.WriteHeader(200)
				out, _ := json.Marshal(namespacedFn)
				_, _ = w.Write(out)
				return
			}

			w.WriteHeader(500)
			fmt.Fprint(w, "Internal Server Error")
			return
		}

		w.WriteHeader(200)
		out, _ := json.Marshal(allFn)
		_, _ = w.Write(out)
	}))
	defer server.Close()

	openfaasClient := NewClient(CreateClient(server), nil, server.URL)

	authenticatedOpenFaaSClient := NewClient(CreateClient(server), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Invalid",
	}, server.URL)

	t.Parallel()

	t.Run("Should return all functions currently deployed", func(t *testing.T) {
		functions, _ := openfaasClient.GetFunctions(context.Background(), "")

		if len(functions) != 2 {
			t.Errorf("Expected %d functions but received %d function(s)", 2, len(functions))
		}

		for _, fn := range functions {
			if fn.Namespace != "faas" && fn.Namespace != "special" {
				t.Errorf("Expected namespace to be either %s or %s but received %s", "faas", "special", fn.Namespace)
			}
		}
	})

	t.Run("Should return all functions currently deployed in the specified namespace", func(t *testing.T) {
		functions, err := openfaasClient.GetFunctions(context.Background(), "special")

		assert.Nil(t, err, "Should not fail")
		assert.Len(t, functions, 1, "Did not receive expected response")

		for _, fn := range functions {
			assert.Equal(t, fn.Namespace, "special", "Received wrong namespace")
		}
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.GetFunctions(context.Background(), "")

		assert.Error(t, err, "OpenFaaS Credentials are invalid", "Did receive unexpected error")
	})

	t.Run("Should throw error on unexpected status code", func(t *testing.T) {
		_, err := openfaasClient.GetFunctions(context.Background(), "what")

		assert.Error(t, err, "Received unexpected Status Code 500", "Did receive unexpected error")
	})
}

func TestClient_GetNamespaces(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespaces := []string{
			"faas",
			"special",
			"namespace",
		}

		if user, pass, ok := r.BasicAuth(); ok {
			if user == "User" && pass == "Pass" {
				w.WriteHeader(502)
				_, _ = w.Write(nil)
				return
			}
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return
		}

		if r.Method != fasthttp.MethodGet {
			w.WriteHeader(400)
			fmt.Fprint(w, "Method not supported")
			return
		}

		w.WriteHeader(200)
		out, _ := json.Marshal(namespaces)
		_, _ = w.Write(out)
	}))
	defer server.Close()

	openfaasClient := NewClient(CreateClient(server), nil, server.URL)

	authenticatedOpenFaaSClient := NewClient(CreateClient(server), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Invalid",
	}, server.URL)

	failingOpenFaaSClient := NewClient(CreateClient(server), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Pass",
	}, server.URL)

	t.Parallel()

	t.Run("Should return a list of all namespaces", func(t *testing.T) {
		namespaces, err := openfaasClient.GetNamespaces(context.Background())

		assert.Nil(t, err, "Should not fail")
		assert.Len(t, namespaces, 3, "Did not receive expected response")
	})

	t.Run("Should return empty list if unexpected response was received", func(t *testing.T) {
		namespaces, err := failingOpenFaaSClient.GetNamespaces(context.Background())

		assert.Nil(t, err, "Should not fail")
		assert.Len(t, namespaces, 0, "Did not receive expected response")
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.GetNamespaces(context.Background())

		assert.Error(t, err, "OpenFaaS Credentials are invalid", "Did receive unexpected error")
	})
}
