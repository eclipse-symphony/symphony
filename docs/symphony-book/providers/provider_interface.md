# Target Provider Interface

Symphony defines a simple Provider interface with 5 methods:

```go
type ITargetProvider interface {
    // apply components to a target
    Apply(ctx context.Context, deployment model.DeploymentSpec) error
    // remove components from a target
    Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error
    // get current component states from a target. The desired state is passed in as a reference
    Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error)
    // the target decides if an update is needed based the the current components and deisred components
    // when a provider re-construct state, it may be unable to re-construct some of the properties
    // in such cases, a provider can choose to ignore some property comparisions
    NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool
    // Provider decides if components should be removed
    NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool
}
```
To simplify developer development, Symphony assumes:

* A provider is thread-unsafe
* A provider is stateless
* A provider may or may not be idempotent

## Get()
The ```Get()``` method probes the target device and returns the currently installed components. Because a target device may host applications from different sources, the ```Get()``` method passes in the expected deployment description as a reference to help the provider to narrow down the scope of interest. For example, a mobile phone may have tens of applications installed. the deployment description informs the provider which application(s) should be checked. Specifically, the deployment description has a ```GetComponentSlice()``` method that returns the exact component list to be checked. 
A provider is expected to:

1. Read installed components.
2. Filter the components by the components in the ```GetComponentSlice()``` list.
3. Return the filtered component list.

## Apply()
The ```Apply()``` method deploys given components, which can be retrieved by calling the ```GetComponentSlice()``` method, to the target device. This method is only called after the ```NeedsUpdate()``` method has returned true. Hence, even if a provider is not idempotent, the side effect is minimized. A provider is expected to take necessary actions to bring its state as specified in the deployment description. 

> **NOTE**: A provider is expected to operate only on components affected by the deployment descriptions. Some providers may need to wipe the whole device during deployment. In such cases, the development description is expected to contain the entire software stack. Or, a device image can be treated as a single component in a solution.

## Remove()
The ```Remove()``` method removes given components, which can be retrieved by calling the ```GetComponentSlice()``` method, from the target device. This method is only called after the ```NeedsRemove()``` method has returned true. 

## NeedsUpdate()
The ```NeedsUpdate()``` method allows a provider to decide if an update operation is needed by comparing the current state and the desired state. Because Symphony allows providers to be stateless, a provider doesn't necessarily have capabilities to reconstruct the state as specified in the original desired state. For example, once an UWP app is installed on a Windows 10 machine, the package name is appended with a random postfix. Such random postfix is unknown to the original desired state and doing a straight property comparison will cause a false positive. With the ```NeedsUpdate()``` method, a provider can implement custom comparison logics to handle such discrepancies. 

## NeedsRemove()
The ```NeedsRemove()``` method allows a provider to decide if a remove operation is needed (see the [NeedsUpdate()](#needsupdate) method).
