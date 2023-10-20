# Prepare your Kubernetes Cluster

Once you've got your Kubernetes cluster up and running, we need to put in some additional components to support our scenarios. To carry out these steps, you'll need

* [Kubectl](https://kubernetes.io/docs/reference/kubectl/overview/)
* [Helm 3](https://helm.sh/)

## 1. Deploy cert manager

Symphony uses [cert-manager](https://cert-manager.io/docs/installation/kubernetes/) to simplfy certifate management tasks. Use ```kubectl``` to deploy cert manager:

```bash
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.4.0/cert-manager.yaml
```

## 2. Deploy Azure IoT Edge on Kubernetes (optional)

Use the steps in this section if you're running K8s for your Kubernetes cluster.

(See official docs [here](https://microsoft.github.io/iotedge-k8s-doc/introduction.html))

```bash
kubectl create ns iotedge
helm install --repo https://edgek8s.blob.core.windows.net/staging edge-crd edge-kubernetes-crd  
export connStr='<Your IoT Hub device connection string>'
helm install --repo https://edgek8s.blob.core.windows.net/staging edge1 edge-kubernetes --namespace iotedge --set "provisioning.deviceConnectionString=$connStr"
```

> **NOTE:** The IoT Hub device connection string is the connection string of the IoT Edge device you've created previously. To get the connection string, use the following command:
>
>```azurecli
>az iot hub device-identity connection-string show --device-id <device name> --hub-name <iot hub name>
>```

## 3. Install Hierarchical Namespaces controller and Kubectl HNS plugin (optional)

Use the steps in this section if you're using hierarchical namespaces.

```bash
# Select the latest version of HNC
HNC_VERSION=v1.0.0

# Install HNC. Afterwards, wait up to 30s for HNC to refresh the certificates on its webhooks.
kubectl apply -f https://github.com/kubernetes-sigs/hierarchical-namespaces/releases/download/${HNC_VERSION}/default.yaml 

HNC_VERSION=v1.0.0
HNC_PLATFORM=linux_amd64 # also supported: darwin_amd64, darwin_arm64, windows_amd64
curl -L https://github.com/kubernetes-sigs/hierarchical-namespaces/releases/download/${HNC_VERSION}/kubectl-hns_${HNC_PLATFORM} -o ./kubectl-hns
chmod +x ./kubectl-hns

# Ensure the plugin is working
kubectl hns
# The help text should be displayed
```

For more information about installing hierarchical namespaces, see [Install HNC v1.0.0](https://github.com/kubernetes-sigs/hierarchical-namespaces/releases/tag/v1.0.0).

## Next steps

Now, you're ready to [deploy Symphony](./deploy.md).
