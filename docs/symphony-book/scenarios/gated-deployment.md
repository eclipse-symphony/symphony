# Gated Deployment
Symphony provides two different ways to implement gated deployment: through the [instance](../uom/instance.md) **Stage** property, or through a HTTP web hook.The stage property can be used for manual control. And the web hook can be used for either manual control or automated control (such as hooking into an automated approval process).

## Gated deployment through instance property

Symphony [instance](../uom/instance.md)  object has a **Stage** property that can be used to control different [stages](../instance-management/instance-management.md#stages) of a deployment. A special stage, **BLOCK**, can be used to block an instance from being deployed. When an Instance object is created with the stage property set to BLOCK, the instance will not be deployed to targets until this property has been cleared or changed. 

## Gated deployment through web hook

Symphony allows you to insert processing gates through web hooks during your [instance](../uom/instance.md) deployments.

<!--
For example, you can invoke an external approval process, as introduced in the [human approval](./human-approval.md) scenario. 
-->

You can set up multiple gates in your [instance](../uom/instance.md). And each gate governs a set of components who set a [dependency](../uom/solution.md#depedencies) on the gate. If a gate is at the root of the dependency graph, the gate controls the deployment of all components.

A gate is “open” if the HTTP request returns 200. Otherwise, the gate is “closed” and will block the deployment from going further.