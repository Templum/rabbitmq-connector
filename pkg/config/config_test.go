/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package config

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	dir, _ := os.Getwd()

	var relativePath string
	// Workaround for Linux
	if runtime.GOOS == "windows" {
		relativePath = path.Join(dir, "..", "..", "..", "artifacts", "example_topology.yaml")
	} else {
		relativePath = path.Join(dir, "..", "..", "artifacts", "example_topology.yaml")
	}

	pathToExampleToplogy, _ := filepath.Abs(relativePath)

	t.Run("With invalid Gateway Url", func(t *testing.T) {
		os.Setenv("OPEN_FAAS_GW_URL", "gateway:8080")
		defer os.Unsetenv("OPEN_FAAS_GW_URL")

		var err error

		_, err = NewConfig(afero.NewMemMapFs())

		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "does not include the protocol http / https", "Did not throw correct error")

		os.Setenv("OPEN_FAAS_GW_URL", "tcp://gateway:8080")
		_, err = NewConfig(afero.NewMemMapFs())
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "does not include the protocol http / https", "Did not throw correct error")
	})

	t.Run("With invalid Rabbit MQ Port", func(t *testing.T) {
		os.Setenv("RMQ_PORT", "is_string")
		defer os.Unsetenv("RMQ_PORT")

		var err error

		_, err = NewConfig(afero.NewMemMapFs())
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "is not a valid port", "Did not throw correct error")

		os.Setenv("RMQ_PORT", "-1")
		_, err = NewConfig(afero.NewMemMapFs())
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "is outside of the allowed port range", "Did not throw correct error")

		os.Setenv("RMQ_PORT", "65536")
		_, err = NewConfig(afero.NewMemMapFs())
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "is outside of the allowed port range", "Did not throw correct error")
	})

	t.Run("With invalid RefreshTime", func(t *testing.T) {
		os.Setenv("TOPIC_MAP_REFRESH_TIME", "is_string")
		defer os.Unsetenv("TOPIC_MAP_REFRESH_TIME")

		var duration time.Duration

		duration = getRefreshTime()
		assert.Equal(t, duration, 30*time.Second, "Should fallback to 30s")

		os.Setenv("TOPIC_MAP_REFRESH_TIME", "66,31h")
		duration = getRefreshTime()
		assert.Equal(t, duration, 30*time.Second, "Should fallback to 30s")
	})

	t.Run("With invalid SkipVerify", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)
		os.Setenv("INSECURE_SKIP_VERIFY", "is_string")

		defer os.Unsetenv("PATH_TO_TOPOLOGY")
		defer os.Unsetenv("INSECURE_SKIP_VERIFY")

		config, err := NewConfig(afero.NewMemMapFs())

		assert.Nil(t, err, "Should not throw")
		assert.False(t, config.InsecureSkipVerify, "Expected default value")
	})

	t.Run("With invalid max clients", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)
		os.Setenv("MAX_CLIENT_PER_HOST", "fifty")

		defer os.Unsetenv("PATH_TO_TOPOLOGY")
		defer os.Unsetenv("MAX_CLIENT_PER_HOST")

		config, err := NewConfig(afero.NewMemMapFs())

		assert.Nil(t, err, "Should not throw")
		assert.Equal(t, config.MaxClientsPerHost, 256, "Expected default value")
	})

	t.Run("With non existing Topology", func(t *testing.T) {
		_, err := NewConfig(afero.NewMemMapFs())
		assert.Error(t, err, "Should throw err")
		assert.Contains(t, err.Error(), "provided topology is either non existing or does not end with .yaml")
	})

	t.Run("With invalid Topology", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", "../../artifacts/connector-cfg.yaml")
		defer os.Unsetenv("PATH_TO_TOPOLOGY")

		_, err := NewConfig(afero.NewMemMapFs())
		assert.NotNil(t, err, "Should throw err")
		assert.Contains(t, err.Error(), " cannot unmarshal")
	})

	t.Run("Default Config", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)
		defer os.Unsetenv("PATH_TO_TOPOLOGY")

		config, err := NewConfig(afero.NewMemMapFs())

		assert.Nil(t, err, "Should not throw")
		assert.Equal(t, config.GatewayURL, "http://gateway:8080", "Expected default value")
		assert.Equal(t, config.RabbitConnectionURL, "amqp://user:pass@localhost:5672/", "Expected default value")
		assert.NotContains(t, config.RabbitSanitizedURL, "user:pass", "Expected credentials not to be present")
		assert.Equal(t, config.RabbitSanitizedURL, "amqp://localhost:5672/", "Expected default value")
		assert.Equal(t, config.TopicRefreshTime, 30*time.Second, "Expected default value")
		assert.False(t, config.InsecureSkipVerify, "Expected default value")
		assert.Equal(t, config.MaxClientsPerHost, 256, "Expected default value")
	})

	t.Run("Override Config", func(t *testing.T) {
		os.Setenv("PATH_TO_TOPOLOGY", pathToExampleToplogy)
		os.Setenv("RMQ_HOST", "rabbit")
		os.Setenv("RMQ_PORT", "1337")
		os.Setenv("RMQ_USER", "username")
		os.Setenv("RMQ_PASS", "password")
		os.Setenv("RMQ_VHOST", "other")
		os.Setenv("OPEN_FAAS_GW_URL", "https://gateway")
		os.Setenv("TOPIC_MAP_REFRESH_TIME", "40s")
		os.Setenv("INSECURE_SKIP_VERIFY", "true")
		os.Setenv("MAX_CLIENT_PER_HOST", "512")

		defer os.Unsetenv("PATH_TO_TOPOLOGY")
		defer os.Unsetenv("RMQ_HOST")
		defer os.Unsetenv("RMQ_PORT")
		defer os.Unsetenv("RMQ_USER")
		defer os.Unsetenv("RMQ_PASS")
		defer os.Unsetenv("RMQ_VHOST")
		defer os.Unsetenv("OPEN_FAAS_GW_URL")
		defer os.Unsetenv("TOPIC_MAP_REFRESH_TIME")
		defer os.Unsetenv("INSECURE_SKIP_VERIFY")
		defer os.Unsetenv("MAX_CLIENT_PER_HOST")

		config, err := NewConfig(afero.NewMemMapFs())

		assert.Nil(t, err, "Should not throw")
		assert.Equal(t, config.GatewayURL, "https://gateway", "Expected override value")
		assert.Equal(t, config.RabbitConnectionURL, "amqp://username:password@rabbit:1337/other", "Expected override value")
		assert.NotContains(t, config.RabbitSanitizedURL, "username:password", "Expected credentials not to be present")
		assert.Equal(t, config.RabbitSanitizedURL, "amqp://rabbit:1337/other", "Expected override value")
		assert.Equal(t, config.TopicRefreshTime, 40*time.Second, "Expected override value")
		assert.True(t, config.InsecureSkipVerify, "Expected override value")
		assert.Equal(t, config.MaxClientsPerHost, 512, "Expected override value")
	})

	t.Run("TLS based Config", func(t *testing.T) {

	})
}
