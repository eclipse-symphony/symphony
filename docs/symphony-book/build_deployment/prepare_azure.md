# Prepare Azure resources

Symphony doesn't have dependencies on Azure resources. However, to run through sample scenarios and demos, you need to create a few Azure resources to support Azure IoT Hub scenarios.

> **NOTE:** All instructions below are provided as Azure CLI commands. Use Azure Cloud Shell if you don't have the CLI installed locally. We assume you've already logged in (az login).

## 1. Create a resource group

We require some standard tier and premium tier offers, so the testing won’t be free. We recommend keeping all of your demo resources under the same resource group so that you can easily clean up your environment once you are done.

```bash
az group create -l <location> -n <resource group name>
#NOTE: Choose a location where all services are available, such as westus2.
```

## 2. Create an IoT Hub

```bash
az iot hub create --resource-group <resource group name> --name <IoT Hub name> --sku S1 --partition-count 2
```

> **NOTE:** The test scenarios require a standard tier (S1, S2 or S3) version of IoT Hub. You can choose a different --partition-count as desired (default is 4). For lite tests a single partition may be sufficient.

## 3. Create an IoT Edge device

Symphony can acts both as a control plane and a deployment target (we call target). When it's used as a target, it corresponds to an IoT Edge device, when IoT Edge modules are used.

```bash
az iot hub device-identity create --hub-name <Iot Hub name> --device-id <device id> --edge-enabled
```

> **NOTE:** In most sample scenarios, we'll use Symphony as both the control plane and the target.

## Next steps

Now that you have your Azure resources in place, create a Kubernetes cluster with one of the following options:

* [AKS](./aks.md)
* [Kind](./kind.md)
* [MicroK8s](./microk8s.md)
* [K3s](./k3s.md)
