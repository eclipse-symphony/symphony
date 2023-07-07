# Campaign Management

## Stages
The instance object has a top-level property named **Stage**, which can be used to control different stages of an instance deployment. This property enables scenarios like [gated deployments](../scenarios/gated-deployment.md), [staged deployments](../scenarios/staged-development.md), [scheduled deployment](../scenarios/scheduled-development.md) as well as [canary deployments](../scenarios/canary-deployment.md).

### Gated deployment
When the stage property is set to a special value **BLOCK**, deployment of the instance is blocked until this property is cleared, or changed to some other values. For example, the following instance will not be deployed to the target until its stage property is changed out of **BLOCK**.

```yaml
apiVersion: solution.symphony/v1
kind: Instance
metadata:
  name: my-instance
spec:
  stage: BLOCK
  solution: my-app
  target:
    name: my-target
      
```

### Staged deployment
In an instance object, you can define any number of stages, and each stage can be associated with a different [Solution](../uom/solution.md) and a different set of [Target](../uom/target.md)s. The following instance defines three stages: ```ring0```, ```ring1``` and the default stage (activated when the stage property is empty or set to ```default```). The instance is currently blocked. By changing the stage property, you can control the deployment to be ```ring0```, ```ring1```, or ```default```. 

```yaml
apiVersion: solution.symphony/v1
kind: Instance
metadata:
  name: my-instance
spec:
  stage: BLOCK
  solution: my-app
  target:
    name: my-target
  stages:
  - name: ring0
    solution: my-app-v1
    target:
      selector:
        label: ring0
  - name: ring1
    solution: my-app-v1
    target:
      selector:
        label: ring1
```

### Canary deployment
Canary deployment is a specialized staged deployment. In a stage, you can define multiple solution version with assigned weights. And Symphony will distribute the versions to the matching targets based on assigned weights. For example, the following canary deployment sends ```my-app-v1``` to 30% of the matching targets, and ```my-app-v2``` to 70% of the matching targets. Obviously, you can define multiple canary deployment stages and switch among them as needed, and eventually get out of canary by switching to the ```default``` stage.

```yaml
apiVersion: solution.symphony/v1
kind: Instance
metadata:
  name: my-instance
spec:
  stage: canary
  solution: my-app
  target:
    name: my-target
  stages:
  - name: canary
    versions:
    - solution: my-app-v1
      weight: 30
    - solution: my-app-v2
      weight: 70
    target:
      selector:
        label: ring0  
```

### Scheduled deployment
The **Stage** property is used in conjunction with the **Schedule** property in this case. To use scheduled deployment, you create an instance object with **BLOCK** stage and a scheduled date/time. Symphony will schedule a job to clear the stage property at given time, allowing the instance to be deployed at specified time. 
