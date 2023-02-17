# Create a Kind Cluster
## WSL2 on Windows

### 1. Setting up Kind
To intall Kind on Ubuntu WSL (or other Linux distributions):
```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.15.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

To install Kind on Windows 11:
```powershell
# Install choco if needed
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
# Install Docker Desktop if needed
choco install docker-desktop -y
# Install Go if needed
choco install go -y
# Install Kind
choco install kind -y
# Install Helm if needed
choco install kubernetes-helm
```
### 2. Creating a new cluster
Create a kind cluster:
```bash
kind create cluster --name symphony
# Use --image switch to select a different node image verison if needed, such as kindest/node:v1.21.1
# Use --name switch to choose a different name if desired.
kubectl config use-context kind-symphony
```

> **NOTE:** Akri doesn't seem to work on Kind.