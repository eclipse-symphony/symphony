# Patch stage provider

Patch stage provider patches an existing solution version — adding, replacing, or removing a component, or updating a property list of a specific component. The patch content either comes from a catalog version or is given inline in the stage inputs.

## Inputs

| Field | Value |
|-------|-------|
| `objectType` | `solutionversion`. |
| `objectName` | Name of the solution version to patch, e.g. `test-app:version1`. |
| `patchSource` | (optional) `catalogversion` (default) or `inline`. |
| `patchContent` | For `catalogversion`: the name of the catalog version that holds the patch. For `inline`: an inline component definition, or a property map when `component` is set. |
| `component` | (optional) Name of the component to patch. When empty, the whole component from the patch content is added, replaced, or removed. |
| `property` | (optional) Name of the component property (a list) to patch. |
| `subKey` | (optional) Key inside a map-typed property that holds the list to patch. |
| `dedupKey` | (optional) Property used to match existing list entries; matching entries are replaced (or removed) instead of appended. |
| `patchAction` | (optional) `add` (default) or `remove`. |
| `objectNamespace` | (optional) Namespace of the object. Defaults to `default`. |

## Outputs

The patch provider doesn't produce outputs.

## Sample

Add a new container component to a solution version:

```yaml
deploy-v2:
  name: "deploy-v2"
  provider: "providers.stage.patch"
  inputs:
    objectType: solutionversion
    objectName: test-app:version1
    patchSource: inline
    patchContent:
      name: backend-v2
      type: container
      properties:
        deployment.replicas: "#1"
        container.image: "ghcr.io/eclipse-symphony/sample-flask-app:latest"
    patchAction: add
  stageSelector: "canary-ingress"
```
