# Symphony Quick Start - Deploying a simulated temperature sensor Solution to an Azure IoT Edge device
This quick start walks you through the steps of deploying a new Symphony solution instance to an Azure IoT Edge device.

> **NOTE**: The following steps are tested under a Ubuntu 20.04.4 TLS WSL system on Windows 11. However, they should work for Linux, Windows, and MacOS systems as well.

![IoT Edge](../images/quick-start-iot-edge.png)

## Configure an IoT Edge device

### 1. Register an IoT Edge device

These steps create a new resource group, a new IoT Hub, and a new IoT Edge device.

```bash
# install Azure IoT extension if needed
az extension add --name azure-iot

# create resource group
az group create --name <REPLACE_WITH_GROUP_NAME> --location westus2

# create IoT Hub
az iot hub create --name <REPLACE_WITH_HUB_NAME> --resource-group <REPLACE_WITH_GROUP_NAME> --sku S1

# create a IoT Edge device
az iot hub device-identity create --device-id s8c-vm --hub-name <REPLACE_WITH_HUB_NAME> --edge-enabled

# get IoT Edge device connection string
# copy the output to use when you configure the device
az iot hub device-identity connection-string show --device-id s8c-vm --resource-group <REPLACE_WITH_GROUP_NAME> --hub-name <REPLACE_WITH_HUB_NAME>
```

### 2. Configure a Linux VM as your IoT Edge device

You need to prepare a Linux virtual machine or physical device for IoT Edge. In this guide, you create a Linux VM. You can use an ARM deployment to automate these steps, or use the Azure CLI.

* ARM deployment steps:

  ```bash
  # use ARM deployment to create a VM and install IoT Edge runtime
  az deployment group create \
  --resource-group <REPLACE_WITH_GROUP_NAME> \
  --template-uri "https://raw.githubusercontent.com/Azure/iotedge-vm-deploy/1.4/edgeDeploy.json" \
  --parameters dnsLabelPrefix='s8c-vm' \
  --parameters adminUsername=<USER_NAME> \
  --parameters deviceConnectionString=$(az iot hub device-identity connection-string show --device-id s8c-vm --hub-name <REPLACE_WITH_HUB_NAME> -o tsv) \
  --parameters authenticationType='password' \
  --parameters adminPasswordOrKey="<REPLACE_WITH_PASSWORD>"
  ```

* Azure CLI steps:

  ```bash
  # create vm
  az vm create --resource-group <REPLACE_WITH_GROUP_NAME> --name s8c-vm --image UbuntuLTS --admin-username <USER_NAME> --generate-ssh-keys --size Standard_D2s_v5

  # SSH into the machine
  ssh <USER_NAME>@<public IP of your VM>
  
  # if the VM requires a private key:
  ssh -i ~/.ssh/id_rsa.pem <YOUR_USER_NAME>@<public IP of your VM>

  # update repo and signing key
  wget https://packages.microsoft.com/config/ubuntu/18.04/multiarch/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
  sudo dpkg -i packages-microsoft-prod.deb
  rm packages-microsoft-prod.deb

  # install container engine
  sudo apt-get update; \
    sudo apt-get install moby-engine

  # install IoT Edge engine runtime
  sudo apt-get update; \
    sudo apt-get install aziot-edge defender-iot-micro-agent-edge

  # update iotedge setting
  sudo iotedge config mp --connection-string '<REPLACE_WITH_DEVICE_CONNECTION_STRING>'
  ```

**Note**
If you encounter below package dependency error during installing `aziot-edge`:
```
The following packages have unmet dependencies:
 aziot-identity-service : Depends: libssl1.1 (>= 1.1.1) but it is not installable
```

This is because Ubuntu 22.04 uses `libssl3` and `libssl1.1` is deprecated on it. To address the issue, check the latest versions for Edge and IoT Identity Service and install them from the release page.

```shell
# check the latest aziot-edge version
wget -qO- https://raw.githubusercontent.com/Azure/azure-iotedge/main/product-versions.json | jq -r '
  .channels[]
  | select(.name == "stable").products[]
  | select(.id == "aziot-edge").components[]
  | select(.name == "aziot-edge").version
'
# result: <AZIOT_EDGE_VERSION>

# check the latest aziot-identity-service version
wget -qO- https://raw.githubusercontent.com/Azure/azure-iotedge/main/product-versions.json | jq -r '
  .channels[]
  | select(.name == "stable").products[]
  | select(.id == "aziot-edge").components[]
  | select(.name == "aziot-identity-service").version
'
# result: <AZIOT_IDENTITY_SERVICE_VERSION>

# download and install
wget https://github.com/Azure/azure-iotedge/releases/download/<AZIOT_EDGE_VERSION>/aziot-edge_<AZIOT_EDGE_VERSION>-1_ubuntu22.04_amd64.deb -O aziot-edge.deb
sudo dpkg -i aziot-edge.deb

wget https://github.com/Azure/azure-iotedge/releases/download/<AZIOT_EDGE_VERSION>/aziot-identity-service_<AZIOT_IDENTITY_SERVICE_VERSION>-1_ubuntu22.04_amd64.deb -O aziot-identity-service.deb
sudo dpkg -i aziot-identity-service.deb
```

### 3. Deploy modules on your IoT Edge device

To deploy your IoT Edge modules, go to your IoT hub in the Azure portal, then:

1. Select Devices from the IoT Hub menu.

2. Select your device to open its page.

3. Select the Set Modules tab.

4. Select Review + create at the bottom.

5. Select Create to deploy the modules.

When the module setup completes, you should see two modules in the list: `$edgeAgent` and `$edgeHub`.
For more information, please refer to [Create and provision an IoT Edge device on Linux using symmetric keys](https://learn.microsoft.com/en-us/azure/iot-edge/how-to-provision-single-device-linux-symmetric?view=iotedge-1.5&tabs=azure-cli%2Cubuntu).

## OPTION 1: Use Maestro

To use this option, first install Maestro and the Symphony API. For more information, see [Use Symphony with the Maestro CLI tool](./quick_start_maestro.md).

Once you have maestro installed, you can launch this sample with the following command:

```bash
maestro samples run hello-iot-edge --set iot-hub-key=<REPLACE_WITH_HUB_KEY> --set iot-hub-name=<REPLACE_WITH_HOST_NAME> --set device-name=s8c-vm
```

You can get your IoT host name, hub key and key name from the connection string:

```bash
az iot hub connection-string show --hub-name <REPLACE_WITH_HUB_NAME>
# sample output:
# {
#  "connectionString": "HostName=<HOST_NAME>;SharedAccessKeyName=<HUB_KEY_NAME>;SharedAccessKey=<HUB_KEY>"
# }
```

Make sure to set the `iot-hub-name` to the complete host name. Otherwise, the iot-edge API may not acuqire the module information correctly.

If the sample runs successfully, you should see two instances running:
```bash
$ kubectl get instances
NAME                         STATUS      TARGETS   DEPLOYED
sample-iot-edge-instance-1   Succeeded   1         1
sample-iot-edge-instance-2   Succeeded   1         1
```

To clean up, use:

```bash
maestro samples remove hello-iot-edge 
```

## OPTION 2: Use Helm and Kubectl

### 0. Prerequisites

* [kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/) is configured with the Kubernetes cluster you want to use as the default context
* [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/)

### 1. Register the IoT Edge device as a Symphony Target

A Symphony *target* is a specific deployment environment that inculdes definitions of infrastructural components and types of resources to be managed.

Create a YAML file called `target.yaml` that registers the IoT Edge device as a Symphony target:

```yaml
apiVersion: fabric.symphony/v1
kind: Target
metadata:
  name: sample-iot-edge-target
spec:  
  forceRedeploy: true
  topologies:
  - bindings:
    - role: instance
      provider: providers.target.azure.iotedge
      config:
        name: "iot-edge"
        keyName: "<REPLACE_WITH_HUB_KEY_NAME>"
        key: "<REPLACE_WITH_HUB_KEY>"
        iotHub: "<REPLACE_WITH_HUB_NAME>"
        apiVersion: "2020-05-31-preview"
        deviceName: "s8c-vm"
```

You can get your IoT hub's host name, hub key and key name from the connection string:

```bash
az iot hub connection-string show --hub-name <REPLACE_WITH_HUB_NAME>
```

This YAML file is also available at [docs/samples/iot-edge/simulated-temperature-sensor/target.yaml](../../samples/iot-edge/simulated-temperature-sensor/target.yaml).

### 2. Create the Symphony Solution

A Symphony *solution* is a template that defies an application workload to be deployed on one or more *targets*.

Create a YAML file called `solution.yaml` that describes a Symphony Solution with a single component, which is based on the `mcr.microsoft.com/azureiotedge-simulated-temperature-sensor:1.0` container.

```yaml
apiVersion: solution.symphony/v1
kind: Solution
metadata:
  name: sample-simulated-temperature-sensor
spec:
  components:
  - name: "simulated-temperature-sensor"
    properties:
      container.version: "1.0"
      container.type: "docker"
      container.image: "mcr.microsoft.com/azureiotedge-simulated-temperature-sensor:1.0"
      container.createOptions: ""
      container.restartPolicy: "always"      
    routes:
    - route: SimulatedTemperatureSensorToIoTHub
      type: iothub
      properties:
        definition: "FROM /messages/modules/simulated-temperature-sensor/* INTO $upstream"
```

This YAML file is also available at [docs/samples/iot-edge/simulated-temperature-sensor/solution.yaml](../../samples/iot-edge/simulated-temperature-sensor/solution.yaml).

### 3. Create the Symphony Solution Instance

A Symphony *solution instance* maps a *solution* to one or multiple *targets*.

Create a YAML file called `instance-1.yaml` that maps the `sample-simulated-temperature-sensor` soltuion to the `sample-iot-edge-target` target above:

```yaml
apiVersion: solution.symphony/v1
kind: Instance
metadata:
  name: sample-iot-edge-instance-1
spec:
  solution: sample-simulated-temperature-sensor            
  target:
    name: sample-iot-edge-target
```

This YAML file is also available at [docs/samples/iot-edge/simulated-temperature-sensor/instance-1.yaml](../../samples/iot-edge/simulated-temperature-sensor/instance-1.yaml).

You can also create additional instances. Symphony does the following for each of the instances:

* It generates IoT Edge module with a `<instance id>-` prefix.
* It rewrites all IoT Edge route definitions in the solution so that messages are routed to the right instance module.

### 4. Create all objects

```bash
kubectl create -f target.yaml
kubectl create -f solution.yaml
kubectl create -f instance-1.yaml
kubectl create -f instance-2.yaml # if you've defined additional instances
```

### 5. Verify

Examine all Symphony objects have created:

```bash
kubectl get targets
kubectl get solutions
kubectl get instances
```

On IoT Hub page, verify all IoT Edge modules are up and running (screen shot shows only one instance):

![IoT Edge](../images/iot-edge.png)

### 6. Clean up Symphony objects

To delete all Symphony objects:

```bash
kubectl delete instance my-instance-1
kubectl delete solution simulated-temperature-sensor
kubectl delete target voe-target
```
