# Symphony agent

A Symphony agent runs on a `target` and provides several services to Symphony payloads running on the same target, including:

* Get object references from the control plane.
* Probe and report on health of associated `device` objects.
* Capture and upload camera images for camera `device` objects.

A Symphony agent is a microservice that exposes an HTTP endpoint to Symphony payloads. We offer a Symphony container (`ghcr.io/azure/symphony/symphony-agent`) as well as a cross-platform binary that can be configured as a system daemon or service.

## Prepare for Symphony agent deployment

In this example, the Symphony agent needs a service principal to access an Azure Storage account to upload camera snapshots. In the current version of Symphony, we support service principal with a secret.

1. Create a service principal. For more information, see [Azure SDK for GO authentication with a service principal](https://learn.microsoft.com/azure/developer/go/azure-sdk-authentication-service-principal?tabs=azure-cli).

    ```azurecli
    # create resource group
    az group create --name <resource group name> --location <location> 
    # create key vault
    az keyvault create --location <location> --name <key vault name> --resource-group <resource group name>
    # create service principal
    az ad sp create-for-rbac --name <service principal name> --role Contributor --scope /subscriptions/<subscription id>/resourceGroups/<resource group name>
    ```

    Copy the `password` output before closing your terminal window.

    > **NOTE:** we'll support certificate-based principals in future versions.

2. If needed, create an Azure Storage account and a container for storing the snapshots:

   ```bash
   az storage account create --name <storage account name> --resource-group <resource group name> --location <location> --sku Standard_LRS
   az storage container create -n snapshots --account-name <storage account name>
   ```

3. If you plan to run Symphony agent as a process or a service, install [FFmpeg](https://ffmpeg.org/) on your target machine. You can skip this step if you plan to run Symphony agent as a container, which has FFmpeg pre-installed.

   ```bash
   sudo apt update
   sudo apt install ffmpeg
   # verify installation
   ffmpeg -version
   ```

## Option 1: Run Symphony agent as a container

To test Symphony agent on your local dev machine, you can use the prebuilt container:

```bash
docker run -p 8088:8088 -e SYMPHONY_URL=http://<Symphony control plane endpoint>:8080/v1alpha2/agent/references -e AZURE_CLIENT_ID=<service principal app id> -e AZURE_TENANT_ID=<service principal tenant id> -e AZURE_CLIENT_SECRET=<service principal client secret> -e STORAGE_ACCOUNT=<storage account name> -e STORAGE_CONTAINER=<storage container name> -e TARGET_NAME=<target name> hbai/symphony-agent:0.1.26
```

Where `<Symphony control plane endpoint>` is the DNS/IP of Symphony control plane endpoint. For example, when you run Symphony control plane on a Kubernetes cluster, the control plane exposes a load-balanced service endpoint for agents. You can get the service endpoint with:

```bash
kubectl get svc symphony-service-ext #use returned EXTERNAL-IP to connect
```

## Option 2: Run Symphony agent as a process

To run Symphony agent as a process, you need to set required environment variables first, and then launch the agent:

```bash
export AZURE_CLIENT_ID=<service principal app id>
export AZURE_TENANT_ID=<service principal tenant id>
export AZURE_CLIENT_SECRET=<service principal client secret>
export STORAGE_ACCOUNT=<storage account name>
export STORAGE_CONTAINER=<storage container name>
export SYMPHONY_URL=http://<symphony API address>:8080/v1alpha2/agent/references # point to your local Symphony API endpoint, or the public Symphony API service endpoint on K8s
export TARGET_NAME=<target name> #the name of the Target object representing the current compute device

./symphony-agent -c ./symphony-agent.json -l Debug
```

## Get object reference

You can get Symphony object specs, such as AI [skill](../concepts/unified-object-model/ai-skill.md) and [solution](../concepts/unified-object-model/solution.md), through the Symphony agent:

* **Route**: `http://<Symphony agent endpoint>:8088/v1alpha2/agent/references`
* **Method**: GET
* **Parameters**:

  | Parameter | Comment |
  |--------|--------|
  | alias | AI Skill alias<sup>1</sup>|
  | field-selector | Field selector (optional), for example: `metadata.name=redis-server`|
  | group | Resource group, like `ai.symphony`, `solution.symphony` and `fabric.symphony`|
  | id | Resource name (optional)|
  | instance | Solution instance id<sup>1</sup>|
  | kind | Resource kind, like `skills`, `solutions` and `devices`|
  | label-selector | Label selector (optional), for example: `foo=bar`|
  | ref | Reference provider type. Use `v1alpha2.ReferenceK8sCRD` to query K8s objects |
  | scope | Namespace, like `default`|
  | version | Resource version, like `v1`|

  **<sup>1</sup>**: This parameter is supposed to be used in `skill` queries only. When supplied, `skill` parameter values will be overridden by corresponding values (named as `<skill name>.<parameter name>`) in the `instance` object. In addition, if the `alias` parameter is specified, Symphony uses `<skill name>.<alias>.<parameter name>` to locate instance overrides instead. For more information, see [parameter management](../ai-management/parameter-management.md).

  The reference endpoint can also be used to resolve Azure Custom Vision model download URLs. Some addition parameters are used for such queries:

  | Parameter | Comment |
  |--------|--------|
  | flavor | Custom Vision model export flavor (such as `TensorFlowNormal`)|
  | lookup |set to `download` when querying download URLs|
  | iteration | Custom Model iteration |
  | platform| Custom Vision model export platform (such as `TensorFlow`)|

  > **NOTE** if `flavor` and `platform` are omitted, the reference endpoint only returns existing exports. If both are provided, the reference endpoint will request a new export if necessary (such as when existing exports have expired).

### Examples

* Get AI Skill with name `cv-model`:

  ```bash
  http://localhost:8088/v1alpha2/agent/references?scope=default&kind=skills&version=v1&group=ai.symphony&id=cv-model&&ref=v1alpha2.ReferenceK8sCRD
  ```

* List AI Skill with label `foo=bar`:

  ```bash
  http://localhost:8088/v1alpha2/agent/references?scope=default&kind=skills&version=v1&group=ai.symphony&label-selector=foo=bar&&ref=v1alpha2.ReferenceK8sCRD
  ```

* Get AI Skill with name `sv-skill`, overwrite its parameters with values in instance `dummy-instance`, aliased as `abc`. For more information, see [parameter management](../ai-management/parameter-management.md).

  ```bash
  http://localhost:8088/v1alpha2/agent/references?scope=default&kind=skills&version=v1&group=ai.symphony&id=cv-skill&ref=v1alpha2.ReferenceK8sCRD&instance=dummy-instance&alias=abc
  ```

## Report object state

You can report object state through Symphony agent.

* **Route**: `http://<Symphony agent endpoint>:8088/v1alpha2/agent/references`
* **Method**: POST
* **Parameters**:
  | Parameter | Comment |
  |--------|--------|
  | group | Resource group, like `ai.symphony`, `solution.symphony`, and `fabric.symphony` |
  | id | Resource name (optional)|
  | kind | Resource type, like `skills`, `solutions`, and `devices`|
  | overwrite | If set to true, the object state will be reset to reported properties. Otherwise, the reported properties are merged into existing state (optional, default = false) |
  | scope | Namespace, like `default` |
  | version | resource version, like `v1` |

* **Body**: A key-value pair collection of reported properties

  >**NOTE:** Symphony always reports object state as a key-value collection named `properties`.

## Capture camera frame

Capturing camera frames happens automatically, and captured image URL will be reported as part of `device` state as a `snapshot` property.
