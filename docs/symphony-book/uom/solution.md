# Solution

```solution.solution.symphony``` defines an intelligent edge solution.

## Schema
| Field | Type | Description |
|--------|--------|--------|
| ```Components```| ```[]ComponentSpec``` | A list of components | 
| ```DisplayName``` | ```string``` | A user friendly name |
| ```Metadata``` | ```map[string]string``` | metadata |
| ```Scope``` | ```string``` | A scope name, which usually maps to a namespace |

## ```ComponentSpec```
| Field | Type | Description |
|--------|--------|--------|
| ```Name```| ```string``` | component name | 
| ```Constraints``` | ```map[string]ConstraintSpec``` | component constraints |
| ```Dependencies``` | ```[]string``` | component depedencies |
| ```Properties``` | ```map[string]string``` | component properties |
| ```Routes``` | ```[]RoutSpec``` | incoming/outgoing routes |
| ```Skills``` | ```[]string``` | Referenced [AI skills](./ai-skill.md) |
| ```Type``` | ```string``` | component type |

### Depedencies
When Symphony deploys a solution, it first sorts all solution components by their depedencies. This allows [providers](../providers/providers.md) that allow ordering to apply the components in order. 

> **NOTE**: Circular references are not allowed.