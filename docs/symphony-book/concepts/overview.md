# Symphony concepts

Symphony operates within an orchestration layer, strategically positioned atop preexisting tools and services. To establish uniformity amidst the diverse edge environment, Symphony relies on a set of fundamental high-level abstractions. These encompass state-seeking, graph, workflow, and app models. These foundational abstractions empower Symphony to deliver robust functionalities across various technologies and platforms while maintaining a concise object model.

## Abstractions

* [State seeking](./state_seeking.md): How Symphony brings the current state of a system to the desired state.
* [Information graph](./information_graph.md): How Symphony organizes, accesses, and synchronizes enterprise information.
* [Workflows](./workflows.md): How Symphony manages multi-stage deployment scenarios.
* [App orchestration model](./orchestration_model.md): How Symphony defines the components that make up a scenario.

## Object model

The Symphony object model defines common elements of intelligent solutions as Kubernetes custom resources that can be managed using standard tools. For more information, see [Unified object model](../uom/uom.md).

### Foundational objects

* [Solution](../solution-management/solution-management.md)
* [Instance](../uom/instance.md)
* [Target](../target-management/target-management.md)
* [Device](../device-management/device-management.md)

### Federation objects

* [Campaign](../campaign-management/campaign.md)
* [Catalog](../information-graph/overview.md)

## AI objects

* AI Model
* [AI Skill](../skill-management/skill-management.md)
* AI Skill Package