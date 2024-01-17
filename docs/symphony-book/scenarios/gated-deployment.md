# Gated deployment

Symphony provides two different ways to implement gated deployment: through the **Stage** property of an [instance](../concepts/unified-object-model/instance.md) object, or through an HTTP web hook. The stage property can be used for manual control. The web hook can be used for either manual control or automated control (such as hooking into an automated approval process).

## Gated deployment through instance property

A Symphony instance object has a **Stage** property that can be used to control different [stages](../instance-management/instance-management.md#stages) of a deployment. A special stage, **BLOCK**, can be used to block an instance from being deployed. When an instance object is created with the stage property set to BLOCK, the instance will not be deployed to targets until this property has been cleared or changed.

## Gated deployment through web hook

Symphony allows you to insert processing gates through web hooks during your instance deployments.

<!--
For example, you can invoke an external approval process, as introduced in the [human approval](./human-approval.md) scenario. 
-->

You can set up multiple gates in your instance. Each gate governs a set of components that set a [dependency](../concepts/unified-object-model/solution.md#depedencies) on the gate. If a gate is at the root of the dependency graph, the gate controls the deployment of all components.

A gate is “open” if the HTTP request returns 200. Otherwise, the gate is “closed” and will block the deployment from going further.
