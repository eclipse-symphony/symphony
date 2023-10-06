# Symphony Docs

_(last edit: 9/18/2023)_

Welcome to Symphony!

<div align="center">
  <img src="./symphony-book/images/symphony.png" alt="Symphony" width="200" height="200">
</div>


Have you encountered the complexities of manually orchestrating multiple toolchains to oversee your edge infrastructure and solutions? Symphony orchestrates existing services and tools to form an end-to-end, consistent intelligent edge experience. Symphony unites various device classes, software artifacts, and service toolchains, seamlessly harmonizing them into a unified system. With Symphony, you become the maestro, holding the complete vision and exerting precise control while enabling each component to realize its full potential, regardless of whether it's on heavy edge, light edge, or tiny edge, and irrespective of whether it's running on Kubernetes, in the cloud, or on-premises services.

## Getting Started

You can start using Symphony in minutes by selecting one of the following paths:

* [Using Symphony cross-platform CLI (Maestro)](./symphony-book/quick_start/quick_start_maestro.md)

* [Using Symphony Docker container](./symphony-book/quick_start/quick_start_docker.md)

* [Using Symphony Helm Chart on an existing Kubernetes cluster](./symphony-book/quick_start/quick_start_helm.md)

* [Running Symphony as a self-contained binary](./symphony-book/quick_start/quick_start_binary.md)

## Quick Start Scenarios

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requries Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploying a Prometheus server to a K8s cluster](./symphony-book/quick_start/deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploying a Redis container with standalone Symphony](./symphony-book/quick_start/deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploying a simulated temperature sensor Solution to an Azure IoT Edge device](./symphony-book/quick_start/deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |

## Concepts

* [Overview](./symphony-book/concepts/overview.md)
* [Information Graph](./symphony-book/concepts/information_graph.md)
* [State Seeking](./symphony-book/concepts/state_seeking.md)
* [Workflows](./symphony-book/concepts/workflows.md)
* [App Orchestration Model](./symphony-book/concepts/orchestration_model.md)

## Using Symphony

* [Modeling Applications](./symphony-book/solution-management/solution-management.md)

## Advanced Scenarios

* [Canary Deployment](./symphony-book/scenarios/canary-deployment.md)

## Contributing to Symphony

* [Developer Guide](./symphony-book/dev-guide/getting-started.md)
* [API Reference](./symphony-book/api/api.md)
* Extending Symphony

## Additional Topics

* [Symphony Portal](./symphony-book/portals/overview.md)
* [Symphony Expression](./symphony-book/uom/property-expressions.md)