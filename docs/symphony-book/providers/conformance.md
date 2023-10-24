# Target provider conformance

_(last edit: 6/28/2023)_

The Symphony object model allows arbitrary properties to be attached to a component, and Symphony tries to minimize its own requirements on these properties. If you want to take advantage of some Symphony features like adaptive deployment, you should use the standardized properties. Otherwise, you are free to attach any properties, including artifact bytes in original formats, to component properties. This gives Symphony great flexibility to accommodate new provider types without forcing users to adapt to a new artifact format.

One challenge of this approach is how to provide predictable, consistent behaviors across providers. To achieve this, Symphony requires each provider to explicitly define what properties it needs to operate. For example, a provider that handles container-based workloads may declare that it requires a `container.image` property to be present to handle the component correctly. 

In addition, Symphony defines a few conformances levels that are designed to help users to pick desired providers to use. For example, if a user wants to manage container-based workloads, she can pick a container-conformant provider and deploy her solution to any supported systems like Docker, Kubernetes and Azure IoT Edge.

| Provider |Container simple|Container standard|Container service| Scope | Isolation|
|--------|--------|--------|--------|--------|--------|
| `providers.target.adb` ||||||
| `providers.target.azure.adu` ||||||
| `providers.target.azure.iotedge` |YES||||YES|
| `providers.target.docker`|YES|YES|||YES|
| `providers.target.helm`||||?|YES|
| `providers.target.k8s` |YES|YES|YES|YES|YES|
| `providers.target.kubectl`||||?|YES|
| `providers.target.win10`||||||

Excluded providers:

* `providers.target.mqtt` and `providers.target.proxy` are proxy providers. The exact behavior depends on the destination provider.
* `providers.target.script` relies on custom implementations.
* `providers.target.staging` stages artifacts on objects without actual deployments.
* `providers.target.http` is a webhook provider. The result depends on the destination web server.

## Container simple

*Container simple* compliant providers can deploy a Docker container specified in the `container.image` property.

## Container standard

*Container standard* compliant providers can deploy a Docker container specified in the `container.image` property. They also support these optional container properties:

    * `container.ports`
    * `container.commands`
    * `container.resources`
    * `container.volumeMounts`
    * `env.*` (environment variables)

## Container service

*Container service* compliant providers can deploy a Docker container specified in the `container.image` property. They also support the following optional service metadata:

    * `service.ports`
    * `service.type`
    * `service.name`
    * `service.loadBalancerIP`
    * `service.annotation.*`

    And it also supports these optional container properties:
    * `container.imagePullPolicy`
    * `container.ports`
    * `container.args`
    * `container.commands`
    * `container.resources`
    * `container.volumeMounts`
    * `env.*` (environment variables)

## Scope

A provider that supports scopes can place components in designated scopes, such as Kubernetes namespaces or Azure ARM resource groups.

## Isolation

When using a provider that supports workload isolation, you can safely deploy multiple instance objects on the same physical target. The provider will take necessary steps to make sure these instances don't interfere with each other. On some systems, this may require the provider to add prefixes or postfixes to instance names to avoid conflicts. In such cases, the provider can inject updated instance names (through environment variables, for instance).

## Change Detection

As part of the `ValidationRule`, a provider can explicitly specify how it would like to detect changes. It can declare what properties will be used in change detection, and how they are compared. For example, the following rule defines that if any environment variable changes, the component is considered changed. Newly defined environment variables or removed environments will be ignored (set `SkipIfMissing` to `true` for a strict match):

```json
{Name: "env.*", IgnoreCase: false, SkipIfMissing: true}
```
