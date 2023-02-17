# providers.target.k8s
This is a Kubernetes native target provider that translates solution component to Kubernetes native types. 

## Deployment Strategy

K8s Target provider supports three deployment strategies:

* **Single Pod:** All components in a solution is deployed into a single pod.
* **Services:** Each component is deployed as a seprate service in a user-specified namespace.
* **Services with instance namespace:** Each component is deployed as a separate service in a namespace named after the instance name.

## Single Pod Strategy

### Component Property Mappings

K8s Provider maps a **ComponentSpec** to a ```Deployment.Spec.Template.Spec.Containers[i]``` (refered as ```Container``` in the following tables). When **Single Pod** strategy is used, **InstanceSpec** metadata is mapped to K8s deployment attributes; For other cases, **ComponentSpec** metadata is used.
![K8s Provider Single Pod Strategy](../images/k8s-provider-single-pod-strategy.png)

| Symphony Instance Object | K8s Deployment |
|--------|--------|
|```Metadata["deployment.imagePullSecrets"]```|```Deployment.Spec.Template.Sepc.ImagePullSecrets```|
|```Metadata["deployment.nodeSelector"]```|```Deployment.Spec.Template.Spec.NodeSelector```|
|```Metadata["deployment.replicas"]```|```Deployment.Spec.Replicas```|
|```Metadata["deployment.scope"]```|```Deployment.ObjectMeta.Namespace```|
|```Metadata["deployment.volumes"]```|```Deployment.Spec.Template.Spec.Volumes```|
|```Metadata["service.annotation.<label>]```|```Service.ObjectMeta.Annotations[<label>]```|
|```Metadata["service.loadBalancerIP]```|```Service.Spec.LoadBalancerIP```|
|```Metadata["service.ports]```|```Service.Spec.Ports```|
|```Metadata["service.type]```|```Service.Spec.Type``` (default is ```ClusterIP```)|

**ComponentSpec** Properties are mapped as the following:

| ComponentSpec Properties | K8s Provider |
|--------|--------|
|```ComponentSpec.Name```| ```Container.Name``` and ```Service.ObjectMeta.Name```|
|```Properties["container.args"]```|```Container.Args```|
|```Properties["container.commands"]```|```Container.Command```|
|```Properties["container.createOptions"]```|---|
|```Properties["container.image"]```|```Container.Image```|
|```Properties["container.imagePullPolicy"]```|```Container.ImagePullPolicy``` (default is ```Always```)|
|```Properties["container.ports"]```|```Container.Ports```|
|```Properties["container.resources"]```|```Container.Resources```|
|```Properties["container.restartPolicy"]```|---|
|```Properties["container.type"]```|---|
|```Properties["container.version"]```|---|
|```Properties["container.volumeMounts"]```|```Container.VolumeMounts```|
|```Properties["desired.<property>"]```|---|

## Services Strategy