# OpenFaaS RabbitMQ Connector

This project is an connector for the OpenFaaS project. It enables the triggering of OpenFaaS functions
based on Rabbit MQ topics (Routing Keys). During deployment of an OpenFaaS function a list of relevant 
topics can be specified, that is used to determine which function should receive the message.

The current implementation is only fire & forget, however this will be adjusted in the future.

## Usage

### General
 
Environment Variables:
* `basic_auth`: The basic auth secret for the OpenFaaS gateway
* `secret_mount_path`: The path to a file containing the basic auth secret for the OpenFaaS gateway
* `OPEN_FAAS_GW_URL`: URL to the OpenFaaS gateway defaults to `http://gateway:8080`

* `RMQ_TOPICS`: Rabbit MQ Topics which are relevant in an comma separated list. E.g. `billing,account,support`
* `RMQ_HOST`: Hostname/ip of Rabbit MQ 
* `RMQ_PORT`: Port of Rabbit MQ
* `RMQ_USER`: Default `guest`
* `RMQ_PASS`: Default `guest`
* `RMQ_QUEUE`: Queue Name will default to `OpenFaasQueue`
* `RMQ_EXCHANGE`: Exchange Name will default to `OpenFaasEx` 

* `REQ_TIMEOUT`: Request Timeout for invocations of OpenFaaS functions defaults to `30s`
* `TOPIC_MAP_REFRESH_TIME`: Refresh time for the topic map defaults to `60s`

### Kubernetes

The project contains kubernetes files needed to deploy the OpenFaaS Rabbit MQ Connector. They are configured out of the box
to work with the default setup of OpenFaaS. If your setup is customize you might want to adjust some values.

## Contribution

> Follows

## Bug Reporting & Feature Requests

Please feel free to report any issues or Feature request on the [Issue Tab](https://github.com/Templum/rabbitmq-connector/issues). 