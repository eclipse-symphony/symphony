# Use Symphony in a Docker container

_(last edit: 9/18/2023)_

You can run the Symphony API as a single Docker container with a configuration file that you mount to the container.

```bash
# assuming you are under the repo root folder
docker run --rm -it -e LOG_LEVEL=Info -v ./api:/configs -e CONFIG=/configs/symphony-api-no-k8s.json possprod.azurecr.io/symphony-api:latest
```

> **Pre-release NOTE**: ```possprod.azurecr.io``` is a private repo. To access the repo, your Azure account needs to be granted access. Then, you need to login to Docker using Azure token: 
>
>```bash
>az login
>TOKEN=$(az acr login --name possprod --expose-token --output tsv --query accessToken)
>docker login possprod.azurecr.io --username 00000000-0000-0000-0000-000000000000 --password $TOKEN
>```

## More topics

Now that you have the Symphony API, try out one of the quick start scenarios:

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requries Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploy a Prometheus server to a K8s cluster](./symphony-book/quick_start/deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploy a Redis container with standalone Symphony](./symphony-book/quick_start/deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploy a simulated temperature sensor Solution to an Azure IoT Edge device](./symphony-book/quick_start/deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |
| [Manage RTSP cameras connected to a gateway](./symphony-book/quick_start/manage_rtsp_cameras.md) | **Yes** | **Yes** | - | - | **Yes** |