<!--
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
SPDX-License-Identifier: MIT
-->
# Local environment

The local environment is a minikube cluster that deploys symphony for testing purposes.

## Prerequisites

Linux environment such as Azure Ubuntu VM or WSL Ubuntu Distribution.

### Docker installation

https://docs.docker.com/engine/install/ubuntu/

Example:
```shell
# Docker installation

# Add Docker's official GPG key:
sudo apt-get update
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Add the repository to Apt sources:
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update

sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

### go installation
https://go.dev/doc/install

Example:

```shell
# Go installation
curl -LO "https://go.dev/dl/go1.22.4.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
rm go1.22.4.linux-amd64.tar.gz
mkdir -p $HOME/go/bin

```

use `vim ~/.bash_profile` to open the configuration file, append following command:
```shell
export GOPATH=~/go
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```
run `source ~/.bash_profile` to finish the environment variable setting.

### helm installation

https://helm.sh/docs/intro/install/#from-the-binary-releases

Example:

```shell
wget https://get.helm.sh/helm-v3.9.3-linux-amd64.tar.gz
tar xvf helm-v3.9.3-linux-amd64.tar.gz
sudo mv linux-amd64/helm /usr/local/bin
rm helm-v3.9.3-linux-amd64.tar.gz
rm -rf linux-amd64
helm version
```

### kubectl installation
https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#install-kubectl-binary-with-curl-on-linux

Example:

```shell
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl.sha256"
echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check # valid output: kubectl: OK

sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
kubectl version --client
rm -rf kubectl kubectl.sha256
```

### mage installation

Example:

https://github.com/magefile/mage#installation

(Please make sure you set go environment variable correctly in previous steps.)

```shell
git clone https://github.com/magefile/mage
cd mage
go run bootstrap.go
```

## Usage of mage

See all commands with `mage -l`

```
> git clone https://github.com/eclipse-symphony/symphony
> cd symphony/test/localenv
> mage -l

Use this tool to quickly get started developing in the symphony ecosystem. The tool provides a set of common commands to make development easier for the team. To get started using Minikube, run 'mage build minikube:start minikube:load deploy'.

Targets:
  acrLogin              Log into the ACR, prompt if az creds are expired
  build:all             Build builds all containers
  build:api             Build api container
  build:k8s             Build k8s container
  buildUp               builds the latest images and starts the local environment
  cluster:deploy        Deploys the symphony ecosystem to your local Minikube cluster.
  cluster:down          Stop the cluster
  cluster:load          Brings the cluster up with all images loaded
  cluster:status        Show the state of the cluster for CI scenarios
  cluster:up            Brings the cluster up and deploys
  destroy               Uninstall all components
  minikube:dashboard    Launch the Minikube Kubernetes dashboard.
  minikube:delete       Deletes the Minikube cluster from you dev box.
  minikube:install      Installs the Minikube binary on your machine.
  minikube:load         Loads symphony component docker images onto the Minikube VM.
  minikube:start        Starts the Minikube cluster w/ select addons.
  minikube:stop         Stops the Minikube cluster.
  pull:all              Pulls all docker images for symphony
  pull:api              Pull symphony-api
  pull:k8s              Pull symphony-k8s
  pullUp                pulls the latest images and starts the local environment
  test:up               Deploys the symphony ecosystem to minikube and waits for all pods to be ready.
  up                    brings the minikube cluster up with symphony deployed
```


# Getting started

Use the `Up` commands to start the local environment and deploy symphony. If it is your first time running the environment and you do not have local images you will need to either pull or build them.

```bash
# Start the cluster and deploy symphony using the images on your dev box
mage Up

# Pull images from ACR and start the cluster
mage PullUp

# Build images from source and start the cluster
mage BuildUp

# View the pods running in the cluster
mage cluster:status
```

For working the cluster use [k9s](https://github.com/derailed/k9s) or `kubectl`
# Local development

For a typical development workflow you can build the image or images you are modifying, then deploy them to the cluster before testing and applying custom resources.

```bash
# Build the image you are working on or use build:all
mage build:k8s

# Deploy to the cluster
mage up
```

You can also run the deployment manually


```bash
# build first
mage build:all

# run the cluster in the background
mage cluster:up

# deploy symphony
mage cluster:deploy
```

To remove symphony from the cluster use

```
mage Destroy all,nowait
```

If packages version in go.mod is updated, prepare dependencies using
```
go mod tidy
```

# Troubleshooting

If you are seeing strange behavior or getting errors the first thing to try is completely deleting minikube and starting over with a fresh cluster. Many commands will recreate minikube for you automatically, but it is worth checking that minikube is actually getting cleaned up.

```bash
minikube delete
```


# Integration tests

CI integration tests can be run locally with the following command:

```
mage test:up
```

See [integration test README](../integration/README.md) for more details.

# Unit tests

## Perquisites for running unit tests

In [k8s](../../k8s) and [api](../../api) folder, we could run unit tests.

To run unit tests, we need to

```shell
# install gcc
sudo apt-get update
sudo apt-get install build-essential

# install jq
sudo apt-get install jq

# configure go env - CGO_ENABLED 
go env -w CGO_ENABLED=1
go env CGO_ENABLED
```
Then go to [k8s](../../k8s) or [api](../../api) folder, run ```mage test``` to launch unit tests.