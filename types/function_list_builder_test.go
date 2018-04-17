// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfaas/faas/gateway/requests"
)

func TestBuildSingleMatchingFunction(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		functions := []requests.Function{}
		labelMap := make(map[string]string)
		labelMap["topic"] = "topic1"

		functions = append(functions, requests.Function{
			Name:   "echo",
			Labels: &labelMap,
		})
		bytesOut, _ := json.Marshal(functions)
		w.Write(bytesOut)
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:     client,
		GatewayURL: srv.URL,
	}

	lookup, err := builder.Build()
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(lookup) != 1 {
		t.Errorf("Lookup - want: %d items, got: %d", 1, len(lookup))
	}
}

func TestBuildNoFunctions(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		functions := []requests.Function{}
		bytesOut, _ := json.Marshal(functions)
		w.Write(bytesOut)
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:     client,
		GatewayURL: srv.URL,
	}

	lookup, err := builder.Build()
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(lookup) != 0 {
		t.Errorf("Lookup - want: %d items, got: %d", 0, len(lookup))
	}
}
