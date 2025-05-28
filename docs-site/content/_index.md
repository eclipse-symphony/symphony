---
type: docs
title: "Project Symphony Documentation"
linkTitle: "Home"
description: ""
weight: 1
no_list: true
---

{{< image src="symphony icon_largetext (3).png" width="500px" >}}

# Welcome to Symphony!

Have you encountered the complexities of manually orchestrating multiple toolchains to oversee your edge infrastructure and solutions? Symphony orchestrates existing services and tools to form an end-to-end, consistent intelligent edge experience. Symphony unites various device classes, software artifacts, and service toolchains, seamlessly harmonizing them into a unified system. With Symphony, you become the maestro, holding the complete vision and exerting precise control while enabling each component to realize its full potential, regardless of whether it's on heavy edge, light edge, or tiny edge, and irrespective of whether it's running on Kubernetes, in the cloud, or on-premises services.

# Features

To enable your edge solution orchestration, Symphony:

- Projects resources as Kubernetes custom resources, allowing the control plane to be managed using K8s native tools such as kubectl.
- Supports different application model formats, including Helm charts, and a ModuleGroup format that is designed to group multiple Azure IoT Edge modules.
- Supports hardware-accelerated AI payloads using media pipelines through Live AI, OpenVINO, etc.
- Shares the same solution management, security management, and device management logic to ensure consistent behavior across cloud and edge.
- Works well with AKS, Arc, IoT Hub and other Azure services.
- Supports dynamic device discovery and update through Akri.
- Supports additional computational nodes such as Azure Sphere through Virtual Kubelet.
- Supports end-to-end observability with distributed tracing using OpenTelemetry.

# Key application scenarios

Symphony aims to deploy and manage secured, hardware-accelerated intelligent edge payloads on a K8s-based fabric that offers adaptive workload scheduling for HA and resource balancing. Some key application scenarios include:

- Manage intelligent payloads on a highly available field gateway cluster, such as an HCI cluster, that manages attached sensors like brown-field cameras.
- Manage intelligent payloads on a in-vehicle cluster for smart cars, construction vehicles, and/or airplanes.
- Provide business continuity for occasionally connected or disconnected environments.
- Manage intelligent payloads on a fully virtualized environments for large-scale scenarios such as simulation and testing.
