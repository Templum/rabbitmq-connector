// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/version"
	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
)

func main() {
	commit, version := version.GetReleaseInfo()
	log.Printf("OpenFaaS RabbitMQ Connector [Version: %s Commit: %s]", version, commit)

	creds := types.GetCredentials()
	if version == "dev" {
		creds = &auth.BasicAuthCredentials{User: "admin", Password: "b43c1de00d8a477d6af007a6516944e3d1b02692a190fe71f68616b678ac959a"} // Local
	}

	controllerConf := &types.ControllerConfig{
		RebuildInterval: config.GetRefreshTime(),
		GatewayURL:      config.GetOpenFaaSUrl(),
		PrintResponse:   false,
	}

	controller := types.NewController(creds, controllerConf)
	fmt.Println(controller)
	controller.BeginMapBuilder()

	// TODO: Wait at least for the first map sync

	connector := rabbitmq.MakeConnector(config.GenerateRabbitMQUrl(), controller)
	connector.StartConnector()
	defer connector.Close()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	sig := <-signalChannel
	switch sig {
	case os.Interrupt:
		log.Printf("Recieved SIGINT preparing for shutdown")
		connector.Close()
	case syscall.SIGTERM:
		log.Printf("Recieved SIGTERM shutting down")
		connector.Close()
	}
}
