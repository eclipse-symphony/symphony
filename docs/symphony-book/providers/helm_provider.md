# providers.target.helm

This provider manages a Helm chart embedded in a component. It supports packaged Helm charts (.tgz file) from either an OCI repository, or a direct download URL.

**ComponentSpec** Properties are mapped as the following:

**Component Type:** `helm.v3`

**Component Name:** mapped to Helm release name

| ComponentSpec properties| Helm provider|
|--------|--------|
| `helm.chart.name` | chart name |
| `helm.repo` | chart repo or URL<sup>1</sup> |
| `helm.chart.version` | chart version<sup>2</sup>|
| `helm.values.*` | chart values<sup>3</sup>|

1: The repo URL can be either an OCI repo address (without the `oci://` prefix), or a URL pointing to a packaged Helm chart (with `.tgz` file extension)

2: The chart version is ignored when full chart URL is used in the `helm.repo` property.

3:  To define override values, add values with a `"helm.values."` prefix to your ComponentSpec properties. For example, to override a `CUSTOM_VISION_KEY` value, add `helm.values.CUSTOM_VISION_KEY` to your component properties.
