---
type: docs
title: "App Orchestration Model"
linkTitle: "App Orchestration Model"
description: ""
weight: 500
---

As an orchestrator, Symphony purposefully designs an application model tailored for orchestration. We refer to this specialized application model as the *app orchestration model* to distinguish it from a standard application model.

An app orchestration model defines a collection of interconnected components. Each component can be represented by a different artifact type, such as a Helm Chart, a Kubernetes deployment spec, a Docker container, a configuration map, or anything else. As you can see, Symphony orchestration model allows multiple artifacts from different systems be assembled into one consistent package. 

The orchestration model sits on top of application models. Symphony doesn’t require an application to be remodeled using the orchestration model. Instead, all components can still use existing artifact formats and only the orchestration layer uses the Symphony model.

Although in typical microservices architecture, components are independent from each other. Symphony allows dependencies to be declared to accommodate legacy systems where explicit or implied dependencies exist. During deployment, Symphony makes sure the components are applied in the correct order.

Symphony distinguishes application components and infrastructural components. Infrastructural components, such as databases and file servers, are usually defined on a Symphony Target object. 