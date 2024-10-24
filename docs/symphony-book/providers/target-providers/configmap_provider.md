# ConfigMap provider
The ConfigMap target provider manages Kubernetes ConfigMaps embedded in components and converts key-value pairs declared in those components to be stored in a ConfigMap.

## Provider configuration
You can modify the provider configuration in the Symphony config file to specify the Kubernetes cluster for which the ConfigMap should be created.

| Field | Comment |
|--------|--------|
| `configType` | Type of K8s configuration, either `path` or `inline`. |
| `configData` | Configuration data<sup>1</sup> |
| `inCluster` | If provider is running inside a K8s cluster (`"true"`). If `true`, `configType` and `configData` are not used. |

1: When `configType` is set to `path`, this property contains the path to a Kubernetes configuration file. If this property is left empty or omitted, the default Kubernetes configuration file on the host will be used. If `configType` is set to `inline`, this property contains the Kubernetes configuration bytes, as shown in the following Target spec:


```yaml
topologies:
- bindings:
  - role: configmap
    provider: providers.target.configmap
    config:
      inCluster: "false"
      configType: "inline"
      configData: |
        apiVersion: v1
        clusters:
        - cluster:
          certificate-authority-data: ...
          server: https://k8s-dns-7c476efd.hcp.westus3.azmk8s.io:443
        name: k8s
      contexts:
      - context:
          cluster: k8s
          user: clusterUser_symphony_k8s
        name: k8s
      current-context: k8s
      kind: Config
      preferences: {}
      users:
      - name: clusterUser_symphony_k8s
        user:
          client-certificate-data: ...
          client-key-data: ...
          token: ...
```

## Solution component example
Here is a simple example for using configMap target provider in solution deployment. The below solution will create two configMaps in the kubernetes cluster for all the targets where the solution is invoked. The first one will create a configMap named configmap1 with key-value pair {"tags", "this is configmap1"} in the data. And it's similar for the second one.

```
apiVersion: solution.symphony/v1
kind: Solution
metadata:
  name: examples
spec:
  components:
    - name: configmap1
      type: config
      properties:
        tags: "this is configmap1"
    - name: configmap2
      type: config
      properties:
        tags: "this is configmap2"
```
More samples can be found [here](../../../samples/configuration/)