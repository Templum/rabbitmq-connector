package openfaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
)

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

		switch r.URL.Path {
		case "/function/exists":
			w.WriteHeader(200)
			fmt.Fprint(w, expectedResponse)
			break
		case "/function/nonexisting":
			w.WriteHeader(404)
			fmt.Fprint(w, "Not Found")
			break
		default:
			w.WriteHeader(500)
			fmt.Fprint(w, "Internal Server Error")
			break
		}
	}))
	defer server.Close()

	openfaasClient := NewClient(server.Client(), nil, server.URL)

	authenticatedOpenFaaSClient := NewClient(server.Client(), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Invalid",
	}, server.URL)

	t.Parallel()

	t.Run("Should invoke the specified function", func(t *testing.T) {
		payload := []byte("Test")
		resp, err := openfaasClient.InvokeSync(context.Background(), "exists", payload)

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, string(resp), expectedResponse, "Did not receive expected response")
	})

	t.Run("Should except nil as body", func(t *testing.T) {
		resp, err := openfaasClient.InvokeSync(context.Background(), "exists", nil)

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, string(resp), expectedResponse, "Did not receive expected response")
	})

	t.Run("Should throw error if function does not exist", func(t *testing.T) {
		payload := []byte("Test")
		_, err := openfaasClient.InvokeSync(context.Background(), "nonexisting", payload)

		assert.Error(t, err, "Function nonexisting is not deployed", "Did receive unexpected error")
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.InvokeSync(context.Background(), "exists", nil)

		assert.Error(t, err, "OpenFaaS Credentials are invalid", "Did receive unexpected error")
	})

	t.Run("Should throw error on unexpected status code", func(t *testing.T) {
		payload := []byte("Test")
		_, err := openfaasClient.InvokeSync(context.Background(), "internal", payload)

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

		switch r.URL.Path {
		case "/async-function/exists":
			w.WriteHeader(202)
			fmt.Fprint(w, "Hello World")
			break
		case "/async-function/nonexisting":
			w.WriteHeader(404)
			fmt.Fprint(w, "Not Found")
			break
		default:
			w.WriteHeader(500)
			fmt.Fprint(w, "Internal Server Error")
			break
		}
	}))
	defer server.Close()

	openfaasClient := NewClient(server.Client(), nil, server.URL)

	authenticatedOpenFaaSClient := NewClient(server.Client(), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Invalid",
	}, server.URL)

	t.Parallel()

	t.Run("Should invoke the specified function", func(t *testing.T) {
		payload := []byte("Test")
		ok, err := openfaasClient.InvokeAsync(context.Background(), "exists", payload)

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, ok, true, "Did not receive expected response")
	})

	t.Run("Should except nil as body", func(t *testing.T) {
		ok, err := openfaasClient.InvokeAsync(context.Background(), "exists", nil)

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, ok, true, "Did not receive expected response")
	})

	t.Run("Should throw error if function does not exist", func(t *testing.T) {
		payload := []byte("Test")
		_, err := openfaasClient.InvokeAsync(context.Background(), "nonexisting", payload)

		assert.Error(t, err, "Function nonexisting is not deployed", "Did receive unexpected error")
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.InvokeAsync(context.Background(), "exists", nil)

		assert.Error(t, err, "OpenFaaS Credentials are invalid", "Did receive unexpected error")
	})

	t.Run("Should throw error on unexpected status code", func(t *testing.T) {
		payload := []byte("Test")
		_, err := openfaasClient.InvokeAsync(context.Background(), "internal", payload)

		assert.Error(t, err, "Received unexpected Status Code 500", "Did receive unexpected error")
	})
}

func TestClient_HasNamespaceSupport(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, pass, ok := r.BasicAuth(); ok {
			if user == "User" && pass == "Pass" {
				w.WriteHeader(502)
				w.Write(nil)
				return
			}
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return
		}

		w.WriteHeader(200)
		w.Write(nil)
		return
	}))
	defer server.Close()

	openfaasClient := NewClient(server.Client(), nil, server.URL)

	authenticatedOpenFaaSClient := NewClient(server.Client(), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Invalid",
	}, server.URL)

	failingOpenFaaSClient := NewClient(server.Client(), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Pass",
	}, server.URL)

	t.Parallel()

	t.Run("Should return true if namespaces endpoint available", func(t *testing.T) {
		ok, err := openfaasClient.HasNamespaceSupport(context.Background())

		assert.Nil(t, err, "Should not fail")
		assert.Equal(t, ok, true, "Did not receive expected response")
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
			types.FunctionStatus{
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
			types.FunctionStatus{
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
			types.FunctionStatus{
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
				w.Write(out)
				return
			}
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return
		}

		namespace := r.URL.Query().Get("namespace")
		if len(namespace) > 0 {
			if namespace == "special" {
				w.WriteHeader(200)
				out, _ := json.Marshal(namespacedFn)
				w.Write(out)
				return
			}

			w.WriteHeader(500)
			fmt.Fprint(w, "Internal Server Error")
			return
		}

		w.WriteHeader(200)
		out, _ := json.Marshal(allFn)
		w.Write(out)
		return
	}))
	defer server.Close()

	openfaasClient := NewClient(server.Client(), nil, server.URL)

	authenticatedOpenFaaSClient := NewClient(server.Client(), &auth.BasicAuthCredentials{
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
				w.Write(nil)
				return
			}
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return
		}

		w.WriteHeader(200)
		out, _ := json.Marshal(namespaces)
		w.Write(out)
		return
	}))
	defer server.Close()

	openfaasClient := NewClient(server.Client(), nil, server.URL)

	authenticatedOpenFaaSClient := NewClient(server.Client(), &auth.BasicAuthCredentials{
		User:     "User",
		Password: "Invalid",
	}, server.URL)

	failingOpenFaaSClient := NewClient(server.Client(), &auth.BasicAuthCredentials{
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
