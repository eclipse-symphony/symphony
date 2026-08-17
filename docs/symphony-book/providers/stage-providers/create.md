# Create stage provider

Create stage provider creates or removes Symphony objects — solution versions or instances — as part of a campaign. When creating an instance, the provider waits until the deployment finishes and reports the result in the outputs.

## Config

| Field | Value |
|-------|-------|
| `wait.count` | Maximum number of checks while waiting for a deployment or a deletion to finish. |
| `wait.interval` | Seconds between checks. Defaults to `20`. |

## Inputs

| Field | Value |
|-------|-------|
| `objectType` | `solutionversion` or `instance`. |
| `objectName` | Name of the object, e.g. `site-app:version1` for a solution version. |
| `action` | `create` or `remove`. |
| `object` | Object definition (`metadata` and `spec`), required for the `create` action. |
| `objectNamespace` | (optional) Namespace of the object. Defaults to `default`. |

## Outputs

| Field | Value |
|-------|-------|
| `objectType` | The object type from the inputs. |
| `objectName` | Name of the created or removed object. |
| `failedDeploymentCount` | Number of failed deployments when creating an instance. |
| `status` | `200` on success, `400` when the deployment failed or timed out (instance creation only). |
| `error` | Error message when `status` is `400`. |

## Sample

Create an instance and branch on the deployment result:

```yaml
create:
  name: "create"
  provider: "providers.stage.create"
  config:
    wait.count: 10
    wait.interval: 20
  inputs:
    action: "create"
    objectName: "site-instance"
    objectType: "instance"
    object:
      metadata:
        name: site-instance
      spec:
        solutionversion: site-app:version1
        target:
          name: site-k8s-target
  stageSelector: ${{$if($equal($output(create, failedDeploymentCount), 0),'succeeded','failed')}}
```
