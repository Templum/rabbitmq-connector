// Copyright (c) Simon Pelczer 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"log"

	"github.com/Templum/rabbitmq-connector/pkg/rabbitmq"
	"github.com/Templum/rabbitmq-connector/pkg/version"
	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/streadway/amqp"
)

func main() {
	conf := rabbitmq.BuildConfig()
	// Local creds := &auth.BasicAuthCredentials{User:"admin", Password: "8bd727b83ae1338d7af2b81ed87a02e053aa7351bb6598bb21196cd660c70098"}
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
