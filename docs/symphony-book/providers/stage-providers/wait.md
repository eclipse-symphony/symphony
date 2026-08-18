# Wait stage provider

Wait stage provider waits until a list of Symphony objects exists. It polls the object list of the given type and finishes once every named object is found. When the wait runs inside an activation, it also stops if the activation itself is deleted.

## Config

| Field | Value |
|-------|-------|
| `wait.interval` | (optional) Seconds to wait between checks. |
| `wait.count` | (optional) Maximum number of checks. `0` means wait indefinitely. |

## Inputs

| Field | Value |
|-------|-------|
| `objectType` | `instance`, `sites`, or `catalogversions`. |
| `names` | List of object names to wait for. |
| `objectNamespace` | (optional) Namespace of the objects. Defaults to `default`. Not used for `sites`. |

## Outputs

| Field | Value |
|-------|-------|
| `objectType` | The object type from the inputs. |
| `status` | `200` once all objects are found. |

## Sample

Wait for two instances to be created before moving on:

```yaml
wait-instances:
  name: "wait-instances"
  provider: "providers.stage.wait"
  config:
    wait.interval: 20
    wait.count: 30
  inputs:
    objectType: instance
    names:
    - "site-instance-1"
    - "site-instance-2"
  stageSelector: "next-stage"
```
