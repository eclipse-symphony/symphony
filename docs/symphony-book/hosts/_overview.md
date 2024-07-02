# Hosts

_(last edit: 3/14/2024)_

Hosts are configurable hosting processes that load a group of configured [vendors](../vendors/_overview.md) and bind them to one or multiple [bindings](../bindings/_overview.md). Because Symphony vendors are independent, you can split Symphony vendors into multiple hosts and scale them independently.

## Host configurations

Under the `api` folder of the `symphony` repo, several `*.json` files define the configurations for various Symphony roles and purposes. For example:

* `symphony-target-agent.json`: This is the default configuration for a Symphony target agent.
* `symphony-api-dev-console-trace.json`: This is the same configuration as `symphony-api-dev.json`, plus a console exporter for [OpenTelemetry](https://opentelemetry.io/).
* `symphony-api-dev-zipkin-trace.json`: This is the same configuration as `symphony-api-dev.json`, plus a [Zipkin](https://zipkin.io/) exporter for [OpenTelemetry](https://opentelemetry.io/).
* `symphony-api-dev.json`: This is a Symphony API configuration for your local tests. When you run a Symphony API process outside a Kubernetes cluster, you should use this configuration file, for example:

  ```bash
  ./symphony-api -c ./symphony-api-dev.json -l Debug
  ```

* `symphony-api-no-k8s.json`: This configuration loads Symphony API with in-memory state providers. This is for local testing only.
* `symphony-api.json`: This is the default configuration for the Symphony API container. It loads all vendors that support the entire Symphony API surface.
* `symphony-script-proxy.json`: A sample proxy deployment with a script provider.
* `symphony-win-proxy.json`: A sample proxy deployment with a Windows 10 sideload provider.

A host configuration contains an `api` element that contains all the vendors to be loaded and a `bindings` element that contains all the bindings to be enabled for this host.

  ### state provider configuration

  [Manager section](../managers/_overview.md) tells what kind of state provider is needed for different managers. Below is an example of loading a Kubernetes state provider as providers.persistentstate for the targets manager. In this context, providers.persistentstate signifies that the targets manager expects to load a persistent state store. However, Symphony doesn’t explicitly forbid users from loading a volatile state provider (such as memory) where a persistent one is expected. Users should be aware that this could lead to crash-consistency issues for the objects.
  ```
  "managers": [
          {
            "name": "targets-manager",
            "type": "managers.symphony.targets",
            "properties": {
              "providers.persistentstate": "k8s-state"
            },
            "providers": {
              "k8s-state": {
                "type": "providers.state.k8s",
                "config": {
                  "inCluster": true
                }
              }
            }
          }
        ]
  ```

## Symphony-API container

By default, the `ghcr.io/eclipse-symphony/symphony-api` container is configured to load `symphony-api.json` with `Debug` log level (this may change in production container build). You can override log level with a `LOG_LEVEL` environment variable, and the configuration file with a `CONFIG` environment variable. For example, to change log level to `Error` while launching the container:

```bash
docker run --rm -it -e LOG_LEVEL=Error ghcr.io/eclipse-symphony/symphony-api:latest
```

And to use a different configuration file:

```bash
docker run --rm -it  -v /configuration/file/path/on/host:/config -e CONFIG=/config/symphony-api-dev.json ghcr.io/eclipse-symphony/symphony-api:latest
```

## Scale out the host

When you run multiple host instances behind a load balancer, and if you have [managers](../managers/overview.md) who use a state store, you need to choose a shared state store that is accessible by all instances. Symphony currently doesn't have a shared state store provider other than a HTTP state provider that can be configured together with sidecars like [Dapr](https://dapr.io/). It's expected some native shared state store provider (like Redis) will be added in future versions.
