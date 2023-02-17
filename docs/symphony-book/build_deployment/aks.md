# Create an AKS cluster
There are no special requirements to run Symphony on a AKS cluster. Simplyg follow [official instructions](https://docs.microsoft.com/en-us/azure/aks/kubernetes-walkthrough) to create a new AKS cluster.

To enable pulling images from an Azure Container Registry, use the following command to update your AKS cluster:
```bash
az aks update -n myAKSCluster -g myResourceGroup --attach-acr <acr-name>
# For example
# az aks update -n k8s -g symphony --attach-acr symphonyk8s
```