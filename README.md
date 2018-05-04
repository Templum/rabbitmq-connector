# OpenFaas Rabbit MQ Connector

This connector allows to receive messages from an Rabbit MQ for defined topics (Routing Keys).
And will invoke functions on the OpenFaas platform that listen on that topic. Currently the
result will be ignored, however in later version it will send the response back on an defined
topic (Routing Key).


## Build it

You can create an local docker image using `$ docker build -t yourname .`

## Try it

> Follows...