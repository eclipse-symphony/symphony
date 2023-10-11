# Multisite Symphony Deployment

You can assemble multiple Symphony control planes to form a cascaded control plane tree. For example, if you have a HQ office and two site offices, you can set up Symphony control plane on all three locations and link site offices as children of the HQ control plane.

Once a control plane is linked to a parent control plane, it synchronizes ```Catalogs``` from the parent. And it can be influenced by a Campaign running on the parent control plane. One typical usage is to use the HQ office to control deployment of a standardized application across multiple site offices.

## Configuring a Parent Site

In Symphony API configuration file, you can specify the address and credentials for a parent site, as shown in the following configuration snippet.
When a child control plane launches, it automatically connects with the parent and registered itself as a ```site``` in the parent control plane.

> **NOTE**: In current versions there are no extra handshaking processes. The child control plane simply logs in to the parent control plane using the configured credential. In future versions, an extra handshaking process is planned for things like attestation and acquisition of an operational certificate before a child can be registered.


```json
{
  "siteInfo": {
    "siteId": "tokyo",
    "parentSite": {
      "baseUrl": "http://<symphony-service-ext>:8080/v1alpha2/",
      "username": "admin",
      "password": ""
    },
    "currentSite": {
      "baseUrl": "http://symphony-service:8080/v1alpha2/",
      "username": "admin",
      "password": ""
    }
  },
```

When installing Symphony with Helm, you can set the site id as well as the parent site URL/credentials using ```--set``` switches, for example:

```bash
helm install symphony oci://possprod.azurecr.io/helm/symphony --version 0.45.31 --set siteId=tokyo --set parent.url=http://<parent's symphony-service-ext IP>:8080/v1alpha2/
```
If a child is successfully connected to a parent site, you should see the site registration from the parent's context with:
```bash
kubectl get sites
```
## Synchronization with the Parent Site

Once a child site is connected to its parent, it starts to gradually copy down Catalog objects from the parent site. The local copy of the Catalog objects are prefixed with the original site id. For example, an ```app-config``` Catalog copied from a hq site is named as ```hq-app-config```.

A Catalog can then be “materialized” by a Campaign into “solid” Symphony objects like ```Solutions```, ```Targets``` and ```Instances```. 

This mechanism allows standardized templates, such as standardized applications, to be defined on a HQ, synchronized to site offices, and deployed locally. See the [Multisite app deployment scenario](../scenarios/multisite-deployment.md) for more details.


