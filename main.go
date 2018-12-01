// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/openfaas/faas-provider/auth"
	"log"

	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/version"
	"github.com/openfaas-incubator/connector-sdk/types"
)

func main() {
	commit, version := version.GetReleaseInfo()
	log.Printf("OpenFaaS RabbitMQ Connector [Version: %s Commit: %s]", version, commit)

	creds := &auth.BasicAuthCredentials{User: "admin", Password: "b43c1de00d8a477d6af007a6516944e3d1b02692a190fe71f68616b678ac959a"} // Local
	// creds := types.GetCredentials()

	controllerConf := &types.ControllerConfig{
		RebuildInterval: config.GetRefreshTime(),
		GatewayURL:      config.GetOpenFaaSUrl(),
		PrintResponse:   false,
	}

	controller := types.NewController(creds, controllerConf)
	fmt.Println(controller)
	controller.BeginMapBuilder()


	connector := rabbitmq.MakeConnector(config.GenerateRabbitMQUrl(), controller)
	connector.StartConnector()
	defer connector.Close()

	// TODO: Listen for Signals

	forever := make(chan bool)
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
