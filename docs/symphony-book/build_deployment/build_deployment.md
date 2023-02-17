# Chapter 2: Build & Deployment

> **NOTE:**  As we optimize build and deployment process, this doc will be updated frequently to reflect latest changes. When you start fresh, we recommend you use the latest code and instructions.

## 1. Build (optional)

Plese follow [these steps](./build.md) to build Symphony K8s container and Symphony API container.

If you don't want to build the containers yourself, you can use these prebuilt images:

* possprod.azurecr.io/symphony-agent:latest (or use a release tag. The latest is 0.38.2)
* possprod.azurecr.io/symphony-api:latest (or use a release tag. The latest is 0.38.2)
* possprod.azurecr.io/symphony-k8s:latest (or use a release tag. The latest is 0.38.2)

> **NOTE:** These images may get removed at any time without prior notice.

## 2. Prepare Azure resources
Please follow [these steps](./prepare_azure.md) to prepare your Azure environment for running test scenarios.

## 3. Create a Kubernetes cluster (as needed)

Please follow one of the follwoing links to create a Kuberentes cluster:

* [AKS](./aks.md)
* [Kind](./kind.md)
* [MicroK8s](./microk8s.md)
* [K3s](./k3s.md)

## 4. Prepare your Kubernetes cluster
Please follow [these steps](./prepare_k8s.md) to prepare your Kubernetes cluster for Azure IoT Hub, Azure Video Analyzer and Azure Arc.

## 5. Deploy Symphony
Now, follw [these steps](./deploy.md) to deploy Symphony.