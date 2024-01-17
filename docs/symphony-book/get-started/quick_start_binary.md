# Use Symphony as a binary

_(last edit: 9/18/2023)_

Symphony can run as a standalone binary with zero external dependencies. You can build Symphony API with [Go](https://go.dev/) (1.20 or above).

Clone the Symphony repo:

```bash
git clone https://github.com/azure/symphony
```

Build the Symphony API binary:

```bash
# assume you are under the repo root folder
cd api
go build -o ./symphony-api
```

Launch Symphony with a [configuration file](../hosts/overview.md):

```bash
./symphony-api -c ./symphony-api-no-k8s.json -l Debug
```

## More topics

Now that you have the Symphony API, try out one of the quick start scenarios:

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requires Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploy a Prometheus server to a K8s cluster](deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploy a Redis container with standalone Symphony](deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploy a simulated temperature sensor Solution to an Azure IoT Edge device](deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |
| [Manage RTSP cameras connected to a gateway](symphony-book/get-started/manage_rtsp_cameras.md) | **Yes** | **Yes** | - | - | **Yes** |
