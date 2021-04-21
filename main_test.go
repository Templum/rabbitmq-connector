// +build integration

/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/openfaas"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	t "github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
)

func getPathToExampleTopology() string {
	dir, _ := os.Getwd()
	return path.Join(dir, "artifacts", "example_topology.yaml")
}

const TOPIC = "Foo"

func TestMain(m *testing.M) {

	fmt.Print("Integration Test")
	_ = os.Setenv("basic_auth", "false")
	_ = os.Setenv("OPEN_FAAS_GW_URL", "http://localhost:8080")
	_ = os.Setenv("RMQ_USER", "user")
	_ = os.Setenv("RMQ_PASS", "pass")
	_ = os.Setenv("RMQ_HOST", "localhost")
	_ = os.Setenv("PATH_TO_TOPOLOGY", getPathToExampleTopology())

	defer os.Unsetenv("basic_auth")
	defer os.Unsetenv("OPEN_FAAS_GW_URL")
	defer os.Unsetenv("RMQ_USER")
	defer os.Unsetenv("RMQ_PASS")
	defer os.Unsetenv("RMQ_HOST")
	defer os.Unsetenv("PATH_TO_TOPOLOGY")

	os.Exit(m.Run())
}

func getOpenFaaSClient() openfaas.FunctionFetcher {
	httpClient := types.MakeHTTPClient(false, 256, 60*time.Second)
	ofClient := openfaas.NewClient(httpClient, nil, os.Getenv("OPEN_FAAS_GW_URL"))
	return ofClient
}

func getIntegrationFaaSFunction(client openfaas.FunctionFetcher) t.FunctionStatus {
	functions, _ := client.GetFunctions(context.Background(), "")
	return functions[0]
}

func publishMessage(client http.Client, topic string, message string) error {
	body := fmt.Sprintf("{\"properties\":{},\"routing_key\":\"%s\",\"payload\":\"%s\",\"payload_encoding\":\"string\"}", topic, message)

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:15672/api/exchanges/%2f/AEx/publish", strings.NewReader(body))
	req.SetBasicAuth("user", "pass")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("unexpected status code")
	}

	return nil
}

func Test_main(t *testing.T) {
	go main()

	time.Sleep(1 * time.Second)

	client := getOpenFaaSClient()
	before := getIntegrationFaaSFunction(client)

	assert.GreaterOrEqual(t, before.InvocationCount, float64(0), "should be 0 or more")
	assert.Contains(t, (*before.Annotations)["topic"], TOPIC, "should listen for TOPIC Foo")

	publishedMessages := 0

	for i := 0; i < 1000; i++ {
		err := publishMessage(http.Client{}, TOPIC, "Hello World!")
		if err == nil {
			publishedMessages += 1
		}
	}

	time.Sleep(5 * time.Second)

	after := getIntegrationFaaSFunction(client)
	assert.Greater(t, after.InvocationCount, before.InvocationCount)
	assert.GreaterOrEqual(t, after.InvocationCount, float64(1000), "should invoked at least 1000 times")
}
