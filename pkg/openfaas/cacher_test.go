package openfaas

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Templum/rabbitmq-connector/pkg/config"
)

func TestCacher(t *testing.T) {
	conf := &config.Controller{TopicRefreshTime: 30 * time.Second}
	client := NewClient(http.DefaultClient, nil, "http://localhost:8080")
	cacher := NewController(conf, client)

	cacher.Start(context.TODO())
}
