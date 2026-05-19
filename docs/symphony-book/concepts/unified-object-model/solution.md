# SolutionVersion

You can assemble components from different artifact formats into an orchestrated application model using Symphony’s `solutionversion` object.

A [solutionversion](../../concepts/unified-object-model/solutionversion.md) describes an application. It consists of a list of [components](../../concepts/unified-object-model/solutionversion.md#componentspec-schema), which can be a container, a Helm chart, a Kubernetes artifact file, a security policy, a firmware, or anything else. Instead of forcing artifacts to adopt the Symphony [component](../../concepts/unified-object-model/solutionversion.md#componentspec-schema) artifact format, Symphony allows existing application artifacts to be directly embedded into Symphony solutionversions.

When modeling a microservice application, components are assumed to be independent from each other. However, in many legacy applications there are implicit or explicit dependencies among components. Symphony allows you to attach optional dependencies to components to build up a dependency tree. When Symphony deploys the solutionversion, it walks the dependency tree and ensures that components are deployed in the correct order.

## SolutionVersion schema

SolutionVersion objects, `solutionversion.solutionversion.symphony`, define an intelligent edge solutionversion.

| Field | Type | Description |
|--------|--------|--------|
| `Components`| `[]ComponentSpec` | A list of components |
| `DisplayName` | `string` | A user friendly name |
| `Metadata` | `map[string]string` | metadata |

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

When Symphony deploys a solutionversion, it sorts all solutionversion components by their dependencies so that [providers](../../providers/_overview.md) that allow ordering can apply the components in order.

Circular references are not allowed.

## Related topics

* [Configuration management](../../configuration-management/_overview.md)
