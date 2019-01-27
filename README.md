# OpenFaaS RabbitMQ Connector

[![Go Report Card](https://goreportcard.com/badge/github.com/Templum/rabbitmq-connector)](https://goreportcard.com/report/github.com/Templum/rabbitmq-connector)
[![Build Status](https://travis-ci.org/Templum/rabbitmq-connector.svg?branch=develop)](https://travis-ci.org/Templum/rabbitmq-connector)
[![codecov](https://codecov.io/gh/Templum/rabbitmq-connector/branch/develop/graph/badge.svg)](https://codecov.io/gh/Templum/rabbitmq-connector)

This project is an unofficial connector/trigger for OpenFaaS, which allows triggering deployed function from RabbitMQ.
It leverages `Routing keys` for which OpenFaaS functions can be registered. More on the usage can be read [here](#Usage).

## Usage

Using the [OpenFaaS CLI](https://github.com/openfaas/faas-cli) or [Rest API](https://github.com/openfaas/faas/tree/master/api-docs) 
deploy a function which has an `annotation` named `topic`, this has to be a comma-separated string of the relevant topics. 
E.g. `log,monitoring,billing`. Please also make sure to check out the official Rabbit MQ documentation [here](https://www.rabbitmq.com/production-checklist.html) and [here](https://www.rabbitmq.com/monitoring.html) 
in order to avoid message dropping. The connector is configured to spawn workers per topic based on the available CPU's,
however it will try to spawn at least 1 worker per topic. Currently, the connector operates on 1 exchange and leverage 1
Queue. Information on the Topology of the Queue can be found [here](#Topology)

### Configuration Options (via Environment Variables):

* `basic_auth`: Toggle to activate or deactivate basic_auth (E.g `1` || `true`) 
* `secret_mount_path`: The path to a file containing the basic auth secret for the OpenFaaS gateway
* `OPEN_FAAS_GW_URL`: URL to the OpenFaaS gateway defaults to `http://gateway:8080`

* `RMQ_TOPICS`: Rabbit MQ Topics which are relevant in an comma separated list. E.g. `billing,account,support`
* `RMQ_HOST`: Hostname/ip of Rabbit MQ 
* `RMQ_PORT`: Port of Rabbit MQ
* `RMQ_USER`: Defaults to `guest`
* `RMQ_PASS`: Defaults to `guest`
* `RMQ_QUEUE`: Queue Name will default to `OpenFaasQueue`
* `RMQ_EXCHANGE`: Exchange Name will default to `OpenFaasEx` 

* `REQ_TIMEOUT`: Request Timeout for invocations of OpenFaaS functions defaults to `30s`
* `TOPIC_MAP_REFRESH_TIME`: Refresh time for the topic map defaults to `60s`

### Topology

1 Exchange & 1 Topic, which can be set using `RMQ_QUEUE` & `RMQ_EXCHANGE`.

**Exchange:**

```
    Kind: "direct",
    Durable: true,
    Auto Delete: false,
    Internal: false,
    No Wait: false
```

**Queue:**

```
    Durable: true,
    Auto Delete: false,
    Internal: false,
    No Wait: false
```

## Deployment 

### Kubernetes

The required files to deploy OpenFaaS RabbitMQ connector can be found under `artifacts`. It assumes that OpenFaaS was
deployed as described [here](https://github.com/openfaas/faas-netes/blob/master/yaml/README.md). The default config is
setup to work with OpenFaaS as described there. Further within [connector-cfg.yaml](./artifacts/connector-cfg.yaml) there
are values that need to be override with your Rabbit MQ setup, they are marked with `replace_me`.

### Docker

The required file to deploy OpenFaaS RabbitMQ connector can be found under `artifacts`. It assumes that OpenFaaS was
deployed using this [script](https://github.com/openfaas/faas/blob/master/deploy_stack.sh). The [docker-compose.yml](./artifacts/docker-compose.yml) ships with an 
RabbitMQ node and an producer, you might want to remove them. If you want to use an existing Rabbit MQ setup make sure to
override `RMQ` accordingly.

### Local

1. Start an local Rabbit MQ Broker, this can be done with`/hack/development_env_setup.sh`.
2. Expose the necessary environment variables, in GoLand this can be done as part of the configuration
3. Deploy at least one function that listens to a topic: `$ faas-cli store deploy figlet --annotation topic="account"`
4. Start a producer to generate messages 



## Bug Reporting & Feature Requests

Please feel free to report any issues or Feature request on the [Issue Tab](https://github.com/Templum/rabbitmq-connector/issues). 