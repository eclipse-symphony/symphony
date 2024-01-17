# Build and deploy

> **NOTE:**  As we optimize build and deployment process, this doc will be updated frequently to reflect latest changes. When you start fresh, we recommend you use the latest code and instructions.

## 1. Build (optional)

To begin, follow the steps to [build Symphony containers](./build.md).

If you don't want to build the containers yourself, you can use these prebuilt images:

* ghcr.io/azure/symphony/symphony-agent:latest (or use a release tag. The latest is 0.38.2)
* ghcr.io/azure/symphony/symphony-api:latest (or use a release tag. The latest is 0.38.2)
* ghcr.io/azure/symphony/symphony-k8s:latest (or use a release tag. The latest is 0.38.2)

> **NOTE:** These images may get removed at any time without prior notice.

## 2. Prepare Azure resources

Next, follow the steps to [prepare Azure resources](./prepare_azure.md) in your Azure environment for running test scenarios.

## 3. Create a Kubernetes cluster (as needed)

Next, create a Kuberentes cluster with one of the following options:

* [AKS](./aks.md)
* [Kind](./kind.md)
* [MicroK8s](./microk8s.md)
* [K3s](./k3s.md)

## 4. Prepare your Kubernetes cluster

Next, follow the steps to [prepare your Kubernetes cluster](./prepare_k8s.md) for Azure IoT Hub, Azure Video Analyzer and Azure Arc.

## 5. Deploy Symphony

Finally, follow the steps to deploy Symphony:

* [Deploy Symphony to a single site](./deploy.md).
* [Deploy Symphony to multiple sites](./multisite-deploy.md)
