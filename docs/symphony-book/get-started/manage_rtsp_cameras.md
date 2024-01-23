# Symphony quickstart - Manage RTSP cameras connected to a gateway

In this scenario, you will:

* Set up a Symphony control plane on Kubernetes.
* Deploy a Symphony agent to your PC, and register your PC as a *target* on Symphony.
* Register an RTSP camera as a *device* and associate the device with the target.
* Set up an Azure Storage account that holds camera snapshots.

The Symphony agent will capture camera snapshots and upload them to the storage account. Then, you can query the latest snapshot URL through the `Device` object state.

![RTSP Cameras](../images/camera-management.png)

## 0. Prerequisites

* A Kubernetes cluster.
* [Helm 3](https://helm.sh/).
* [kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/), configured with the Kubernetes cluster you want to use as the default context.
* [Azure CLI](https://docs.microsoft.com/cli/azure/).
* **Azure Storage Account** and **Azure Service Principal**. For instructions on configuring these resources for the Symphony agent, see [Symphony agent](../agent/symphony-agent.md).
* An **RTSP camera** on your local network. If the camera is password protected, you'll need the corresponding user name and password. 
* **Symphony Agent** binary. For build instructions, see [Build Symphony containers](../build_deployment/build.md).
* **Symphony** deployed to your cluster. For instructions, see [Use Symphony on Kubernetes clusters with Helm](./quick_start_helm.md).

## 1. Create the camera object

Manually register your camera as a Symphony device by creating a `device` object using `kubectl`. You need to set the `ip` property, the `user` property, and the `password` property to match with your camera settings. Note that the camera is labeled to be connected to a `gateway-1` target, which you'll create next.

Create a YAML file called `camera-1.yaml` with the following device information:

```bash
apiVersion: fabric.symphony/v1
kind: Device
metadata:
  name: camera-1
  labels:
    gateway-1: "true"    
spec:
  properties:
    ip: "<RTSP camera IP>"    
    user: "<RTSP camera user>"
    password: "<RTSP camera password>"
```

This YAML file is also available at [docs/samples/k8s/device-management/camera-1.yaml](../../samples/k8s/device-management/camera-1.yaml).

Once you save the YAML file, apply it with `kubectl`:

```bash
kubectl create -f ./camera-1.yaml
```

## 2. Create the target object

Next, you'll create a `target` object that represents the gateway device. The association between the gateway and the camera is done through Kubernetes [selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).

Create a YAML file called `gateway-1.yaml` with the following target information:

```bash
apiVersion: fabric.symphony/v1
kind: Target
metadata:
  name: gateway-1
spec:
  properties:
    cpu: x64
    acceleration: "Nvidia dGPU"
    os: "Ubuntu 20.04"
```

This YAML file is also available at [docs/samples/k8s/device-management/gateway-1.yaml](../../samples/k8s/device-management/gateway-1.yaml).

Apply the gateway target using `kubectl`:

```bash
kubectl create -f ./gateway-1.yaml
```

## 3. Launch the Symphony agent as a process

To run the Symphony agent as a process, first set required environment variables and then launch the agent.

To get the Symphony API address, use kubectl:

```bash
kubectl get svc symphony-service-ext
```

Copy the `EXTERNAL-IP` field, as shown in the following sample:

```bash
NAME                   TYPE           CLUSTER-IP    EXTERNAL-IP    PORT(S)          AGE
symphony-service-ext   LoadBalancer   10.0.251.58   20.40.25.217   8080:31924/TCP   41m
```

Then, configure your environment variables:

```bash
export AZURE_CLIENT_ID=<service principal app id>
export AZURE_TENANT_ID=<service principal tenant id>
export AZURE_CLIENT_SECRET=<service principal client secret>
export STORAGE_ACCOUNT=<storage account name>
export STORAGE_CONTAINER=<storage container name>
export SYMPHONY_URL=http://<symphony API address>:8080/v1alpha2/agent/references # point to your local Symphony API endpoint, or the public Symphony API service endpoint on K8s
export TARGET_NAME=<target name> #the name of the Target object representing the current compute device
```

Now, you can launch the Agent:

```bash
./symphony-agent -c ./symphony-agent.json -l Debug
```

## 4. Verify the results

Once the agent is running, you should see a snapshot image saved in your storage container (named `camera-1-snapshot.jpg` in this case).

Get the latest snapshot URL from device status:

```bash
kubectl get device camera-1 -o yaml
```

Observe the snapshot URL:

```bash
apiVersion: fabric.symphony/v1
kind: Device
metadata:
  ...
spec:
  ...
status:
  properties:
    snapshot: https://voestore.blob.core.windows.net/snapshots/camera-1-snapshot.jpg
            # ^--- snapshot URL is here
```
