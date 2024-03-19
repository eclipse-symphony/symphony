# Versioning

> **NOTE:** Planned feature (P0). Versioning experience can be emulated with the current API. The versioned API provides native versioning supports.

Symphony follows the immutable infrastructure paradigm, where any manipulation of the desired state warrants a new version. And because in Kubernetes and ARM, objects are uniquely keyed, Symphony uses a naming convention to indicate different versions of an object, such as “`app-v1`” and “`app-v2`”.

Because some customers have expressed desire to use a versioned API instead of naming conventions, Symphony is adding versioned APIs to create such experiences on top of the above mechanism. With the versioned API, a user can operate on multiple versions of the same object, though underneath Symphony still uses the naming conventions to satisfy the unique key requirements of the platforms.

## Versioned API syntax

Symphony introduces a series of new versioned objects, such as versioned solutions and versioned catalogs, on top of the existing objects. Versioned objects are containers that hold multiple versions. For example, a “`my-config`” object may hold multiple versions of a configuration Catalog objects.
In general, a versioned API has the following routes following a typical REST pattern:

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
```

## Referencing a versioned object
When referencing a specific version of a versioned object, you can use a “:<version>” postfix in you Symphony expressions, such as:

```yaml
${{$config(‘my-config:v3’, ‘my-field’)}}
```

> **NOTE:** This syntax is to be expanded in the future to include cross-cluster and cross-namespace references, such as `<cluster>/<namespace>/<object>:<version tag>`.

## Version scheme

Symphony doesn’t impose a single versioning scheme. A user can choose to use a simple version number, semantic versioning (major.minor.patch) or any other versioning patterns. These versions are stored in the versioned objects as a list. When going back and forth in versions, Symphony simply accesses list items at different indices. 

> **NOTE:** A custom experience built on top of Symphony may want to choose a specific versioning scheme and provides some additional assistance in versioning. That’s out of the scope of Symphony itself.


