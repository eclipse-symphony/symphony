# Create an AKS cluster

There are no special requirements to run Symphony on a AKS cluster. Follow the instructions in [Deploy an Azure Kubernetes Service (ASK) cluster](https://docs.microsoft.com/azure/aks/kubernetes-walkthrough).

To enable pulling images from an Azure Container Registry, use the following command to update your AKS cluster:

```bash
az aks update -n myAKSCluster -g myResourceGroup --attach-acr <acr-name>
# For example
# az aks update -n k8s -g symphony --attach-acr symphonyk8s
```

## Next steps

Next, [prepare your Kubernetes cluster](./prepare_k8s.md) for Azure IoT Hub, Azure Video Analyzer and Azure Arc.
