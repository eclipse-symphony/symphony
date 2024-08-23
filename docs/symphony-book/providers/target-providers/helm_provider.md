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
| chart[wait] | wait<sup>3</sup>|
| chart[timeout] | timeout<sup>3</sup>|
| `values` | chart values|

1: The repo URL can be either an OCI repo address (without the `oci://` prefix), or a URL pointing to a packaged Helm chart (with `.tgz` file extension).

2: The chart version is ignored when full chart URL is used in the `helm.repo` property.

3. Waits until all Pods are in a ready state, PVCs are bound, Deployments have minimum (Desired minus maxUnavailable) Pods in ready state and Services have an IP address (and Ingress if a LoadBalancer) before marking the release as successful. It will wait for as long as the `timeout` value. 

Find full scenarios at [this location](../../../samples/canary/solution.yaml)