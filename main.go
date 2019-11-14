// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/subscriber"
	"github.com/Templum/rabbitmq-connector/pkg/version"
	"github.com/openfaas-incubator/connector-sdk/types"
)

func main() {
	commit, tag := version.GetReleaseInfo()
	log.Printf("OpenFaaS RabbitMQ Connector [Version: %s Commit: %s]", tag, commit)

	if path, ok := os.LookupEnv("secret_mount_path"); ok {
		log.Printf("Will read basic64 secret from path %s which was set via 'secret_mount_path'", path)
	}
	creds := types.GetCredentials()

	// Building Connector Specific Config
	connectorCfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("During Config validation %s occured.", err)
	}

	// Building SDK Specific Config
	controllerCfg := &types.ControllerConfig{
		RebuildInterval: connectorCfg.TopicRefreshTime,
		GatewayURL:      connectorCfg.GatewayURL,
		PrintResponse:   false,
	}

	// Start SDK
	controller := types.NewController(creds, controllerCfg)
	controller.BeginMapBuilder()
	log.Printf("Started Map Building. Be Aware it will take %s until the first map is avaliable.", connectorCfg.TopicRefreshTime)

	factory, err := rabbitmq.NewQueueConsumerFactory(connectorCfg)
	if err != nil {
		log.Fatalf("Connector could not be started Received %s", err)
	}

	connector := subscriber.NewConnector(connectorCfg, controller, factory)
	connector.Start()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	sig := <-signalChannel
	switch sig {
	case os.Interrupt:
		log.Printf("Received SIGINT preparing for shutdown")
		connector.End()
	case syscall.SIGTERM:
		log.Printf("Received SIGTERM shutting down")
		connector.End()
	}
}
