# Versioning

> **NOTE:** Planned feature (P0). Versioning experience can be emulated with the current API. The versioned API provides native versioning supports.

Symphony follows the immutable infrastructure paradigm, where any manipulation of the desired state warrants a new version. And because in Kubernetes and ARM, objects are uniquely keyed, Symphony uses a naming convention to indicate different versions of an object, such as “`app-v-version1`” and “`app-v-version2`” under "`app`" container object.

Because some customers have expressed desire to use a versioned API instead of naming conventions, Symphony is adding versioned APIs to create such experiences on top of the above mechanism. With the versioned API, a user can operate on multiple versions of the same object, though underneath Symphony still uses the naming conventions to satisfy the unique key requirements of the platforms.

## Create a versioned object - Symphony API standalone 

Symphony introduces a series of new versioned objects, such as versioned solutions, versioned catalogs and versioned campaigns, on top of the existing objects. Container objects hold multiple versions. For example, a “`my-config`” object may hold multiple versions of a configuration Catalog objects.

To create a versioned object in a standalone mode, a user needs to send below requests:
1. Create a container object, for example solution container using Create API in the [solution container API doc](../api/solutioncontainers-api.md)
2. Create a versioned resource, for exmaple solution using Create API in the [solution API doc](../api/solutions-api.md)

<!-- TODO: add back when containers/versions APIs are supported.
```bash
/<versioned-objects>/<versioned object id>/versions/<id>
```
And the following table summarizes different queries to be carried out:
| Path | Queries|
|--------|--------|
| `/<versioned-objects>` | List of the versioned objects |
| `/<versioned-objects>/<versioned-object id>` | Get a specific versioned object|
| `/<versioned-objects>/<versioned-object id>/versions` | List all versions (individual objects) |
| `/<versioned-objects>/<versioned-object id>/versions/<id>` | Get a specific version |

For example, to get `v3` of a `my-config`, uses:
```bash
/versioned-catalogs/my-config/versions/v3
```  -->

## Create a versioned object - Kubernetes mode
Under Kubernetes mode, it is required to create a `container` object before creating a versioned object. For example, to create a solution `myapp-v1`, customer needs to create a `solutioncontainer` object and then `solution` object (yaml files available at [docs/samples/k8s/hello-world/solution.yaml](../../samples/k8s/hello-world/solution.yaml).)

## Naming convention 
When creating container and version objects, version resource creation should follow the rules bother in a standalone mode or K8s mode:
1. A version object should follow the naming convention: `<container name>-v-<version name>`.
2. A version object should `spec.rootResource` which is the name of corresponding container object name. The container object should be created beforehand and exist when version object is created.

## Referencing a versioned object
When referencing a specific version of a versioned object, you can use a “:<version>” postfix in you Symphony expressions, such as:

```yaml
${{$config('my-config:version3', 'my-field')}}
```

> **NOTE:** This syntax is to be expanded in the future to include cross-cluster and cross-namespace references, such as `<cluster>/<namespace>/<object>:<version tag>`.

## Version scheme

Symphony doesn’t impose a single versioning scheme. A user can choose to use a simple version number, semantic versioning (major.minor.patch) or any other versioning patterns. These versions are stored in the versioned objects as a list. When going back and forth in versions, Symphony simply accesses list items at different indices. 

> **NOTE:** A custom experience built on top of Symphony may want to choose a specific versioning scheme and provides some additional assistance in versioning. That’s out of the scope of Symphony itself.


