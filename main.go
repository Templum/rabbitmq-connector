// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"log"
	"time"

	"github.com/Templum/rabbitmq-connector/connector"
	"github.com/Templum/rabbitmq-connector/sdk"
	"github.com/streadway/amqp"
)

func syncTopicMapTask(ticker *time.Ticker, client *sdk.OpenFaaSClient, topicMap *sdk.TopicFunctionMap) {
	for {
		<-ticker.C
		functions, err := client.FetchFunctions()

		// TODO: Maybe except 3 sync failures before shutting down
		if err != nil {
			log.Printf("Could not sync functions received %s. Will keep running with existing function map.", err)
		} else {
			log.Printf("Syncing Map with %d Functions", len(*functions))
			topicMap.Sync(functions)
		}
	}
}

func main() {
	conf := connector.BuildConfig()
	httpClient := sdk.MakeClient(conf.RequestTimeout)
	openFaasClient := sdk.NewOpenFaaSClient(conf.GatewayURL, httpClient)
	topicMap := sdk.NewTopicFunctionMap()
	ticker := time.NewTicker(conf.TopicMapRefreshTime)

	forever := make(chan bool)

	go syncTopicMapTask(ticker, &openFaasClient, &topicMap)
	go connector.MakeConnector(&conf, func(delivery amqp.Delivery) {
		log.Printf("Received Message [%s] on Topic [%s] of Type [%s]", delivery.Body, delivery.RoutingKey, delivery.ContentType)
		functions := topicMap.Match(delivery.RoutingKey)

		for _, function := range functions {
			response, _ := openFaasClient.InvokeFunction(function, delivery.Body)
			log.Printf("Function [%s] returned [%s]", function, response)
		}
	})
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
