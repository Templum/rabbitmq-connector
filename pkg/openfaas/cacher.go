package openfaas

import (
	"context"
	types2 "github.com/Templum/rabbitmq-connector/pkg/types"
	"log"
	"strings"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
	"github.com/openfaas/faas-provider/types"
)

// Copyright (c) Simon Pelczer 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// Controller is responsible for building up and maintaining a
// Cache with all of the deployed OpenFaaS Functions across
// all namespaces
type Controller struct {
	conf   *config.Controller
	client FunctionCrawler
	cache  TopicMap
}

// NewController returns a new instance
func NewController(conf *config.Controller, client FunctionCrawler, cache TopicMap) *Controller {
	return &Controller{
		conf:   conf,
		client: client,
		cache:  cache,
	}
}

// Start setups the cache and starts continuous caching
func (c *Controller) Start(ctx context.Context) {
	hasNamespaceSupport, _ := c.client.HasNamespaceSupport(ctx)
	timer := time.NewTicker(c.conf.TopicRefreshTime)

	// Initial populating
	c.refreshTick(ctx, hasNamespaceSupport)
	go c.refresh(ctx, timer, hasNamespaceSupport)
}

// Invoke triggers a call to all functions registered to the specified topic
func (c *Controller) Invoke(topic string, invocation *types2.OpenFaaSInvocation) {
	functions := c.cache.GetCachedValues(topic)

	for _, fn := range functions {

		//nolint:golint,errcheck
		go c.client.InvokeSync(context.Background(), fn, invocation)
	}
}

func (c *Controller) refresh(ctx context.Context, ticker *time.Ticker, hasNamespaceSupport bool) {
loop:
	for {
		select {
		case <-ticker.C:
			c.refreshTick(ctx, hasNamespaceSupport)
			break
		case <-ctx.Done():
			log.Println("Received done via context will stop refreshing cache")
			break loop
		}
	}
}

func (c *Controller) refreshTick(ctx context.Context, hasNamespaceSupport bool) {
	builder := NewFunctionMapBuilder()
	var namespaces []string
	var err error

	if hasNamespaceSupport {
		log.Println("Crawling namespaces for functions")
		namespaces, err = c.client.GetNamespaces(ctx)
		if err != nil {
			log.Printf("Received the following error during fetching namespaces %s", err)
			namespaces = []string{}
		}
	} else {
		namespaces = []string{""}
	}

	log.Println("Crawling for functions")
	c.crawlFunctions(ctx, namespaces, builder)

	log.Println("Crawling finished will now refresh the cache")
	c.cache.Refresh(builder.Build())
}

func (c *Controller) crawlFunctions(ctx context.Context, namespaces []string, builder TopicMapBuilder) {
	for _, ns := range namespaces {
		found, err := c.client.GetFunctions(ctx, ns)
		if err != nil {
			log.Printf("Received %s while fetching functions on namespace %s", err, ns)
			found = []types.FunctionStatus{}
		}

		for _, fn := range found {
			topics := c.extractTopicsFromAnnotations(fn)

			for _, topic := range topics {
				builder.Append(topic, fn.Name)
			}
		}
	}
}

func (c *Controller) extractTopicsFromAnnotations(fn types.FunctionStatus) []string {
	topics := []string{}

	if fn.Annotations != nil {
		annotations := *fn.Annotations
		if topicNames, exist := annotations["topic"]; exist {
			topics = strings.Split(topicNames, ",")
		}
	}

	return topics
}
