#!/usr/bin/env bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

cd test/localenv
mage cluster:up

# Get the server CA certificate
kubectl get secret -n default symphony-api-serving-cert  -o jsonpath="{['data']['ca\.crt']}" | base64 --decode > localCA.crt
sudo cp localCA.crt /usr/local/share/ca-certificates/localCA.crt
sudo update-ca-certificates
ls -l /etc/ssl/certs | grep localCA

# config client CA and subjects in values.yaml and use the client cert sample in sample folder
# add symphony-service to DNS mapping
sudo vi /etc/hosts
# add the following line
# 127.0.0.1 symphony-service

# create the remote target
kubectl apply -f ../../remote-agent/bootstrap/sample_target.yaml
# send agent bootstrap request
curl --cert ../../remote-agent/bootstrap/sample/client.pem --key ../../remote-agent/bootstrap/sample/key.pem -X POST "https://symphony-service:8081/v1alpha2/targets/bootstrap/remote-target?namespace=default&osPlatform=linux"
