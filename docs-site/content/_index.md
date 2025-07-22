---
type: docs
title: "Project Symphony Documentation"
linkTitle: "Home"
description: ""
weight: 1
no_list: true
---

{{< image src="symphony icon_largetext (3).png" width="500px" >}}


Have you encountered the complexities of manually orchestrating multiple toolchains to oversee your edge infrastructure and solutions? Symphony orchestrates existing services and tools to form an end-to-end, consistent intelligent edge experience. Symphony unites various device classes, software artifacts, and service toolchains, seamlessly harmonizing them into a unified system. With Symphony, you become the maestro, holding the complete vision and exerting precise control while enabling each component to realize its full potential, regardless of whether it's on heavy edge, light edge, or tiny edge, and irrespective of whether it's running on Kubernetes, in the cloud, or on-premises services.

# Features

To enable your edge solution orchestration, Symphony:

- Defines a meta model that orchestrate different application package formats such as Helm, Docker, Azure IoT Edge module, binaries, WebAssemblies, eBFP modules and disk images.
- Supports continuous state seeking to ensure system remain in the desired state without drifting.
- Works natively on Kubernetes through Kubernetes API, or as a standalone service with REST API. 
- Protocol bindings allow Symphony APIs to be consumed through protocols other than HTTPS, such as MQTT.
- Supports end-to-end observability with distributed tracing using OpenTelemetry.
- Enables complex deployment scenarios like canary depoloyment, gated deployment, scheduled deployment etc. through workflows.
- Supports hardware-accelerated AI payloads using media pipelines through Live AI, OpenVINO, etc.
- Works well with AKS, Arc, IoT Hub and other Azure services.

# Key application scenarios

Symphony aims to deploy and manage heterogenous, distirbuted edge computing infrastructure and payloads. Some key application scenarios include:

- Provide consistent, end-to-end managment of edge payloads that span multiple operation systems, application package formats, distribution channels and update mechanisms.
- Provide business continuity for occasionally connected or disconnected environments.
- Manage intelligent payloads on a fully virtualized environments for large-scale scenarios such as simulation and testing.
