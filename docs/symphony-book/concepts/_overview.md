# Symphony concepts

Symphony operates within an orchestration layer, strategically positioned atop preexisting tools and services. To establish uniformity amidst the diverse edge environment, Symphony relies on a set of fundamental high-level abstractions. These encompass state-seeking, graph, workflow, and app models. These foundational abstractions empower Symphony to deliver robust functionalities across various technologies and platforms while maintaining a concise object model.

## Abstractions

* [State seeking](./state_seeking.md): How Symphony brings the current state of a system to the desired state.
* [Information graph](./information_graph.md): How Symphony organizes, accesses, and synchronizes enterprise information.
* [Workflows](./workflows.md): How Symphony manages multi-stage deployment scenarios.
* [App orchestration model](./orchestration_model.md): How Symphony defines the components that make up a scenario.

## Object model

The Symphony object model defines common elements of intelligent solutions as Kubernetes custom resources that can be managed using standard tools. For more information, see [Unified object model](../concepts/unified-object-model/_overview.md).

### Foundational objects

Foundational objects represent the hardware and software for your intelligent solution.

Use a `device` object to register managed devices like cameras or other sensors. A `target` object represents a computational resource that can receive Symphony deployments. Devices can be attached to and managed by target objects. A `solution` object defines an application to be deployed, and an `instance` object applies a specific solution to one or more targets.

For more information, see:

* [Device](../concepts/unified-object-model/device.md)
* [Target](../concepts/unified-object-model/target.md)
* [Solution](../concepts/unified-object-model/solution.md)
* [Instance](../concepts/unified-object-model/instance.md)

### Federation objects

Federation objects support your business processes as you define and roll out Symphony deployments. A `campaign` object is Symphony's representation of a deployment workflow. You can use campaigns to define multi-stage deployments. A `catalog` object is a generic graph data structure that you can use to create information models for your organization.

For more information, see:

* [Campaign](./unified-object-model/campaign.md)
* [Catalog](./unified-object-model/catalog.md)

## AI objects

AI objects represent elements of AI workflows. Use the `model` object to represent AI models or transformations that you apply to your data, and use the `skill` object to define the processing pipelines that implement the models.

For more information, see:

* [AI skill](../skill-management/skill-management.md)
* [AI model](../concepts/unified-object-model/ai-model.md)