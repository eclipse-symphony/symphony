# Materialize stage provider

Materialize stage provider creates Symphony objects from catalog versions. Each entry in the `names` input references a catalog version whose `spec.properties` holds an embedded object definition; the provider reads each catalog version and creates the embedded object in the given namespace. Supported catalog types are `instance`, `solutionVersion`, `target`, and wrapped `catalogVersion`/`config` objects. Parent containers (solution or catalog) are created automatically when they don't exist.

## Config

| Field | Value |
|-------|-------|
| `waitForDeployment` | (optional) When `true`, the provider waits until the deployments of the created instances and targets finish. Currently supported in Kubernetes mode only. |
| `waitTimeout` | (optional) How long to wait for deployments, e.g. `5m`. Defaults to `5m` when `waitForDeployment` is `true`. |

## Inputs

| Field | Value |
|-------|-------|
| `names` | List of catalog version references to materialize, e.g. `gated-prometheus-instance`. |
| `objectNamespace` | (optional) Namespace in which the objects are created. Defaults to `default`. |

## Outputs

| Field | Value |
|-------|-------|
| `failedDeployment` | List of failed deployments (only when `waitForDeployment` is `true`). |
| `failedDeploymentCount` | Number of failed deployments (only when `waitForDeployment` is `true`). |

## Sample

Materialize an instance from a catalog version:

```yaml
deploy:
  name: "deploy"
  provider: "providers.stage.materialize"
  inputs:
    names:
    - "gated-prometheus-instance"
```
