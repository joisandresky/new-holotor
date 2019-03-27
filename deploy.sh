#!/bin/bash

docker stop holotor-service
docker rm holotor-service
docker build -t "$0" .
docker run -it -d -p 8989:8989 --name holotor-service --restart on-failure "$0"