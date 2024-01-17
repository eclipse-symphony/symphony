# Target

A `target` in Symphony is an endpoint to which Symphony `components` can be deployed. A target can be a server, a PC, a mobile device, a cluster, or any other endpoints that support the Symphony [provider interface](../providers/provider_interface.md). When a target is registered, Symphony allows a full-stack description of all software components, policies, and configurations that are required on the target, and Symphony uses its state-seeking mechanism to make sure that the target is configured properly.

![target](../images/target.png)

Symphony ships a number of providers out-of-box to support various target types. And Symphony is extensible with either [native providers](../providers/_overview.md#provider-types) (that come with Symphony builds), [script providers](../providers/script_provider.md) or [standalone providers](../providers/standalone_providers.md) via a proxy.

## Components

You can define the full software stack on a `target` by adding software artifacts to its `components` collection. Each component can use a different artifact format and is handled by a different provider. For example, you can embed a Kubernetes Yaml file as a component and uses a Kubernetes native provider to install it.

Each component can have one or multiple dependencies. Symphony sorts the dependency graph by topological order and ensures that the components are installed in the correct order.

For example, the following components collection defines three components: symphony-agent, policies, and gatekeeper. Symphony-agent relies on both policies and gatekeeper, and policies relies on gatekeeper. This makes gatekeeper the first component to be installed, policies is the second, and symphony-agent is the last.

```yaml
components:
  - name: "symphony-agent"
    properties:
      ...
    dependencies:
    - gatekeeper
    - policies
  - name: "gatekeeper"
    type: yaml.k8s
    properties:
      ...    
  - name: "policies"
    type: yaml.k8s
    properties:
      ...
    dependencies:
    - gatekeeper
```

## Role bindings

A component is bound to a [provider](../providers/target_provider.md) through a role binding by component type. For example, the following binding binds a `yaml.k8s` component type to a `providers.target.kubectl` provider.

```yaml
topologies:
  - bindings:
    - role: yaml.k8s
      provider: providers.target.kubectl
      config:
        inCluster: "true"
```

## Related topics

* [Providers](../providers/_overview.md)