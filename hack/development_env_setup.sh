#! /bin/sh

# Check the availability of required tools

if ![ -x "$(command -v docker)" ]; then
    echo "Missing Docker, please install it"
    exit 1
fi

# Spawn Rabbit MQ

substr="dev_rabbitmq"
if [[ $(docker ps --filter name=dev_rabbitmq) =~ $substr ]]; then
    docker run -p 5672:5672 -e RABBITMQ_DEFAULT_USER="user" -e RABBITMQ_DEFAULT_PASS="pass" -d --restart always --name dev_rabbitmq rabbitmq:3.7.4
    echo "Started Rabbit MQ with the following config:"
    echo "Credentials: [Username: user Password: pass]"
    echo "Port: 5672"
else
    echo "Rabbit MQ is already running"
fi
