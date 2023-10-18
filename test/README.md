# Local environment setup

Setup your local environment to run unit tests and integration tests.

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
curl -LO "https://go.dev/dl/go1.21.3.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version

mkdir -p $HOME/go/bin

```

### helm installation

https://helm.sh/docs/intro/install/#from-the-binary-releases

Example:

```shell
wget https://get.helm.sh/helm-v3.9.3-linux-amd64.tar.gz
tar xvf helm-v3.9.3-linux-amd64.tar.gz
sudo mv linux-amd64/helm /usr/local/bin
rm helm-v3.4.1-linux-amd64.tar.gz
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
```

### mage installation

Example:

https://github.com/magefile/mage#installation

```shell
git clone https://github.com/magefile/mage
cd mage
go run bootstrap.go
```

## Perquisites for running unit tests

In [k8s](../k8s) and [api](../api) folder, we could run unit tests.

To run unit tests, we need to

```shell
# install gcc
sudo apt-get update
sudo apt-get install build-essential

# configure go env - CGO_ENABLED 
go env -w CGO_ENABLED=1
go env CGO_ENABLED
```

Then go to [k8s](../k8s) or [api](../api), run ```mage test``` to launch unit tests.