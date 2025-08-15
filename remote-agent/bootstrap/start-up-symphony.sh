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
# openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256 -extfile openssl.cnf -extensions v3_req
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256
# verify the client certificate
openssl verify -CAfile ca.crt client.crt
openssl pkcs12 -export -out client.pfx -inkey client.key -in client.crt -certfile ca.crt
# create a client cert secret: secret name is client-cert-secret, key is client-cert-key, value is client.crt
kubectl create namespace cert-manager
kubectl create secret generic client-cert-secret --from-file=client-cert-key=ca.crt -n cert-manager

# judge if the secret public key is the same as the client.crt
kubectl get secret client-cert-secret -n cert-manager -o jsonpath='{.data.client-cert-key}' | base64 -d > secret-client.crt
diff ca.crt secret-client.crt
if [ $? -ne 0 ]; then
    echo "Error: client.crt and secret public key are different!"
fi

cd ../../test/localenv

mage cluster:deployWithSettings "--set remoteAgent.remoteCert.used=true --set remoteAgent.remoteCert.trustCAs.secretName=<secret name> --set remoteAgent.remoteCert.trustCAs.secretKey=<secret key> --set installServiceExt=true"
# default is : mage cluster:deployWithSettings "--set remoteAgent.remoteCert.used=true --set remoteAgent.remoteCert.trustCAs.secretName=client-cert-secret --set remoteAgent.remoteCert.trustCAs.secretKey=client-cert-key --set installServiceExt=true"

# start a new terminal
# minikube tunnel

# remove the localCA.crt from the system (optional)
sudo rm /etc/ssl/certs/localCA.pem
sudo rm /etc/ssl/certs/8ce967e1.0
echo "localCA.crt removed from the certificate store."

max_wait=300
waited=0
while ! kubectl get secret -n default symphony-api-serving-cert >/dev/null 2>&1; do
  if [ $waited -ge $max_wait ]; then
    echo "Timeout: symphony-api-serving-cert not found after $max_wait seconds."
    exit 1
  fi
  echo "Waiting for symphony-api-serving-cert to be created..."
  sleep 10
  waited=$((waited + 10))
done

# Get the server CA certificate
kubectl get secret -n default symphony-api-serving-cert  -o jsonpath="{['data']['ca\.crt']}" | base64 --decode > localCA.crt
sudo cp localCA.crt /usr/local/share/ca-certificates/localCA.crt
sudo update-ca-certificates
ls -l /etc/ssl/certs | grep localCA

# config client CA and subjects in values.yaml and use the client cert sample in sample folder
# add symphony-service to DNS mapping
# may not be able to modify host file but to add DNS record
# sudo vi /etc/hosts
# add the following line
# 127.0.0.1 symphony-service

# create the remote target
kubectl apply -f ../../remote-agent/bootstrap/sample_target.yaml
# call the bootstrap.sh script
# sign bootstrap script
# sign binary
sudo systemctl stop remote-agent.service
cd ../../remote-agent/bootstrap
./bootstrap.sh http https://symphony-service:8081/v1alpha2 client.crt client.key remote-demo default topologies.json <user> <group>