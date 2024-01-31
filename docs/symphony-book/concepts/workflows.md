# Workflows

Symphony employs a [state-seeking](./state_seeking.md) approach to maintain a system according to user-defined desired states. Nevertheless, there are situations where orchestrating system control necessitates capabilities beyond state-seeking. For example, a scenario may require an approval process prior to initiating a deployment; in such cases, Symphony needs the capability to trigger an approval workflow before proceeding with its state-seeking operations. Another example is managing canary deployments, which requires the multi-step operation of rolling out a new version, adjusting your traffic shapes, validating the new version, and then gradually shifting traffic to the new version (and rolling back as needed).

These complex deployment processes are represented by a workflow, which is called a *campaign* in Symphony. A campaign comprises one or multiple *stages*. When one stage finishes, it runs a selector that selects the next stage. And when no new stages are selected, the campaign finishes.

Comparing to some other workflow engines, Symphony has the following features specifically designed for distributed edge scenarios:

* **Remote execution**: Symphony can dispatch a specific stage to a distinct Symphony control plane for execution, retrieve the resultant output, and seamlessly progress to the subsequent stage in the workflow.

* **Simple map-reduce**: A stage can fan out across multiple contexts, running in parallel, and aggregate the results into a cohesive outcome. When combined with remote execution, the map-reduce feature can be used to coordinate multi-site rollouts from a central location.

* **External driver**: A Symphony campaign can drive itself to completion. Or, it can be put into a stepped mode where it needs an external workflow engine to drive it forward. In this model, a Symphony campaign provides a collection of stateful nano services that can be invoked by another workflow engine. For example, many customers have their existing CI/CD pipelines in systems like GitHub Actions and Jenkins. Symphony allows customers to maintain their CI/CD pipelines in these systems, leveraging the rich ecosystem of ingredients. Only when they need to call out to specific Symphony functionalities like multi-site deployment, they can call out to Symphony campaigns. And because Symphony campaigns are stateful, customers can carry states across systems and form a continuous flow although stages may interact among multiple systems.

## Activation and events

A campaign can be activated by creating the corresponding *activation* object. Unless explicitly disabled, a campaign can be activated multiple times in parallel. Each activation object maintains its own state and context. Symphony uses an event-based approach to drive the activation objects. And it allows steps in a campaign to be executed remotely on a different Symphony control plane.

After a stage's execution, the campaign enters a dormant state, awaiting an incoming event to awaken and activate the next stage.

## Related topics

* [Campaign scenarios](./campaign-scenarios.md)