// Copyright (c) Simon Pelczer 2018. All rights reserved.
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
	"github.com/openfaas/faas-provider/auth"
)

func main() {
	commit, tag := version.GetReleaseInfo()
	log.Printf("OpenFaaS RabbitMQ Connector [Version: %s Commit: %s]", tag, commit)

	var creds *auth.BasicAuthCredentials
	if tag == "dev" {
		log.Printf("Connector is in local mode will use the debug credentials")
		creds = &auth.BasicAuthCredentials{User: "admin", Password: "b43c1de00d8a477d6af007a6516944e3d1b02692a190fe71f68616b678ac959a"}
	} else {
		log.Printf("Connector is in productive mode will read credentials from path")
		creds = types.GetCredentials()
	}

	// Building Connector Specific Config
	connectorCfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("During Config validation %s occured.", err)
	}

	// Building SDK Specific Config
	controllerCfg := &types.ControllerConfig{
		RebuildInterval: connectorCfg.TopicRefreshTime,
		GatewayURL:      connectorCfg.GatewayUrl,
		PrintResponse:   false,
	}

	// Start SDK
	controller := types.NewController(creds, controllerCfg)
	controller.BeginMapBuilder()
	log.Printf("Started Map Building. Be Aware it will take %s until the first map is avaliable.", connectorCfg.TopicRefreshTime)

	factory, err := rabbitmq.NewQueueConsumerFactory(connectorCfg)
	if err != nil {
		log.Fatalf("Connector could not be started recieved %s", err)
	}

	connector := subscriber.NewConnector(connectorCfg, controller, factory)
	connector.Start()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	sig := <-signalChannel
	switch sig {
	case os.Interrupt:
		log.Printf("Recieved SIGINT preparing for shutdown")
		connector.End()
	case syscall.SIGTERM:
		log.Printf("Recieved SIGTERM shutting down")
		connector.End()
	}
}
