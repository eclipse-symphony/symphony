# Target provider interface

_(last edit: 6/26/2023)_

Symphony defines a simple provider interface with four methods:

```go
type ITargetProvider interface {
    Init(config providers.IProviderConfig) error
    // get validation rules
    GetValidationRule(ctx context.Context) model.ValidationRule
    // get current component states from a target. The desired state is passed in as a reference
    Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error)
    // apply components to a target
    Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error)
}
```

To simplify developer development, Symphony assumes:

* A provider is thread-unsafe
* A provider is stateless
* A provider may or may not be idempotent

## Get

The `Get()` method probes the target device and returns the currently installed components. Because a target device may host applications from different sources, the `Get()` method passes in the expected deployment description as well as a component list as a reference to help the provider narrow down the scope of interest. For example, a mobile phone may have dozens of applications installed. The deployment description informs the provider which application(s) should be checked.

A provider works mostly with the reference component list. If additional context is needed, the provider can consult the deployment spec.

## Apply

The `Apply()` method runs a `DeploymentStep`, which contains a list of `ComponentStep` items. The provider needs to operate on each of the `ComponentStep` to apply the desired state to the target system.

A `ComponentStep` can be either an `"update"` or `"delete"` step. Your provider should perform the corresponding actions based on this flag.

> **NOTE**: A provider is expected to operate only on components affected by the deployment descriptions. Some providers may need to wipe the whole device during deployment. In such cases, the development description is expected to contain the entire software stack. Or a device image can be treated as a single component in a solution. If a provider maintains a collection resource that holds components, it should remove the collection resource when the collection becomes empty.

## Get validation rule

The `GetValidationRule()` method explicitly defines what properties the provider handles. The Symphony component spec has an open properties collection that may contain any key-value pairs. This method allows Symphony to probe the provider for the exact properties that are expected by the specific provider. This allows Symphony to offer consistent validation behaviors across all providers.

## Dry run

When the `isDryRun` flag is set, the provider validates the component specs without doing actual deployments. You can access the validation result through the returned `err` object.
