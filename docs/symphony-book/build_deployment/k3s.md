# Create a K3s Cluster

An easy way to set up K3s on a Windows, Mac or Linux PC is by using [Rancher Desktop](https://rancherdesktop.io/), which can be downloaded and installed as a native app.





<hr />

## Appendix
### WSL2 on Windows
#### 1. Setting up K3s
```bash
# Download k3s binary
 wget https://github.com/k3s-io/k3s/releases/download/v1.26.3%2Bk3s1/k3s-arm64
# Rename file
mv k3s-arm64 k3s
# Enable execution
chmod +x k3s
# Move bindary
sudo mv k3s /usr/local/bin
# Luanch K3s server
sudo k3s server
```
#### 2. Configure kubectl
On another terminal:
```bash
# Copy K8s config file
sudo cp /etc/rancher/k3s/k3s.yaml $HOME/.kube
# Add the config to KUBECONFIG variable (edit your shell profile to auo-load)
export KUBECONFIG=$HOME/.kube/config:$HOME/.kube/k3s.yaml
# Switch to the "default" context
kubectl config use-context default
# Display cluster info
kubectl cluster-info
```