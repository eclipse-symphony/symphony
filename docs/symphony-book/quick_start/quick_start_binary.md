# Symphony Quick Start with binary

_(last edit: 9/18/2023)_

Symphony can run as a standalone binary with zero external depedencies. You can easily build Symphony API with [Go](https://go.dev/) (1.20 or above):

```bash
# assume you are under the repo root folder
cd api
go build -o ./symphony-api
```

Then, you can launch Symphony with a [configuration file](../hosts/overview.md):

```bash
./symphony-api -c ./symphony-api-no-k8s.json -l Debug
```

## More Topics

Try out one of the quick start scenario

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requries Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploying a Prometheus server to a K8s cluster](./deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploying a Redis container with standalone Symphony](./deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploying a simulated temperature sensor Solution to an Azure IoT Edge device](./deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |