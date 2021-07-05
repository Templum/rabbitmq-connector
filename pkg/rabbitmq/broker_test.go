/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package rabbitmq

import (
	"context"
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SkipIfProviderIsNotHealthy(t *testing.T) {
	_, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		t.Skipf("Docker is not running. TestContainers can't perform is work without it: %s", err)
	}
}

func TestBroker_Dial(t *testing.T) {
	SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.7.4",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForListeningPort(nat.Port("5672/tcp")),
		Env:          map[string]string{"RABBITMQ_DEFAULT_USER": "user", "RABBITMQ_DEFAULT_PASS": "pass"},
	}

	rabbitmq, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	//nolint:golint,errcheck
	defer rabbitmq.Terminate(ctx)

	ip, err := rabbitmq.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	port, err := rabbitmq.MappedPort(ctx, nat.Port("5672/tcp"))
	if err != nil {
		t.Error(err)
	}

	broker := NewBroker()
	url := fmt.Sprintf("amqp://user:pass@%s:%s/", ip, port.Port())

	con, err := broker.Dial(url)

	assert.NotNil(t, con, "should return connection")
	assert.NoError(t, err, "should not throw")
}

func TestBroker_DialTLS(t *testing.T) {
	SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.7.4",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForListeningPort(nat.Port("5672/tcp")),
		Env:          map[string]string{"RABBITMQ_DEFAULT_USER": "user", "RABBITMQ_DEFAULT_PASS": "pass"},
	}

	rabbitmq, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	//nolint:golint,errcheck
	defer rabbitmq.Terminate(ctx)

	ip, err := rabbitmq.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	port, err := rabbitmq.MappedPort(ctx, nat.Port("5672/tcp"))
	if err != nil {
		t.Error(err)
	}

	broker := NewBroker()
	url := fmt.Sprintf("amqps://%s:%s/", ip, port.Port())

	con, err := broker.DialTLS(url, &tls.Config{})

	assert.Nil(t, con, "should not support TLS")
	assert.Error(t, err, "should raise an error")
}
