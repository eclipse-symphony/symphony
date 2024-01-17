# Instance

`instance.solution.symphony` maps a [`solution.solution.symphony`](./solution.md) object to one or more [`target.fabric.symphony`](./target.md) objects. Creating an instance object triggers deployments to targets.

## Schema

| Field | Type | Description |
|--------|--------|--------|
| `DisplayName` | `string` | A user friendly name |
| `Metadata` | `map[string]string` | Deployment metadata |
| `Parameters` | `map[string]string` | Parameters. A parameter can be used anywhere in the skill definition. See the [parameters](#parameters) sections below |
| `Pipelines` | `[]PipelineSpec` | AI pipeline references |
| `Schedule` | `string` | Deployment schedule |
| `Scope` | `string` | Deployment scope (such as Kubernetes namespace) |
| `Solution` | `string` | Solution name |
| `Stage` | `string` | Deployment stage |
| `Target` | `TargetSelector` | Targets (see [Target selection](#target-selection)) |
| `Topologies` | `[]TopologySpec` | Device topologies |

## Target selection

You can specify a single target using `name`; Or, you can select a group of targets using `selector`. A selector specifies a list of key-value pairs. Any target with the matching properties are considered a match. For example, the selector:

```yaml
target:
  selector:
    foo: bar
    group: group-1
```

Matches any targets that hold the `foo=bar` property **AND** the `group=group-1` property:

```yaml
properties:
  foo: bar
  group: group-1
  other: properties
```
