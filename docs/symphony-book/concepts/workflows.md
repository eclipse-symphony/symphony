# Workflows

Many complex management scenarios go beyond a simple state seeking. For example, when you perform a canary deployment, you need to roll out a new version, adjust your traffic shapes, validate the new version, and then gradually shift your traffic to the new version (and roll back as needed). This process is represented by a workflow, which is called a Campaign in Symphony. 

The concept of Symphony Campaign is simple: A Campaign is comprised of multiple stages. And when a stage finishes, it runs a selector that selects the next stage. And when no new stages are selected, the Campaign finishes.

Comparing to some other workflow engines, Symphony Campaign has the following features specifically designed for distributed edge scenarios:

* **Remote execution**: This feature allows Symphony to dispatch a specific stage to a distinct Symphony control plane for execution, retrieve the resultant output, and seamlessly progress to the subsequent stage in the workflow.

* **Simple map-reduce**: A stage can fan out across multiple contexts, running in parallel, and subsequently aggregating the results into a cohesive outcome. When combined with remote execution, the map-reduce feature can be used to coordinate multi-site rollouts from a central location.

* **External driver**: A Symphony campaign can drive itself to completion. Or, it can be put into a stepped mode where it needs an external workflow engine to drive it forward. In this model, A Symphony campaign essentially provides a collection of stateful nano service that can be invoked by another workflow engine. For example, many customers have their existing CI/CD pipelines in systems like GitHub Actions and Jenkins. Symphony allow the customers to continue to maintain their CI/CD pipelines in these systems, leveraging the rich ecosystem of ingredients. Only when they need to call out to specific Symphony functionalities like multi-site deployment, they can call out to Symphony Campaigns. And because Symphony Campaigns are stateful, the customers can carry states across systems and form a continuous flow although stages may interleaved among multiple systems.

## Activation and Events
Unless explicitly disabled, a Campaign can be activated multiple times in parallel. And each activation maintains its own state.

After a stage's execution, the Campaign essentially enters a dormant state, awaiting an incoming event to awaken and activate the next stage.

## Topics

* [Workflow Management](../campaign-management/overview.md)