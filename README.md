# OpenFaaS RabbitMQ Connector

[![Go Report Card](https://goreportcard.com/badge/github.com/Templum/rabbitmq-connector)](https://goreportcard.com/report/github.com/Templum/rabbitmq-connector)
[![CodeFactor](https://www.codefactor.io/repository/github/templum/rabbitmq-connector/badge)](https://www.codefactor.io/repository/github/templum/rabbitmq-connector)
![CI](https://github.com/Templum/rabbitmq-connector/workflows/CI/badge.svg)
![Docker Release](https://github.com/Templum/rabbitmq-connector/workflows/Docker%20Release/badge.svg)
[![codecov](https://codecov.io/gh/Templum/rabbitmq-connector/branch/develop/graph/badge.svg)](https://codecov.io/gh/Templum/rabbitmq-connector)

This project is an unofficial trigger for OpenFaaS functions based on RabbitMQ Messages. Where it leverages the
`Routing keys` to call OpenFaaS functions which listen to that `topic`. For usage information please go to [here](#Usage).

## Usage

Using the [OpenFaaS CLI](https://github.com/openfaas/faas-cli) or [Rest API](https://github.com/openfaas/faas/tree/master/api-docs)
deploy a function which has an `annotation` named `topic`, this has to be a comma-separated string of the relevant topics.
E.g. `log,monitoring,billing`.

In case an error occurred during the invocation of the function(s) the message is attempted to be transferred back to the Queue. Therefore you should ensure your functions can handle being called potentially twice with the same payload.

Further the returned output from the function is ignored, as the connector currently only supports fire & forget flows.

Please also make sure to check out the official Rabbit MQ documentation [here](https://www.rabbitmq.com/production-checklist.html) and [here](https://www.rabbitmq.com/monitoring.html) in order to avoid message dropping.

### Configuration

General Connector:

* `basic_auth`: Toggle to activate or deactivate basic_auth (E.g `1` || `true`)
* `secret_mount_path`: The path to a file containing the basic auth secret for the OpenFaaS gateway
* `OPEN_FAAS_GW_URL`: URL to the OpenFaaS gateway defaults to `http://gateway:8080`
* `REQ_TIMEOUT`: Request Timeout for invocations of OpenFaaS functions defaults to `30s`
* `TOPIC_MAP_REFRESH_TIME`: Refresh time for the topic map defaults to `60s`
* `INSECURE_SKIP_VERIFY`: Allows to skip verification of HTTP Cert for Communication Connector <=> OpenFaaS default is `false`. It is recommended to keep false, as enabling it opens up the possibility of a man in the middle attack.
* `MAX_CLIENT_PER_HOST`: Allows to specify the maximum number connections/clients that will be opened to an individual host (function), defaults to `256`.

RabbitMQ Related:

* `RMQ_HOST`: Hostname/ip of Rabbit MQ
* `RMQ_PORT`: Port of Rabbit MQ
* `RMQ_VHOST`: Used to specify the vhost for Rabbit MQ, will default to `/`
* `RMQ_USER`: Defaults to `guest`
* `RMQ_PASS`: Defaults to `pass`
* `PATH_TO_TOPOLOGY`: Path to the yaml describing the topology, has _no_ default and is *required*

### Topology Configuration

Compared to v0 this is the biggest change, you can bring your existing Exchange definition to the connector.
The topology is defined in the following format ([Example](./artifacts/example_topology.yaml)):

```yaml
# Name of the exchange
- name: Exchange_Name # Required
  topics: [Foo, Bar] # Required
  # Do we need to declare the exchange ? If it already exists it verifies that the exchange matches the configuration
  declare: true # Default: false
  # Either direct or topic
  type: "direct" # Required 
  # Persistence of Exchange between Rabbit MQ Server restarts
  durable: false # Default: false
  # Auto Deletes Exchange once all consumer are gone
  auto-deleted: false # Default: false
```

Queues will be configured accordingly to there exchange declaration in regards to `durable` & `auto-deleted`. Further the name of the queue
will be generated based on the following schema: `OpenFaaS_{Exchange_Name}_${Topic}`.

## Bug Reporting & Feature Requests

Please feel free to report any issues or Feature request on the [Issue Tab](https://github.com/Templum/rabbitmq-connector/issues).
