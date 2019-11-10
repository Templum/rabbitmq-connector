package openfaas

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Invoke(t *testing.T) {

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/function/exists":
			w.WriteHeader(200)
			fmt.Fprintln(w, "Hello World")
			break
		case "/function/nonexisting":
			w.WriteHeader(404)
			fmt.Fprintln(w, "Not Found")
			break
		default:
			w.WriteHeader(500)
			fmt.Fprintln(w, "Internal Server Error")
			break
		}

	}))
	defer server.Close()

	openfaasClient := Client{
		url:    server.URL,
		Client: server.Client(),
	}

	t.Parallel()

	t.Run("Should invoke the specified function", func(t *testing.T) {
		payload := []byte("Test")
		resp, err := openfaasClient.Invoke(context.Background(), "exists", payload)

		if err != nil {
			t.Errorf("Recieved unexpected error %s", err)
		}

		println(&resp)
	})

	t.Run("Should throw error if function does not exist", func(t *testing.T) {
		payload := []byte("Test")
		_, err := openfaasClient.Invoke(context.Background(), "nonexisting", payload)

		if err.Error() != "Function nonexisting is not deployed" {
			t.Errorf("Expeced not deployed error but recieved %s", err)
		}
	})

	t.Run("Should ? if nil is provided", func(t *testing.T) {
		_, err := openfaasClient.Invoke(context.Background(), "nonexisting", nil)

		if err.Error() != "Function nonexisting is not deployed" {
			t.Errorf("Expeced not deployed error but recieved %s", err)
		}
	})
}

func BenchmarkClient_Invoke(b *testing.B) {
	var r []byte

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/function/exists":
			w.WriteHeader(200)
			fmt.Fprintln(w, "Hello World")
			break
		default:
			w.WriteHeader(500)
			fmt.Fprintln(w, "Internal Server Error")
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
		resp, _ := openfaasClient.Invoke(context.Background(), "exists", payload)
		r = resp
	}

	println(r)
}
