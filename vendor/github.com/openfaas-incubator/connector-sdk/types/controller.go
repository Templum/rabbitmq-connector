package types

import (
	"log"
	"time"

	"github.com/openfaas/faas-provider/auth"
)

// ControllerConfig configures a connector SDK controller
type ControllerConfig struct {
	UpstreamTimeout time.Duration
	GatewayURL      string
	PrintResponse   bool
	RebuildInterval time.Duration
}

// Controller for the connector SDK
type Controller struct {
	Config      *ControllerConfig
	Invoker     *Invoker
	TopicMap    *TopicMap
	Credentials *auth.BasicAuthCredentials
}

// NewController create a new connector SDK controller
func NewController(credentials *auth.BasicAuthCredentials, config *ControllerConfig) *Controller {

	invoker := Invoker{
		PrintResponse: config.PrintResponse,
		Client:        MakeClient(config.UpstreamTimeout),
		GatewayURL:    config.GatewayURL,
	}
	topicMap := NewTopicMap()

	return &Controller{
		Config:      config,
		Invoker:     &invoker,
		TopicMap:    &topicMap,
		Credentials: credentials,
	}
}

// Invoke attempts to invoke any functions which match the
// topic the incoming message was published on.
func (c *Controller) Invoke(topic string, message *[]byte) {
	c.Invoker.Invoke(c.TopicMap, topic, message)
}

// BeginMapBuilder begins to build a map of function->topic by
// querying the API gateway.
func (c *Controller) BeginMapBuilder() {

	lookupBuilder := FunctionLookupBuilder{
		GatewayURL:  c.Config.GatewayURL,
		Client:      MakeClient(c.Config.UpstreamTimeout),
		Credentials: c.Credentials,
	}

	ticker := time.NewTicker(c.Config.RebuildInterval)
	go synchronizeLookups(ticker, &lookupBuilder, c.TopicMap)
}

func synchronizeLookups(ticker *time.Ticker,
	lookupBuilder *FunctionLookupBuilder,
	topicMap *TopicMap) {

	for {
		<-ticker.C
		lookups, err := lookupBuilder.Build()
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Syncing topic map")
		topicMap.Sync(&lookups)
	}
}
