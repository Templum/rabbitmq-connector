# OpenFaas Rabbit MQ Connector

This connector allows to receive messages from an Rabbit MQ for defined topics (Routing Keys).
And will invoke functions on the OpenFaas platform that listen on that topic. Currently the
result will be ignored, however in later version it will send the response back on an defined
topic (Routing Key).


## Build it

**Dockerfile:**

You can build the connector using `$ docker build -t yourname .` and for the producer `$ cd producer` followed by `$ docker build -t yourname .`

**Compile:**

You can compile it locally using `$ go build main.go`

## Try it

You can use the following section to try out the current version using docker or kubernetes

### Docker

Will deploy an environment which contains:
* Rabbit MQ `Version 3.7.4`
* Rabbit MQ Connector `Version Latest`
* Sample Producer `Version Latest`

The connector is attached to the **functions**, which should be present if you have deployed
OpenFaas as described [here](https://docs.openfaas.com/deployment/docker-swarm/#20-deploy-the-stack).

1. Deploy Environment `$ docker stack deploy env --compose-file=./deployment/docker-compose.yml`
2. Deploy a function with `topic=account` `$ faas-cli store deploy figlet --label topic="account"`
3. You should be able to follow the successfull invocations using `$ docker service logs -f env.connector`
   * Should Print that it received something along with the result of the figlet invocation.

### Kubernetes

> Nothing available yet