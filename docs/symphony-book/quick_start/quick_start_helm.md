# Symphony Quick Start with Helm

_(last edit: 9/18/2023)_

You can install Symphony on your Kubernetes clusters using Helm (Helm 3 is required):

```bash
helm install symphony oci://possprod.azurecr.io/helm/symphony --version 0.44.6
```

Or, if you already have the ```symphony``` repository cloned:

```bash
cd symphony-extension/helm
helm install symphony ./symphony
```

## More Topics

Try out one of the quick start scenario

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requries Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploying a Prometheus server to a K8s cluster](./deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploying a Redis container with standalone Symphony](./deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploying a simulated temperature sensor Solution to an Azure IoT Edge device](./deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |


## Appendix

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