# Use Symphony on Kubernetes clusters with Helm

_(last edit: 9/18/2023)_

You can install Symphony on your Kubernetes clusters using Helm (Helm 3 is required):

```bash
helm install symphony oci://possprod.azurecr.io/helm/symphony --version 0.44.6
```

Or, if you already have the **symphony** repository cloned:

```bash
cd <path>/symphony/.azure/symphony-extension/helm
helm install symphony ./symphony
```

If you need to install the Helm chart from a private ACR like ```possprod.azurecr.io```, you need to log in first:

```bash
# login as necessary. Note once the repo is turned public no authentication is needed
export HELM_EXPERIMENTAL_OCI=1
USER_NAME="00000000-0000-0000-0000-000000000000"
PASSWORD=$(az acr login --name possprod --expose-token --output tsv --query accessToken)
helm registry login possprod.azurecr.io   --username $USER_NAME --password $PASSWORD

# install using Helm chart
helm install symphony oci://possprod.azurecr.io/helm/symphony --version 0.40.8
```

## More topics

Now that you have the Symphony API, try out one of the quick start scenarios:

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requries Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploy a Prometheus server to a K8s cluster](./symphony-book/quick_start/deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploy a Redis container with standalone Symphony](./symphony-book/quick_start/deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploy a simulated temperature sensor Solution to an Azure IoT Edge device](./symphony-book/quick_start/deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |
| [Manage RTSP cameras connected to a gateway](./symphony-book/quick_start/manage_rtsp_cameras.md) | **Yes** | **Yes** | - | - | **Yes** |