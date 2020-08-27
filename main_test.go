/*
 * Copyright (c) Simon Pelczer 2020. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

// +build !race

package main

import (
	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSystem(t *testing.T) {
	const ip = "localhost"
	const port = "5672"

	t.Run("should configure itself from env", func(t *testing.T) {
		os.Setenv("RMQ_USER", "user")
		os.Setenv("RMQ_PASS", "pass")
		os.Setenv("RMQ_HOST", ip)
		os.Setenv("RMQ_PORT", port)
		os.Setenv("PATH_TO_TOPOLOGY", "./artifacts/topology.yaml")

		conf, err := config.NewConfig()

		assert.Nil(t, err, "Should not throw")
		assert.NotNil(t, conf, "Should not be nil")
	})

	t.Run("should start listening to rabbit", func(t *testing.T) {
		os.Setenv("RMQ_USER", "user")
		os.Setenv("RMQ_PASS", "pass")
		os.Setenv("RMQ_HOST", ip)
		os.Setenv("RMQ_PORT", port)
		os.Setenv("PATH_TO_TOPOLOGY", "./artifacts/topology.yaml")

		main()
	})

}
