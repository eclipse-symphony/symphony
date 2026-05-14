---
type: docs
title: "Project Symphony Documentation"
linkTitle: "Home"
description: ""
weight: 1
no_list: true
---

{{< image src="symphony icon_largetext (3).png" width="250px" >}}

The edge is heterogeneous, dynamic, and distributed by nature. It is not a fixed cluster with stable membership. Devices come and go. Workloads move. Networks disconnect and recover. Capabilities vary across heavy edge, light edge, and tiny edge environments. Traditional orchestration, built around static membership and centralized assumptions, is not enough for this reality.

Many assumptions made by cloud-centric orchestrators do not hold at the edge: uniform infrastructure, well-protected security boundaries, a single hardware or software vendor, stable connectivity, and uniformly scaled workloads. We want to bring cloud-like management benefits to edge environments, but that does not mean we can simply apply cloud-centric technologies and designs as-is. Edge orchestration must be designed with the edge in mind.

Symphony embraces the fluid nature of the edge. It brings together existing services, device classes, software artifacts, and operational toolchains into a consistent, end-to-end intelligent edge experience. Whether components run on Kubernetes, in the cloud, on-premises, or on constrained devices, Symphony provides the coordination layer needed to harmonize distributed components into a unified system without forcing them into a one-size-fits-all model.

# Symphony Characteristics

Symphony is designed from the ground up for heterogeneous, dynamic, and distributed edge environments. It orchestrates devices and workloads, including AI workloads, without assuming a fixed cluster, a single platform, or a uniform software stack. Its key characteristics include:

* **Neutral by design.** Symphony is platform-neutral, protocol-neutral, and artifact-schema-neutral. It works with the toolchains you already have, rather than forcing you to replace them. Existing artifacts, protocols, and operational tools can continue to be used as-is, while Symphony coordinates them as part of a larger system.

* **Adoption that scales with you.** Symphony can run as a single process for simple scenarios or scale out on Kubernetes for production deployments. This allows teams to start small, prove value quickly, and grow into larger deployments without changing the overall operating model.

* **Consistent operations across fragmented toolchains.** Symphony brings security, observability, and policy enforcement across diverse systems, even when individual tools do not provide these capabilities natively. It helps teams apply consistent operational practices across otherwise disconnected environments.

* **Built for constrained and disconnected environments.** Symphony minimizes network traffic and management overhead. Its smallest agent is only 4 KB and can be configured to phone home rarely, or not at all, with a tiny payload. For many scenarios, Symphony also supports agentless deployment, requiring no Symphony components on target devices.

* **End-to-end orchestration capabilities.** Symphony includes device management, workload management, configuration management, workflow management, policy management, template management, and more, providing a unified control layer for complex edge solutions.

* **Deeply extensible.** Symphony uses providers to integrate with different platforms, protocols, and tools. Its internal system components, including the state store, pub-sub channel, and identity provider, can also be replaced or customized to match the needs of your environment.

# Key Application Scenarios

From large-scale validation systems to fleet operations and edge AI infrastructure, Symphony is already being used to solve some of the most demanding edge orchestration challenges:

* **Hardware-in-the-loop testing at scale**, where Symphony coordinates distributed test processes end to end using its workflow capabilities.

* **Configuration management across product lines**, where Symphony centrally manages millions of configuration variations through modularization, inheritance, and dynamic expression evaluation.

* **Fleet management**, where Symphony manages vehicles and other distributed assets with workflows, scheduled and on-demand updates, and multi-homing support.

* **Edge AI infrastructure management**, where Symphony simplifies model rollout, coordinates related software updates, and connects edge AI infrastructure with cloud data pipelines.

# Getting Started


The easiest way to get started with Symphony is by using Symphony's CLI tool, called maestro. Use the appropriate command for your platform and then simply use ```maestro up``` to launch your Symphony instance:

{{< tabpane >}}
{{< tab header="Bash (Linux/WSL/Mac)" lang="bash" >}}
wget -q https://raw.githubusercontent.com/eclipse-symphony/symphony/master/cli/install/install.sh -O - | /bin/bash
maestro up --no-k8s
{{< /tab >}}

{{< tab header="PowerShell (Windows)" lang="powershell" >}}
powershell -Command "iwr -useb https://raw.githubusercontent.com/eclipse-symphony/symphony/master/cli/install/install.ps1 | iex"
maestro up --no-k8s
{{< /tab >}}
{{< /tabpane >}}

# What's Next

* [Getting Started Tutorials]({{< relref "/tutorials" >}})
* [Sample Gallery]({{< relref "/samples" >}})
* [Embracing the Ecosystem]({{< relref "/ecosystem" >}})
