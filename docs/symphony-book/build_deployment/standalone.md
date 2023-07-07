# Running Symphony in Standalone Mode

_(last edit: 6/27/2023)_

## Overview
Symphony runs as a Kubernetes controller that operates Symphony CRDs. When running as a Kubernetes controller, Symphony delegates all state management of these CRDs to Kubernetes, and it uses webhooks to invoke Symphony API when the state of these objects change. So, Symphony Kubernetes controller is a very thin layer hooking up to Kubernetes API server, while all Symphony business logic resides behind Symphony API. Using the same architecture, we allow Symphony to run natively as an Azure ARM Resource Provider (RP) that forwards external resource operations (such as update and remove) to Symphony API, as shown in the following diagram.
![Architecture](../images/architecture.png)

When running in a standalone mode, Symphony takes over object state management by itself. Symphony allows object state to be stored in any supported data storage via a State Provider. And it runs a state reconciliation loop that periodically generate a reconciliation job. Symphony listens to these jobs and invokes corresponding APIs for state reconciliation. This separation of job creation and job execution allows Symphony to handle both automatic reconciliation as well as on-demand reconciliation using the same Job Manager.   

## State reconcilation flows

Under Kubernetes mode, users interact with Symphony Kubernetes CRDs using Kubernetes APIs and tools. Symphony controller queries Symphony API (via ```/queue GET```) to decide if a new reconciliation job needs to be queued. If so, it queues a new job through the ```/queue POST``` route. The queued job is forward to a event bus, to which a [Job Vendor](../vendors/job.md) subscribes. When a reconcile job is received, the Job Vendor generates a Symphony deployment spec and calls the [Solution Vendor](../vendors/solution.md)'s ```/reconcile``` route. 

When running under the standalone mode, Symphony REST API queues reconciliation jobs when it updates Symphony objects. In addition, the Job Vendor is configured to peridocially trigger reconcilation events by itself at configured intervals. 

![no-k8s](../images/no-k8s.png)



## Launching Symphony in standalone mode
Symphony runs as a single process in standalone mode. To launch Symphony in standalone mode, simply launch the ```symphony-api``` binary with a configuration file and an optional tracing level switch (such as ```Error```, ```Info``` and ```Debug```):
```bash
./symphony-api -c ~/symphony-api-no-k8s.json -l Debug
```
Once launched, you can access Symphony's [REST API](../api/api.md) using any web clients.

## Next

* [Quick start - launching a Redis container with standalone Symphony](../quick_start/deploy_redis_no_k8s.md)