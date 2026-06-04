---
type: docs
title: "Orchestration Model"
linkTitle: "Orchestration Model"
description: ""
weight: 500
---

As an orchestrator, Symphony purposefully designs a modeling language for orchestration purposes. We refer to this model as the *orchestration model*. Objects in the orchestration model establish desired state of a system. Symphony runs a continuous state seeking loop that brings the probed system current state towards the desired state. 

## Three core object types

Symphony desired state is described using three core object types: **Solution**, **Target**, and **Instance**.

A **Solution** describes a software stack. A **Target** describes an environment where software can be deployed. An **Instance** describes a deployment of a Solution to one or more Targets.

Why use three separate objects? The design separates three distinct concerns that may have different lifecycles and involve different personas, as summarized in the following table:

| Object Type | Usage | User Persona |
|---|---|---|
| Solution | Describes a software stack | Developers, ISVs |
| Target | Describes an infrastructure environment | IT, Ops |
| Instance | Describes a deployment of software to infrastructure | IT, Ops |

### Solution

A **Solution** describes a software stack. It consists of a list of Components, each representing a piece of software, such as a container, an eBPF module, a binary image, or anything else supported by the Symphony ecosystem.

Components can have dependencies, and Symphony ensures that they are deployed according to their dependency relationships.

### Target

A **Target** describes an infrastructure environment where software can be deployed. A Target is not limited to a physical piece of hardware. It can be a server, a virtual machine, an edge device, a cluster such as Kubernetes, a service endpoint, or simply a virtual target.

As a side note, a Target can also define its own Component list. These are infrastructure-level components that are independent of Solutions. Symphony applies the same state-seeking loop to Targets to ensure that their desired infrastructure components are provisioned.

### Instance

An **Instance** takes a Solution and maps its Components to one or more Targets.

A Solution can be deployed to selected Targets as a complete stack. Its Components can also be distributed across multiple Targets based on target characteristics. For example, a frontend component could be deployed to a Windows machine, while a backend service is deployed to a Kubernetes cluster.


## Other object types

In addition to the three core object types, Symphony also provides additional object types for different scenarios.

### Catalog

A **Catalog** is a versatile object in Symphony. It is a generic carrier for a piece of information, such as application configuration, security policies, solution templates, and other reusable artifacts.

When multiple Symphony installations are linked together, Catalog is the object type that is synchronized across Symphony installations, which are called **sites**.

Catalog objects support information ontology modeling capabilities such as inheritance, overrides, and composition. They also support Symphony's expression language for dynamic evaluation.

### Campaign and Activation

A **Campaign** defines a workflow, such as a canary deployment, gated deployment, or scheduled deployment.

An **Activation** is a running instance of a Campaign. You can have multiple Activations of the same Campaign, and they run independently from one another.

## Multi-versioned objects

Some Symphony objects support multi-versioning, including **Solution**, **Catalog**, and **Campaign**. These objects can hold multiple versions.

For example, a Solution object can be linked to multiple SolutionVersion objects, each representing a specific version of the Solution. This enables additional operations such as rollback.

