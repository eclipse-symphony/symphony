# List stage provider

List stage provider lists Symphony objects of a given type and puts the result into the stage outputs. Listing `sites` returns the child sites of the current site. The result is typically consumed by a following stage through the `$output()` function, for example to fan out a deployment to multiple sites.

## Inputs

| Field | Value |
|-------|-------|
| `objectType` | `instance`, `sites`, or `catalogversions`. |
| `namesOnly` | (optional) When `true`, `items` contains only the object names instead of the full object states. Defaults to `false`. |
| `objectNamespace` | (optional) Namespace to list from. Defaults to `default`. Not used for `sites`. |

## Outputs

| Field | Value |
|-------|-------|
| `items` | List of objects (or object names when `namesOnly` is `true`). |
| `objectType` | The object type from the inputs. |

## Sample

List all child sites and pass their names to the next stage:

```yaml
list:
  name: "list"
  provider: providers.stage.list
  inputs:
    objectType: sites
    namesOnly: true
  stageSelector: wait-sync
```
