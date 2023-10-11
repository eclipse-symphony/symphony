# Deploy Symphony

_(last edit: 6/28/2023)_

## Option 1: Using maestro

Maestro is Symphony’s CLI that allows you to bootstrap Symphony with your Kubernetes clusters, or to run latest Symphony build in standalone mode. It also allows you to quickly deploy sample scenarios using prebuilt samples. It’s a great tool to get started with Symphony quickly! Please see maestro instructions [here](../cli/cli.md).

## Option 2: Using Helm
You can also deploy Symphony to a Kubernetes cluster using Helm 3:
```bash
helm install symphony oci://possprod.azurecr.io/helm/symphony --version 0.45.31
```
Or, to update your existing Symphony release to a new version:
```bash
helm upgrade --install symphony oci://possprod.azurecr.io/helm/symphony --version 0.45.31
```

## Option 3: Using Docker
You can run Symphony API in standalone mode as a Docker container 
```bash
docker run --rm -it  -v /configuration/file/path/on/host:/config -e CONFIG=/config/symphony-api-no-k8s.json possprod.azurecr.io/symphony-api:latest
```
> **NOTE**: You can find various Symphony API configuration files under the ```api``` folder of the Symphony repo. Please see [this doc](../hosts/overview.md) for details on different configurations you can use.

## Deployment at scale
At the time of writing, the default Symphony configuration uses in-memory state stores and pub/sub message buses. To deploy Symphony at scale, you'll need to choose a different state store and pub/sub message bus such as Cosmos DB and Redis. 

### Scaling out
By default, all Symphony vendors are hosted on a single [Host](../hosts/overview.md). If you need to scale these vendors independently, you can create multiple host configurations, each loading only the desired vendors, and run multiple host processes or containers in your environment. Because Symphony doesn't allow horizontal dependencies, you can slice up vendors into different topologies freely. However, for these vendors to communicate with each other through messaging, they need to share the same pub/sub message bus, such as a Redis cluster.

Symphony's [Job Manager](../managers/overview.md) invokes Symphony's reconcile API on the [Solution Vendor](../vendors/solution.md) through HTTP. So, you'll need to make sure your Job Manager is configured to talk to the Solution Vendor host FDN (or load-balancer FDN) instead of ```localhost```.

### State stores
Most Symphony components are stateless, with exception of the [Instance Manager](../managers/instance-manager.md). Instance Manager uses a state store to remember the last deployment it has successfully applied. When you have multiple Instance Managers running (by scaling out the [Solution Vendor](../vendors/solution.md)), they need to use a shared state store instead of the in-memory state store.

In addition to the default in-memory store (which doesn't scale beyond a single process), Symphony also supports a HTTP-proxy store, through which you can connect to [most of the popular databases](https://docs.dapr.io/reference/components-reference/supported-state-stores/) via [Dapr](https://dapr.io/). 

### Pub/sub
If you host vendors on multiple processes or containers, you need to ensure that these vendors share the same pub/sub message bus, such as a Redis cluster. 

> **NOTE**: By default, Symphony deploys a Reids pod as its pub/sub backbone.

Symphony is extensible to support additional state stores and pub/sub message buses through its [Providers](../providers/overview.md) mechanism.

## Customize Helm Deployment

The following table summarizes some Helm chart value switches you can use during Symphony deployment with Helm (by adding ```--set <key>=<value>``` switches):

| Value | Description|
|--------|--------|
| ```installServiceExt``` | when set to ```true``` (default), Symphony deploys a publicly accessible ```symphony-service-ext``` service for agents and child sites. If you don't have such needs, you can turn off this service by setting the value to ```false```. |
| ```redis.enabled``` | When set to ```true``` (default), Symphony deploys a Redis pod as its pub/sub messaging backbone. If you turn this to ```false```, an in-memory backbone is used, which can't be scaled beyond a single API pod. |
| ```siteId``` | You can change the default site id, which is ```hq```, to a different value with this switch. |
| ```parent.url``` | When this value is set, the current Symphony control plane is linked to a parent Symphony control plane. See [here](./multisite-deploy.md) for more details.|




