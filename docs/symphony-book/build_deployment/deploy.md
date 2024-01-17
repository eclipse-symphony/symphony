# Deploy Symphony

_(last edit: 6/28/2023)_

Choose your preferred tool for deploying Symphony:

* **Maestro**: Maestro is Symphony’s CLI that allows you to bootstrap Symphony with your Kubernetes clusters, or to run latest Symphony build in standalone mode. It also allows you to quickly deploy sample scenarios using prebuilt samples. It’s a great tool to get started with Symphony quickly!

  [Use Symphony with the Maestro CLI tool](../get-started/quick_start_maestro.md).

* **Helm**: You can deploy Symphony to a Kubernetes cluster using Helm 3.

  [Use Symphony on Kubernetes clusters with Helm](../get-started/quick_start_helm.md)

* **Docker**: You can run Symphony API in standalone mode as a Docker container.

  [Use Symphony in a Docker container](../get-started/quick_start_docker.md)

* **Binary**: You can build Symphony from a binary.

  [Use Symphony as a binary](../get-started/quick_start_binary.md)

> **NOTE**: You can find various Symphony API configuration files under the `api` folder of the Symphony repo. For more information, see [host configurations](../hosts/overview.md).

## Deployment at scale

The default Symphony configuration uses in-memory state stores and pub/sub message buses. To deploy Symphony at scale, choose a different state store and pub/sub message bus such as Cosmos DB and Redis.

### Scale out

By default, all Symphony vendors are hosted on a single [host](../hosts/overview.md). If you need to scale these vendors independently, you can create multiple host configurations, each loading only the desired vendors, and run multiple host processes or containers in your environment. Because Symphony doesn't allow horizontal dependencies, you can slice up vendors into different topologies freely. However, for these vendors to communicate with each other through messaging, they need to share the same pub/sub message bus, such as a Redis cluster.

Symphony's [job manager](../managers/overview.md) invokes Symphony's reconcile API on the solution vendor through HTTP. Make sure that your job manager is configured to talk to the solution vendor host FDN (or load-balancer FDN) instead of `localhost`.

### State stores

Most Symphony components are stateless, with exception of the [instance manager](../managers/instance-manager.md). The instance manager uses a state store to remember the last deployment it has successfully applied. When you have multiple instance managers running (by scaling out the solution vendor), they need to use a shared state store instead of the in-memory state store.

In addition to the default in-memory store (which doesn't scale beyond a single process), Symphony also supports an HTTP-proxy store through which you can connect to [most of the popular databases](https://docs.dapr.io/reference/components-reference/supported-state-stores/) via [Dapr](https://dapr.io/).

### Pub/sub

If you host vendors on multiple processes or containers, you need to ensure that these vendors share the same pub/sub message bus, such as a Redis cluster.

> **NOTE**: By default, Symphony deploys a Redis pod as its pub/sub backbone.

Symphony is extensible to support additional state stores and pub/sub message buses through its [providers](../providers/_overview.md) mechanism.
