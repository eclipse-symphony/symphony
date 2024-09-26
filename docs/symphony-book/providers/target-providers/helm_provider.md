# providers.target.helm

This provider manages a Helm chart embedded in a component. It supports packaged Helm charts (.tgz file) from either an OCI repository, or a direct download URL.

**ComponentSpec** Properties are mapped as the following:

**Component Type:** `helm.v3`

**Component Name:** mapped to Helm release name

| ComponentSpec properties| Helm provider|
|--------|--------|
| chart[name] | chart name |
| chart[repo] | chart repo or URL<sup>1</sup> |
| chart[version] | chart version<sup>2</sup>|
| `values` | chart values<sup>3</sup>|

1: The repo URL can be either an OCI repo address (without the `oci://` prefix), or a URL pointing to a packaged Helm chart (with `.tgz` file extension).

2: The chart version is ignored when full chart URL is used in the `helm.repo` property.

Find full scenarios at [this location](../../../samples/canary/solution.yaml)