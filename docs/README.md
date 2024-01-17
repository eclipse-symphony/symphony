# Symphony docs

_(last edit: 9/18/2023)_

Welcome to Symphony!

<div align="center">
  <img src="./symphony-book/images/symphony.png" alt="Symphony" width="200" height="200">
</div>

Have you encountered the complexities of manually orchestrating multiple toolchains to oversee your edge infrastructure and solutions? Symphony orchestrates existing services and tools to form an end-to-end, consistent intelligent edge experience. Symphony unites various device classes, software artifacts, and service toolchains, seamlessly harmonizing them into a unified system. With Symphony, you become the maestro, holding the complete vision and exerting precise control while enabling each component to realize its full potential, regardless of whether it's on heavy edge, light edge, or tiny edge, and irrespective of whether it's running on Kubernetes, in the cloud, or on-premises services.

## Features

To enable your edge solution orchestration, Symphony:

* Projects resources as Kubernetes custom resources, allowing the control plane to be managed using K8s native tools such as kubectl.
* Supports different application model formats, including Helm charts, and a ModuleGroup format that is designed to group multiple Azure IoT Edge modules.
* Supports hardware-accelerated AI payloads using media pipelines through Live AI, OpenVINO, etc.
* Shares the same solution management, security management, and device management logic to ensure consistent behavior across cloud and edge.
* Works well with AKS, Arc, IoT Hub and other Azure services.
* Supports dynamic device discovery and update through Akri.
* Supports additional computational nodes such as Azure Sphere through Virtual Kubelet.
* Supports end-to-end observability with distributed tracing using OpenTelemetry.

## Key application scenarios

Symphony aims to deploy and manage secured, hardware-accelerated intelligent edge payloads on a K8s-based fabric that offers adaptive workload scheduling for HA and resource balancing. Some key application scenarios include:

* Manage intelligent payloads on a highly available field gateway cluster, such as an HCI cluster, that manages attached sensors like brown-field cameras.
* Manage intelligent payloads on a in-vehicle cluster for smart cars, construction vehicles, and/or airplanes.
* Provide business continuity for occasionally connected or disconnected environments.
* Manage intelligent payloads on a fully virtualized environments for large-scale scenarios such as simulation and testing.

## Get started

First, install Symphony on your device using one of the following methods:

* [Use Symphony with the Maestro cross-platform CLI tool](./symphony-book/get-started/quick_start_maestro.md)

* [Use Symphony on a Kubernetes cluster with Helm](./symphony-book/get-started/quick_start_helm.md)

* [Use Symphony as a self-contained binary](./symphony-book/get-started/quick_start_binary.md)

Then, try one of the quickstart scenarios that use Symphony to deploy a sample solution:

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requires Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploy a Prometheus server to a K8s cluster](./symphony-book/get-started/deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploy a Redis container with standalone Symphony](./symphony-book/get-started/deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploy a simulated temperature sensor Solution to an Azure IoT Edge device](./symphony-book/get-started/deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |
| [Manage RTSP cameras connected to a gateway](./symphony-book/get-started/manage_rtsp_cameras.md) | **Yes** | **Yes** | - | - | **Yes** |

## Concepts

* [Overview](./symphony-book/concepts/_overview.md)
* [Information graph](./symphony-book/concepts/information_graph.md)
* [State seeking](./symphony-book/concepts/state_seeking.md)
* [Workflows](./symphony-book/concepts/workflows.md)
* [App orchestration model](./symphony-book/concepts/orchestration_model.md)

## Use Symphony

* [Deploy Symphony to a single site](./symphony-book/build_deployment/deploy.md)
* [Deploy Symphony to multiple sites](./symphony-book/build_deployment/multisite-deploy.md)
* [Troubleshoot](./symphony-book/dev-guide/troubleshooting.md)
* [Model applications](./symphony-book/concepts/unified-object-model/solution.md)

## Advanced scenarios

* [Canary deployment](./symphony-book/scenarios/canary-deployment.md)
* [Multi-site deployment](./symphony-book/scenarios/multisite-deployment.md)

## Contribute to Symphony

* [Developer guide](./symphony-book/dev-guide/getting-started.md)
* [API reference](./symphony-book/api/_overview.md)

## Additional topics

* [Symphony agent](./symphony-book/agent/_overview.md)
* [Symphony portal](./symphony-book/portals/overview.md)
* [Symphony expressions](./symphony-book/concepts/unified-object-model/property-expressions.md)

## Supported technologies

Symphony is platform neutral and protocol neutral. Through its extensible architecture, Symphony supports a broad range of AI frameworks, devices, hardware, services and many more. The following table summarizes some of the supported technologies and the list grows rapidly.

| Area | Supported Technologies |
|--------|--------|
| AI Pipeline | ![ONNX](./symphony-book/images/onnx.png) ![OpenCV](./symphony-book/images/opencv.png), DeepStream* |
| Application Model | ![Helm](./symphony-book/images/helm.png), Symphony, Radius*, ARM* |
| Device Updates | ![GitOps](./symphony-book/images/gitops.png) ![Flux](./symphony-book/images/flux.png)\*, ADU for IoT Hub, ![Arc](./symphony-book/images/arc.png) ![pyOCD](./symphony-book/images/pyocd.png)\*|
| Discovery | ![ONVIF](./symphony-book/images/onvif.png) ![OPC UA](./symphony-book/images/opcua.png)\*, udev (via Akri) |
| Hardware | Azure Stack HCI, MIMXRT1170-EVK, Nvidia Jeston Orin, Nvidia dGPU |
| K8s Distributions | ![Kubernetes](./symphony-book/images/k8s.png) ![Azure Kubernetes Service](./symphony-book/images/aks.png) ![MicroK8s](./symphony-book/images/microk8s.png) ![Kind](./symphony-book/images/kind.png) ![K3s](./symphony-book/images/k3s.png) AKS-IoT |
| Observability | ![Open Telemetry](./symphony-book/images/open-telemetry.png), Azure Monitor |
| OS | ![Ubuntu](./symphony-book/images/ubuntu.png) ![Windows 10](./symphony-book/images/windows.png) ![MacOS](./symphony-book/images/macos.png), CBL-Mariner, Azure RTOS |
| Other Azure Service | Azure Storage, Azure Logic Apps, Azure Functions |
| Policies | Gatekeeper, ![Kyverno](./symphony-book/images/kyverno.png) |
| Runtime | Azure IoT Edge, Kubernetes, Windows 10, Samsung DERAM* |
| Scripting | Bash, ![PowerShell](./symphony-book/images/powershell.png), Windows Batch |

_*:upcoming_