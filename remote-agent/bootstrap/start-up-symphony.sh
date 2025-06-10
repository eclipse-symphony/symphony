#!/usr/bin/env bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##


minikube start
# install openssl
sudo apt update
sudo apt install openssl

# create a local CA
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=MyLocalCA"

# create a client key and CSR
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/CN=target.symphony.microsoft.com"

# use ca to sign 
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256

# verify the client certificate
openssl verify -CAfile ca.crt client.crt

# create a client cert secret: secret name is client-cert-secret, key is client-cert-key, value is client.crt
kubectl create namespace cert-manager
kubectl create secret generic client-cert-secret --from-file=client-cert-key=client.crt -n cert-manager

# judge if the secret public key is the same as the client.crt
kubectl get secret client-cert-secret -n cert-manager -o jsonpath='{.data.client-cert-key}' | base64 -d > secret-client.crt
diff client.crt secret-client.crt
if [ $? -ne 0 ]; then
    echo "Error: client.crt and secret public key are different!"
fi

cd test/localenv

mage cluster:deployWithSettings "--set remoteAgent.used=true --set RemoteCert.ClientCAs.SecretName=<secret name> --set RemoteCert.ClientCAs.SecretKey=<secret key>"
# default is : mage cluster:deployWithSettings "--set remoteAgent.used=true --set RemoteCert.ClientCAs.SecretName=client-cert-secret --set RemoteCert.ClientCAs.SecretKey=client-cert-key"  
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
./bootstrap.sh https://symphony-service:8081/v1alpha2 ../client-cert.pem ../client-key.pem remote-demo default topologies.json <user> <group>