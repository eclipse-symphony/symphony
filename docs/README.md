# Symphony docs

_(last edit: 9/18/2023)_

Welcome to Symphony!

<div align="center">
  <img src="./symphony-book/images/symphony.png" alt="Symphony" width="200" height="200">
</div>


Have you encountered the complexities of manually orchestrating multiple toolchains to oversee your edge infrastructure and solutions? Symphony orchestrates existing services and tools to form an end-to-end, consistent intelligent edge experience. Symphony unites various device classes, software artifacts, and service toolchains, seamlessly harmonizing them into a unified system. With Symphony, you become the maestro, holding the complete vision and exerting precise control while enabling each component to realize its full potential, regardless of whether it's on heavy edge, light edge, or tiny edge, and irrespective of whether it's running on Kubernetes, in the cloud, or on-premises services.

## Get started

First, install Symphony on your device using one of the following methods:

* [Use Symphony with the Maestro cross-platform CLI tool](./symphony-book/quick_start/quick_start_maestro.md)

* [Use Symphony on a Kubernetes cluster with Helm](./symphony-book/quick_start/quick_start_helm.md)

* [Use Symphony as a self-contained binary](./symphony-book/quick_start/quick_start_binary.md)

Then, try one of the quickstart scenarios that use Symphony to deploy a sample solution:

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requries Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploy a Prometheus server to a K8s cluster](./symphony-book/quick_start/deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploy a Redis container with standalone Symphony](./symphony-book/quick_start/deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploy a simulated temperature sensor Solution to an Azure IoT Edge device](./symphony-book/quick_start/deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |
| [Manage RTSP cameras connected to a gateway](./symphony-book/quick_start/manage_rtsp_cameras.md) | **Yes** | **Yes** | - | - | **Yes** |

## Concepts

* [Overview](./symphony-book/concepts/overview.md)
* [Information Graph](./symphony-book/concepts/information_graph.md)
* [State Seeking](./symphony-book/concepts/state_seeking.md)
* [Workflows](./symphony-book/concepts/workflows.md)
* [App Orchestration Model](./symphony-book/concepts/orchestration_model.md)

## Use Symphony

* [Deploy Symphony (single site)](./symphony-book/build_deployment/deploy.md)
* [Deploy Symphony (multiple sites)](./symphony-book/build_deployment/multisite-deploy.md)
* [Troubleshoot](./symphony-book/dev-guide/troubleshooting.md)
* [Modeling Applications](./symphony-book/solution-management/solution-management.md)
* [Use Symphony in a Docker container](./symphony-book/quick_start/quick_start_docker.md)

## Advanced Scenarios

* [Canary Deployment](./symphony-book/scenarios/canary-deployment.md)
* [Multisite Deployment](./symphony-book/scenarios/multisite-deployment.md)

## Contributing to Symphony

* [Developer Guide](./symphony-book/dev-guide/getting-started.md)
* [API Reference](./symphony-book/api/api.md)

## Additional Topics

* [Symphony Portal](./symphony-book/portals/overview.md)
* [Symphony Expression](./symphony-book/uom/property-expressions.md)