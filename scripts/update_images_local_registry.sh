#!/bin/sh
# docker tag [OPTIONS] IMAGE[:TAG] [REGISTRYHOST/][USERNAME/]NAME[:TAG]

set -eu

cd "$(dirname "$0")/.."

echo building nc-server locally
docker build -t ncserver ./nc-server
docker image tag ncserver docker-registry:5000/ncserver
docker push docker-registry:5000/ncserver

echo building db-service locally
docker build -t db-service ./db-service
docker image tag db-service docker-registry:5000/db-service
docker push docker-registry:5000/db-service
