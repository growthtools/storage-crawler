#!/bin/bash

set -e

if [ -z "$1" ]
then
  echo "Please provide a version number, like v1.3.5"
  exit 1
fi

IMAGE_NAME=storage-crawler
IMAGE_PATH=lucid-ceremony-176113/$IMAGE_NAME:$1
IMAGE_URL=gcr.io/$IMAGE_PATH

echo "Compiling binary"
docker run --rm -it -v "$GOPATH":/gopath \
  -v "$(pwd)":/app \
  -e "GOPATH=/gopath" \
  -w /app golang:latest sh \
  -c 'CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s" -o bin/$IMAGE_NAME'

echo "Building image $1"
docker build -t $IMAGE_PATH -f Dockerfile .
docker tag $IMAGE_PATH $IMAGE_URL
docker push $IMAGE_URL
