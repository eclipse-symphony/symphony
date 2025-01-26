#!/usr/bin/env bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

cd test/localenv
mage cluster:up

# start a new terminal
minikube tunnel

# remove the localCA.crt from the system (optional)
sudo rm /etc/ssl/certs/localCA.pem
sudo rm /etc/ssl/certs/8ce967e1.0
echo "localCA.crt removed from the certificate store."

# Get the server CA certificate
kubectl get secret -n default symphony-api-serving-cert  -o jsonpath="{['data']['ca\.crt']}" | base64 --decode > localCA.crt
sudo cp localCA.crt /usr/local/share/ca-certificates/localCA.crt
sudo update-ca-certificates
ls -l /etc/ssl/certs | grep localCA

# config client CA and subjects in values.yaml and use the client cert sample in sample folder
# add symphony-service to DNS mapping
# may not be able to modify host file but to add DNS record
sudo vi /etc/hosts
# add the following line
# 127.0.0.1 symphony-service

# create the remote target
kubectl apply -f ../../remote-agent/bootstrap/sample_target.yaml
# call the bootstrap.sh script
# sign bootstrap script
# sign binary
./bootstrap.sh https://symphony-service:8081/v1alpha2 ../client-cert.pem ../client-key.pem remote-demo default topologies.json ../config.json