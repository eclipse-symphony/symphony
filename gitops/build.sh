#!/bin/bash

# Build the project
echo "Building the project..."
go mod vendor
docker build -t symphonycr.azurecr.io/gitops-service:$1 .
rm -rf vendor