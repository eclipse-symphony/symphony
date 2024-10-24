# Ingress provider
The Ingress target provider manages Kubernetes Ingress embedded in components and converts component spec to a kubernetes Ingress object.

## Provider configuration
You can modify the provider configuration in the Symphony config file to specify the Kubernetes cluster for which the Ingress object should be created.

| Field | Comment |
|--------|--------|
| `configType` | Type of K8s configuration, either `path` or `inline`. |
| `configData` | Configuration data<sup>1</sup> |
| `inCluster` | If provider is running inside a K8s cluster (`"true"`). If `true`, `configType` and `configData` are not used. |

1: When `configType` is set to `path`, this property contains the path to a Kubernetes configuration file. If this property is left empty or omitted, the default Kubernetes configuration file on the host will be used. If `configType` is set to `inline`, this property contains the Kubernetes configuration bytes, as shown in the following Target spec:


```yaml
topologies:
- bindings:
  - role: ingress
    provider: providers.target.ingress
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
Here is a simple example for using ingress target provider in solution deployment. The below solution will create a kubernetes Ingress object in the kubernetes clusters for all the targets where the solution is invoked. 

```
apiVersion: solution.symphony/v1
kind: Solution
metadata: 
  name: ingress-solution
spec:  
  components:
  - name: minimal-ingress
    type: ingress
    metadata:
      annotations.nginx.ingress.kubernetes.io/rewrite-target: /
    properties:
      ingressClassName: nginx-example
      rules:
      - http:
          paths:
          - path: /testpath
            pathType: Prefix
            backend:
              service:
                name: test
                port:
                  number: 80        
```
Find full scenarios at [this location](../../../samples/canary/solution.yaml)