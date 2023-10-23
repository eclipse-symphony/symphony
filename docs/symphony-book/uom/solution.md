# Solution

Solution objects, `solution.solution.symphony`, define an intelligent edge solution.

## Solution schema

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
| `Dependencies` | `[]string` | component depedencies |
| `Properties` | `map[string]string` | component properties |
| `Routes` | `[]RoutSpec` | incoming/outgoing routes |
| `Skills` | `[]string` | Referenced [AI skills](./ai-skill.md) |
| `Type` | `string` | component type |

### Dependencies

When Symphony deploys a solution, it sorts all solution components by their dependencies so that [providers](../providers/overview.md) that allow ordering can apply the components in order.

Circular references are not allowed.

## Related topics

* [Modeling applications](../solution-management/solution-management.md)