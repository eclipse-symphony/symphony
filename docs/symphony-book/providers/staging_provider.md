# Staging Provider

Staging provider allows [components](../uom/solution.md#componentspec) to be recorded on an [Target](../uom/target.md) object without being deployed to the actual target. This allows components to be **staged** and retrieved later, such as by a polling agent. 

## Provider Configuration

| Field | Comment |
|--------|--------|
| ```configType``` | Type of K8s configuration<sup>1</sup> |
| ```configData``` | Configuration data<sup>2</sup> |
| ```inCluster``` | If provider is running inside a K8s cluster (```"true"```). If ```true```, ```configType``` and ```configData``` are not used. |
| ```singleSolution``` | If only one solution can be staged on the default. Default is ```true```<sup>3</sup>. | 
| ```targetName``` | Name of the target |

<sup>1</sup>: Support types: ```path``` or ```bytes```. 

<sup>2</sup>: When ```configType``` is set to ```path```, this property contains the path to a Kubernetes configuration file. if this property is left empty or omitted, the default Kubernetes configuration file on the host will be used. If ```configType``` is set to ```bytes```, this property contains the Kubernetes configuration bytes, as shown in the following Target spec:
```
  topologies:
  - bindings:
    - role: instance
      provider: providers.target.k8s
      config:
        inCluster: "false"
        configType: "bytes"
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
<sup>3</sup>: When ```singleSolution``` is set to ```true```, staing a new solution to an target wipes all previously stated components. Otherwise, components from the solution are merged into the currently staged component list.