# Solution

You can assemble components from different artifact formats into an orchestrated application model using Symphony’s `solution` object.

A [solution](../concepts/unified-object-model/solution.md) describes an application. It consists of a list of [components](../concepts/unified-object-model/solution.md#componentspec), which can be a container, a Helm chart, a Kubernetes artifact file, a security policy, a firmware, or anything else. Instead of forcing artifacts to adopt the Symphony [component](../concepts/unified-object-model/solution.md#componentspec) artifact format, Symphony allows existing application artifacts to be directly embedded into Symphony solutions.

When modeling a microservice application, components are assumed to be independent from each other. However, in many legacy applications there are implicit or explicit dependencies among components. Symphony allows you to attach optional dependencies to components to build up a dependency tree. When Symphony deploys the solution, it walks the dependency tree and ensures that components are deployed in the correct order.

## Solution schema

Solution objects, `solution.solution.symphony`, define an intelligent edge solution.

| Field | Type | Description |
|--------|--------|--------|
| `Components`| `[]ComponentSpec` | A list of components |
| `DisplayName` | `string` | A user friendly name |
| `Metadata` | `map[string]string` | metadata |
| `Scope` | `string` | A scope name, which usually maps to a namespace |

## ComponentSpec schema

| Field | Type | Description |
|--------|--------|--------|
| `Name`| `string` | component name | 
| `Constraints` | `map[string]ConstraintSpec` | component constraints |
| `Dependencies` | `[]string` | component dependencies |
| `Properties` | `map[string]string` | component properties |
| `Routes` | `[]RoutSpec` | incoming/outgoing routes |
| `Skills` | `[]string` | Referenced [AI skills](./ai-skill.md) |
| `Type` | `string` | component type |

### Dependencies

When Symphony deploys a solution, it sorts all solution components by their dependencies so that [providers](../providers/_overview.md) that allow ordering can apply the components in order.

Circular references are not allowed.

## Related topics

* [Solution schema](../concepts/unified-object-model/solution.md)
* [Configuration management](./configuration-management.md)
