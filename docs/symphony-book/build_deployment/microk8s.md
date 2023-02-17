# Create a MicroK8s Cluster

## WSL2 on Windows
### 1. Setting up MicroK8s

1. Install MicroK8s
   ```bash
   sudo snap install microk8s --classic --channel=1.19/stable
   ```
2. Allow your user to run MicroK8s commands
   ```bash
   sudo usermod -a -G microk8s $USER
   sudo chown -f -R $USER ~/.kube
   su - $USER
   ```
3. Check if MicroK8s is running
   ```bash
   microk8s status --wait-ready
   ```
4. Enable serveral features
   ```bash
   microk8s enable dns helm3 rbac
   ```
5. Enable pods to run in priviledged context.
   ```bash
   echo "--allow-privileged=true" >> /var/snap/microk8s/current/args/kube-apiserver
   microk8s.stop
   microk8s.start
   ```
> **NOTE:** See instructions [here](https://ubuntu.com/tutorials/install-microk8s-on-windows#2-installation) to install MicroK8s on Windows.

### 2. Install crictl (optional - for Akri)
Akri depends on crictl to tarck pod information. Follow these commands to install and config crictl for Akri.
   ```bash
   VERSION="v1.17.0"
   curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/$VERSION/crictl-${VERSION}-linux-amd64.tar.gz --output crictl-${VERSION}-linux-amd64.tar.gz
   sudo tar zxvf crictl-$VERSION-linux-amd64.tar.gz -C /usr/local/bin
   rm -f crictl-$VERSION-linux-amd64.tar.gz
   
   export AKRI_HELM_CRICTL_CONFIGURATION="--set agent.host.crictl=/usr/local/bin/crictl --set agent.host.dockerShimSock=/var/snap/microk8s/common/run/containerd.sock"
   ```
### 3. Create command alias
To use ```kubectl``` or ```helm``` commands against a MicroK8s cluster, you need to put a ```microk8s``` prefix, such as ```microk8s kubectl get nodes```. You should create command alias before trying to deploy Symphony.

```bash
alias kubectl='microk8s kubectl'
alias helm='microk8s helm3'
```
### 4. Enable insecured API port
When you deploy Symphony, the make file tries to access Kubernetes API through insecured port (8080), which is disabled by default. To enable this, edit ```/var/snap/microk8s/current/args/kube-apiserver``` and change ```--insecure-port``` from ```0``` to ```8080```. Then, restart MicroK8s with:
```bash
microk8s.stop
microk8s.start
```
### 5. (Optional) Connect to Arc
Microk8s doesn't create/update Kubernetes configuration files by default. But Arc extension expects such configurations exist. To fix this, you can export Microk8s config into a regular Kubernetes configuration file:
```bash
microk8s config view > ~/.kube/config
```
To connect your cluster to Arc:
```bash
az connectedk8s connect --name <Arc registered cluster name> --resource-group <resource group name> --kube-context microk8s
To create Microsoft.Symphony extension:

# As needed
az extension add --name connectedk8s
az extension add --name k8s-extension

# TO update
# az extension update --name connectedk8s
# az extension update --name k8s-extension

az provider register --namespace Microsoft.KubernetesConfiguration
az feature register --namespace Microsoft.KubernetesConfiguration --name extensions

# Create extension instance
az k8s-extension create --cluster-type connectedClusters --cluster-name arc-microk8s --resource-group symphony-review --name symphony-1 --extension-type Microsoft.Symphony
```
> **NOTE:** To use Arc extension under DogFood enviornment, see instructions here.

### 6. To clean up
To reset your MicroK8s cluster, use:
```bash
microk8s.reset
```
To uninstall Microk8s, use:
```bash
sudo snap remove microk8s
```
And to remove Arc registration, use:
```bash
az connectedk8s delete  --resource-group <resource group name> --name <Arc registered cluster name>
```
Once the cluster is reset, you can re-enabled the addons (see step 1.4 above), re-deploy Symphony and start over.

## Mac Book

### 1. Install Microk8s
Follow instructions [here](https://ubuntu.com/tutorials/install-microk8s-on-mac-os#1-overview) to install Microk8s. This uses [Multipass](https://multipass.run/docs/installing-on-macos) to create a VM on your Mac.

### 2. Helm support
Microk8s's Helm support can be enabled by:
```bash
microk8s enable helm3
```
> **NOTE**: Helm versions included in Microk8s is not upgradable. At the time of writing, the latest Microk8s packages with Helm 3.5, which doesn't support ```oci://``` repositories.

In order to use ```helm install``` against your local Helm chart, you need to mount the folder to your multipass instance, such as:

```bash
multipass mount <full path on your host> <multipass instance name>
```
Once you mount the folder, you can access the host folder from you multipass instance using the same folder path as you have on the host. Now, you can use ```microk8s helm3 install symphony <full path on your host>``` to install Symphony Helm chart.

> **NOTE**: Use ```multipass list``` to list your VMs; and use ```multipass shell <vm name>``` to open a shell on your VM.