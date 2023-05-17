# Chapter 1: Introduction

_(last edit: 4/11/2023)_

Welcome to Symphony! Symphony orchestrates existing services and tools to form an end-to-end, consistent intelligent edge experience. Symphony offers a K8s-native control plane that can be deployed and used on Kubernetes clusters, such as AKS clusters on Azure Stack Edge and K3s clusters on Raspberry Pis. Symphony also runs in a standalone mode as a single process. For more information, please see [Running Symphony in Standalone Mode](../build_deployment/standalone.md).

Symphony is platform neutral and protocol neutral. Through its extensible architecture, Symphony supports a broad range of AI frameworks, devices, hardware, services and many more. The following table summarizes some of the supported technologies and the list grows rapidly.

| Area | Supported Technologies |
|--------|--------|
| AI Pipeline | ![ONNX](../images/onnx.png) ![OpenCV](../images/opencv.png), DeepStream* |
| Application Model | ![Helm](../images/helm.png), Symphony, Radius*, ARM* |
| Device Updates | ![GitOps](../images/gitops.png) ![Flux](../images/flux.png)*, ADU for IoT Hub, ![Arc](../images/arc.png) ![pyOCD](../images/pyocd.png)*|
| Discovery | ![ONVIF](../images/onvif.png) ![OPC UA](../images/opcua.png)*, udev (via Akri) |
| Hardware | Azure Stack HCI, MIMXRT1170-EVK, Nvidia Jeston Orin, Nvidia dGPU |
| K8s Distributions | ![Kubernetes](../images/k8s.png) ![Azure Kubernetes Service](../images/aks.png) ![MicroK8s](../images/microk8s.png) ![Kind](../images/kind.png) ![K3s](../images/k3s.png) AKS-IoT |
| Observability | ![Open Telemetry](../images/open-telemetry.png), Azure Monitor |
| OS | ![Ubuntu](../images/ubuntu.png) ![Windows 10](../images/windows.png) ![MacOS](../images/macos.png), CBL-Mariner, Azure RTOS |
| Other Azure Service | Azure Storage, Azure Logic Apps, Azure Functions |
| Policies | Gatekeeper, ![Kyverno](../images/kyverno.png) |
| Runtime | Azure IoT Edge, Kubernetes, Windows 10, Samsung DERAM* |
| Scripting | Bash, ![PowerShell](../images/powershell.png), Windows Batch |

_*:upcoming_

## Feature set
Symphony offers the following key features:

* Symphony resources are projected as native K8s CRDs, allowing the control plane to be managed using K8s native tools such as kubectl.
* Support different application model formats, including Helm charts, OAM/Kubevela*, Radius*, Docker Compose*, and a ModuleGroup format that is designed to group multiple Azure IoT Edge modules.
* Support hardware-accelerated AI payloads using media pipelines through Live AI, OpenVINO, etc.
* Share the same solution management, security management and device management logic to ensure consistent behavior across cloud and edge.
* Designed to work well with AKS, Arc, IoT Hub and other Azure services.
* Support dynamic device discovery and update through Akri.
* Support additional computational nodes such Azure Sphere through Virutal Kubelet.
* End-to-end observability with distributed tracing using OpenTelemetry.

And for longer term, Symphony aims to enable:

* Azure billing integration.
* Suppport UX-based apps/components.
* Support RTOS (as compute nodes).
* Dynamic hybrid scenarios such as bursting to cloud.
* A graphic management portal, likely as a mobile applicaiton.

## Key application scenarios
Symphony aims to deploy and manage secured, hardware-accelerated intelligenet edge payload on a K8s-based fabric that offers adaptive workload scheduling for HA and resource balancing. Some key applicaiton scenarios include:

* Manage intelligent payloads on a HA field gateway cluster, such as a HCI cluster, that manages attached sensors like brown-field cameras.
* Manage intelligent payloads on a in-vehicle cluster for smart cars, construction vechicile and/or aireplanes.
* Provide business continuity for occationally connected or disconnected environments.
* Manage intelligent payloads on a fully virutalized environments for large-scale scenarios such as simulation and testing.

## Development strategy

We’ll co-develop Symphony in collaboration with first-party partners and third-party partners. And the aim is to release the control plane and related implementations to the community as open source projects.