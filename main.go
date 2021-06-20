/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/connector"
	"github.com/Templum/rabbitmq-connector/pkg/openfaas"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/types"
	"github.com/Templum/rabbitmq-connector/pkg/version"

	_ "go.uber.org/automaxprocs"
)

func main() {
	commit, tag := version.GetReleaseInfo()
	log.Printf("OpenFaaS RabbitMQ Connector [Version: %s Commit: %s]", tag, commit)

	if rawValue, ok := os.LookupEnv("basic_auth"); ok {
		active, _ := strconv.ParseBool(rawValue)
		if path, ok := os.LookupEnv("secret_mount_path"); ok && active {
			log.Printf("Will read basic64 secret from path %s which was set via 'secret_mount_path'", path)
		}
	}

	// Building our Config from envs
	conf, validationErr := config.NewConfig()
	if validationErr != nil {
		log.Fatalf("During Config validation %s occurred.", validationErr)
	}

	// Setup Application Context to ensure gracefully shutdowns
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient := types.MakeHTTPClient(conf.InsecureSkipVerify, conf.MaxClientsPerHost, 60*time.Second)
	// Setup OpenFaaS Controller which is used for querying and more
	ofSDK := openfaas.NewController(conf, openfaas.NewClient(httpClient, conf.BasicAuth, conf.GatewayURL), openfaas.NewTopicFunctionCache())
	go ofSDK.Start(ctx)
	log.Printf("Started Cache Task which populates the topic map")

	c := connector.New(rabbitmq.NewConnectionManager(rabbitmq.NewBroker(), conf.TLSConfig), rabbitmq.NewFactory(), ofSDK, conf)
	err := c.Run()

	if err != nil {
		log.Fatalf("Received %s during Connector starting", err)
	}

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	sig := <-signalChannel
	switch sig {
	case os.Interrupt:
		log.Printf("Received SIGINT preparing for shutdown")

		c.Shutdown()
		cancel()
	case syscall.SIGTERM:
		log.Printf("Received SIGTERM shutting down")
		c.Shutdown()
		cancel()
	}
}
