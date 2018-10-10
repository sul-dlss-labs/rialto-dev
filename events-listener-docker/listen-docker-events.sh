#!/bin/bash

echo "Listening"

docker events --filter 'event=create' --filter 'type=container' --filter 'image=lambci/lambda:go1.x' --format '{{.Actor.Attributes.name}}' | while read container_name

do
    echo "$container_name was created"

    # Connect any container whose name contains localstack_lambda_ to the local_default network
    docker network connect rialto-dev_rialto $container_name
    echo "The container $container_name is now connected to the rialto-dev_rialto network"
done
