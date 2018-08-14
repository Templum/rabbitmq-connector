# OpenFaas Rabbit MQ Connector

This connector allows the receiving of messages from Rabbit MQ on defined topics (Routing Keys).
Each message received on the defined (whitelisted) topics, will invoke any deployed function listening on that topic. 

A function has to define its topics in an annotation called `topic`. It is further possible to define more than one topic, in that case, they need to be comma separated (e.g. `billing,accounting,reporting`).

In the current implementation, it is only fire-and-forget. So the returned response will be ignored. But in a feature version, it is planned to allow the writing to a specified topic  (Routing Keys).

## Build it

**Dockerfile:**

The image is available on dockerhub under `templum/rabbitmq-connector:latest`. It is an automated build.

> You can build the connector using `$ docker build -t yourname .` and for the producer `$ cd producer` followed by `$ docker build -t yourname .`

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