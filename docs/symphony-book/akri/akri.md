# Akri integration

## Build

Symphony Akri custom broker container can be built from the `symphony-k8s` repository:

```bash
cd ../../samples/akri-discover-job/
# build binary
go build -o discover-job
# build container
docker build -t ghcr.io/azure/symphony/symphony-akri:<version> -f ./Dockerfile.microsoft .
# push container
docker push ghcr.io/azure/symphony/symphony-akri:<version>
```

## Configure Akri on a K8s cluster

To test with ONVIF cameras on your network, you need a local Kubernetes cluster (instead of AKS) such as [K3s](../build_deployment/k3s.md) or [MicroK8s](../build_deployment/microk8s.md). At the time of writing, Akri doesn't seem to work on **Kind**.

```bash
export AKRI_HELM_CRICTL_CONFIGURATION="--set kubernetesDistro=k8s"
helm repo add akri-helm-charts https://project-akri.github.io/akri/
helm template akri akri-helm-charts/akri \
    --set onvif.configuration.enabled=true \
    --set onvif.discovery.enabled=true
    --set onvif.configuration.brokerPod.image.repository="ghcr.io/azure/symphony/symphony-akri" \
    --set rbac.enabled=false \
    --set controller.enabled=false \
    --set agent.enabled=false \ 
    --set onvif.configuration.brokerPod.image.tag="0.39.9" \
    --set onvif.configuration.brokerProperties.AKRI_CONFIG_NAME="akri-onvif" \
    --set onvif.configuration.brokerProperties.AKRI_INSTANCE_NAMESPACE="default" \
    --set onvif.configuration.brokerProperties.AKRI_DEVICE_TYPE="camera" \
    --set onvif.configuration.brokerProperties.IN_CLUSTER="true" > configuration.yaml
```
