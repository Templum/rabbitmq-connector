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

	openfaasClient := Client{
		url:    server.URL,
		client: server.Client(),
	}

	authenticatedOpenFaaSClient := Client{
		url:    server.URL,
		client: server.Client(),
		credentials: &auth.BasicAuthCredentials{
			User:     "User",
			Password: "Invalid",
		},
	}

	t.Parallel()

	t.Run("Should invoke the specified function", func(t *testing.T) {
		payload := []byte("Test")
		resp, err := openfaasClient.InvokeSync(context.Background(), "exists", payload)

		if err != nil {
			t.Errorf("Received unexpected error %s", err)
		}

		if string(resp) != expectedResponse {
			t.Errorf("Expected response to be %s but received %s", expectedResponse, string(resp))
		}
	})

	t.Run("Should except nil as body", func(t *testing.T) {
		resp, err := openfaasClient.InvokeSync(context.Background(), "exists", nil)

		if err != nil {
			t.Errorf("Received unexpected error %s", err)
		}

		if string(resp) != expectedResponse {
			t.Errorf("Expected response to be %s but received %s", expectedResponse, string(resp))
		}
	})

	t.Run("Should throw error if function does not exist", func(t *testing.T) {
		payload := []byte("Test")
		_, err := openfaasClient.InvokeSync(context.Background(), "nonexisting", payload)

		if err.Error() != "Function nonexisting is not deployed" {
			t.Errorf("Expected not deployed error but received %s", err)
		}
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.InvokeSync(context.Background(), "exists", nil)

		if err.Error() != "OpenFaaS Credentials are invalid" {
			t.Errorf("Expected invalid credentials error but received %s", err)
		}
	})

	t.Run("Should throw error on unexpected status code", func(t *testing.T) {
		payload := []byte("Test")
		_, err := openfaasClient.InvokeSync(context.Background(), "internal", payload)

		if err.Error() != "Received unexpected Status Code 500" {
			t.Errorf("Expected not unexpected status code error but received %s", err)
		}
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

	openfaasClient := Client{
		url:    server.URL,
		client: server.Client(),
	}

	authenticatedOpenFaaSClient := Client{
		url:    server.URL,
		client: server.Client(),
		credentials: &auth.BasicAuthCredentials{
			User:     "User",
			Password: "Invalid",
		},
	}

	t.Parallel()

	t.Run("Should invoke the specified function", func(t *testing.T) {
		payload := []byte("Test")
		ok, err := openfaasClient.InvokeAsync(context.Background(), "exists", payload)

		if err != nil {
			t.Errorf("Received unexpected error %s", err)
		}

		if !ok {
			t.Errorf("Expected %t but received %t", true, ok)
		}
	})

	t.Run("Should except nil as body", func(t *testing.T) {
		ok, err := openfaasClient.InvokeAsync(context.Background(), "exists", nil)

		if err != nil {
			t.Errorf("Received unexpected error %s", err)
		}

		if !ok {
			t.Errorf("Expected %t but received %t", true, ok)
		}
	})

	t.Run("Should throw error if function does not exist", func(t *testing.T) {
		payload := []byte("Test")
		_, err := openfaasClient.InvokeAsync(context.Background(), "nonexisting", payload)

		if err.Error() != "Function nonexisting is not deployed" {
			t.Errorf("Expected not deployed error but received %s", err)
		}
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.InvokeAsync(context.Background(), "exists", nil)

		if err.Error() != "OpenFaaS Credentials are invalid" {
			t.Errorf("Received unexpected error %s", err)
		}
	})

	t.Run("Should throw error on unexpected status code", func(t *testing.T) {
		payload := []byte("Test")
		_, err := openfaasClient.InvokeAsync(context.Background(), "internal", payload)

		if err.Error() != "Received unexpected Status Code 500" {
			t.Errorf("Expected unexpected status code error but received %s", err)
		}
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

	openfaasClient := Client{
		url:    server.URL,
		client: server.Client(),
	}

	authenticatedOpenFaaSClient := Client{
		url:    server.URL,
		client: server.Client(),
		credentials: &auth.BasicAuthCredentials{
			User:     "User",
			Password: "Invalid",
		},
	}

	failingOpenFaaSClient := Client{
		url:    server.URL,
		client: server.Client(),
		credentials: &auth.BasicAuthCredentials{
			User:     "User",
			Password: "Pass",
		},
	}

	t.Parallel()

	t.Run("Should return true if namespaces endpoint available", func(t *testing.T) {
		ok, _ := openfaasClient.HasNamespaceSupport(context.Background())

		if !ok {
			t.Errorf("Expected %t but received %t", false, ok)
		}
	})

	t.Run("Should return false if namespace endpoint is not available", func(t *testing.T) {
		ok, _ := failingOpenFaaSClient.HasNamespaceSupport(context.Background())

		if ok {
			t.Errorf("Expected %t but received %t", true, ok)
		}
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.HasNamespaceSupport(context.Background())

		if err.Error() != "OpenFaaS Credentials are invalid" {
			t.Errorf("Received unexpected error %s", err)
		}
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

	openfaasClient := Client{
		url:    server.URL,
		client: server.Client(),
	}

	authenticatedOpenFaaSClient := Client{
		url:    server.URL,
		client: server.Client(),
		credentials: &auth.BasicAuthCredentials{
			User:     "User",
			Password: "Invalid",
		},
	}

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
		functions, _ := openfaasClient.GetFunctions(context.Background(), "special")

		if len(functions) != 1 {
			t.Errorf("Expected %d functions but received %d function(s)", 1, len(functions))
		}

		for _, fn := range functions {
			if fn.Namespace != "special" {
				t.Errorf("Expected namespace to be %s but received %s", "special", fn.Namespace)
			}
		}
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.GetFunctions(context.Background(), "")

		if err.Error() != "OpenFaaS Credentials are invalid" {
			t.Errorf("Received unexpected error %s", err)
		}
	})

	t.Run("Should throw error on unexpected status code", func(t *testing.T) {
		_, err := openfaasClient.GetFunctions(context.Background(), "what")

		if err.Error() != "Received unexpected Status Code 500" {
			t.Errorf("Expected not unexpected status code error but received %s", err)
		}
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

	openfaasClient := Client{
		url:    server.URL,
		client: server.Client(),
	}

	authenticatedOpenFaaSClient := Client{
		url:    server.URL,
		client: server.Client(),
		credentials: &auth.BasicAuthCredentials{
			User:     "User",
			Password: "Invalid",
		},
	}

	failingOpenFaaSClient := Client{
		url:    server.URL,
		client: server.Client(),
		credentials: &auth.BasicAuthCredentials{
			User:     "User",
			Password: "Pass",
		},
	}

	t.Parallel()

	t.Run("Should return a list of all namespaces", func(t *testing.T) {
		namespaces, _ := openfaasClient.GetNamespaces(context.Background())

		if len(namespaces) != 3 {
			t.Errorf("Expected %d namespaces but received %d namespaces", 3, len(namespaces))
		}
	})

	t.Run("Should return empty list if unexpected response was received", func(t *testing.T) {
		namespaces, _ := failingOpenFaaSClient.GetNamespaces(context.Background())

		if len(namespaces) > 0 {
			t.Errorf("Expected namespaces to be empty but received %d namespaces", len(namespaces))
		}
	})

	t.Run("Should throw error if unauthorized", func(t *testing.T) {
		_, err := authenticatedOpenFaaSClient.GetNamespaces(context.Background())

		if err.Error() != "OpenFaaS Credentials are invalid" {
			t.Errorf("Received unexpected error %s", err)
		}
	})
}
