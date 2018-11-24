// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"github.com/Templum/openfaas-rabbitmq-connector/pkg/version"
	"log"

	"github.com/Templum/openfaas-rabbitmq-connector/pkg/rabbitmq"
	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/streadway/amqp"
)

func main() {
	conf := rabbitmq.BuildConfig()
	creds := types.GetCredentials()

	config := &types.ControllerConfig{
		RebuildInterval: conf.TopicMapRefreshTime,
		GatewayURL:      conf.GatewayURL,
		PrintResponse:   true,
	}

	controller := types.NewController(creds, config)
	fmt.Println(controller)
	controller.BeginMapBuilder()

	commit, version := version.GetReleaseInfo()
	log.Printf("OpenFaaS RabbitMQ Connector [Version: %s Commit: %s]", version, commit)

	forever := make(chan bool)

	go rabbitmq.MakeConnector(&conf, func(delivery amqp.Delivery) {
		log.Printf("Received Message [%s] on Topic [%s] of Type [%s]", delivery.Body, delivery.RoutingKey, delivery.ContentType)
		controller.Invoker.Invoke(controller.TopicMap, delivery.RoutingKey, &delivery.Body)
	})
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
