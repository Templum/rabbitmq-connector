package openfaas

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	types "github.com/Templum/rabbitmq-connector/pkg/types"
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
		Client: server.Client(),
	}

	authenticatedOpenFaaSClient := Client{
		url:    server.URL,
		Client: server.Client(),
		credentials: &types.Credentials{
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
		Client: server.Client(),
	}

	authenticatedOpenFaaSClient := Client{
		url:    server.URL,
		Client: server.Client(),
		credentials: &types.Credentials{
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
			t.Errorf("Expected not unexpected status code error but received %s", err)
		}
	})
}

func BenchmarkClient_Invoke(b *testing.B) {
	var r []byte

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/function/exists":
			w.WriteHeader(200)
			fmt.Fprint(w, "Hello World")
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
		Client: server.Client(),
	}

	for n := 0; n < b.N; n++ {
		payload := []byte("afsdasdfgfghsgdfhdfghdghfdghfdghdfghdfghdfghf")
		resp, _ := openfaasClient.InvokeSync(context.Background(), "exists", payload)
		r = resp
	}

	println(r)
}
