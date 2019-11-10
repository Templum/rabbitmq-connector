package openfaas

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestClient_Invoke(t *testing.T) {

	restClient := &http.Client{Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}}

	openfaasClient := Client{
		url:    "http://localhost:8080",
		Client: restClient,
	}

	t.Parallel()

	t.Run("Should invoke the specified function", func(t *testing.T) {
		payload := []byte("Test")
		resp, err := openfaasClient.Invoke(context.Background(), "figlet", payload)

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

	restClient := &http.Client{Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}}

	openfaasClient := Client{
		url:    "http://localhost:8080",
		Client: restClient,
	}

	for n := 0; n < b.N; n++ {
		payload := []byte("Test")
		resp, _ := openfaasClient.Invoke(context.Background(), "figlet", payload)
		r = resp
	}

	println(r)
}
