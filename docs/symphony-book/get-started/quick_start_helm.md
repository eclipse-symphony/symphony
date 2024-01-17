# Use Symphony on Kubernetes clusters with Helm

_(last edit: 9/18/2023)_

To install Symphony on your Kubernetes clusters using Helm 3, use `helm install`:

```bash
helm install symphony oci://ghcr.io/azure/symphony/helm/symphony --version 0.44.6 --set imagePullSecrets='{YOUR_GITHUB_PAT_TOKEN}'
```

Or, if you already have the **symphony** repository cloned:

```bash
cd <path>/symphony/packages/helm
helm install symphony ./symphony --set imagePullSecrets='{YOUR_GITHUB_PAT_TOKEN}'
```

If you need to install the Helm chart from a private ACR like ```ghcr.io```, you need to log in first:

```bash
# login as necessary. Note once the repo is turned public no authentication is needed
export HELM_EXPERIMENTAL_OCI=1
USER_NAME="USERNAME"
PASSWORD='{YOUR_GITHUB_PAT_TOKEN}'
helm registry login ghcr.io --username $USER_NAME --password $PASSWORD

# install using Helm chart
helm install symphony oci://ghcr.io/azure/symphony/helm/symphony --version 0.40.8 --set imagePullSecrets='{YOUR_GITHUB_PAT_TOKEN}'
```

## Update Symphony

To update your existing Symphony release to a new version, use `helm upgrade`:

```bash
helm upgrade --install symphony oci://ghcr.io/azure/symphony/helm/symphony --version 0.45.31 --set imagePullSecrets='{YOUR_GITHUB_PAT_TOKEN}'
```

## Customize Helm deployment

Use the following Helm chart value switches to customize your Symphony deployment with Helm (by adding `--set <key>=<value>` switches):

| Switch | Description|
|--------|--------|
| `installServiceExt` | When set to `true` (default), Symphony deploys a publicly accessible `symphony-service-ext` service for agents and child sites. If you don't have such needs, you can turn off this service by setting the value to `false`. |
| `redis.enabled` | When set to `true` (default), Symphony deploys a Redis pod as its pub/sub messaging backbone. If you turn this to `false`, an in-memory backbone is used, which can't be scaled beyond a single API pod. |
| `siteId` | You can change the default site id, which is `hq`, to a different value with this switch. |
| `parent.url` | When this value is set, the current Symphony control plane is linked to a parent Symphony control plane. For more information, see [multi-site Symphony deployment](../build_deployment/multisite-deploy.md). |

## More topics

Now that you have the Symphony API, try out one of the quick start scenarios:

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requires Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploy a Prometheus server to a K8s cluster](deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploy a Redis container with standalone Symphony](deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploy a simulated temperature sensor Solution to an Azure IoT Edge device](deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |
| [Manage RTSP cameras connected to a gateway](symphony-book/get-started/manage_rtsp_cameras.md) | **Yes** | **Yes** | - | - | **Yes** |
