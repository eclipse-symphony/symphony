# Use Symphony with the Maestro CLI tool

_(last edit: 9/18/2023)_

The easiest way to get started with Symphony is to use Maestro, Symphony's CLI tool. You can install Maestro on your Windows, Mac, or Linux machine using a single command. And then, you can use `maestro up` to set up your Symphony environment and use `maestro samples` to run through predefined sample scenarios.

## 1. Install Maestro

* Install on Linux/Mac

  ```bash
  wget -q https://raw.githubusercontent.com/Haishi2016/Vault818/master/cli/install/install.sh -O - | /bin/bash
  ```
  
  > **Pre-release NOTE**: The ```Haishi2016``` repo is a temporary parking repo, which will be replaced before release.

* Install on Windows

  ```bash
  powershell -Command "iwr -useb https://raw.githubusercontent.com/Haishi2016/Vault818/master/cli/install/install.ps1 | iex"
  ```

  > **Pre-release NOTE**: The ```Haishi2016``` repo is a temporary parking repo, which will be replaced before release.

## 2. Set up Symphony API

Use `maestro up` to configure all dependencies and set up Symphony. If you already have **kubectl** configured, maestro will install Symphony API to your current Kubernetes context.

```bash
maestro up
```

> **NOTE**: ```maestro up``` will try to install a [kind](https://kind.sigs.k8s.io/) Kubernetes cluster on your machine, if you don't already have **kubectl** installed and configured.

## 3. Browse and run samples

* To browse samples:

  ```bash
  maestro samples list
  ```

* To deploy a sample:

  ```bash
  maestro samples run <sample name>
  ```

* To remove a sample:

  ```bash
  maestro samples remove <sample name>
  ```

## More topics

Now that you have the Symphony API, try out one of the quick start scenarios:

| Scenario | Requires K8s | Requires Azure | Requires Azure IoT Edge| Requires Docker | Requires RTSP Camera |
|--------|--------|--------|--------|--------|--------|
| [Deploy a Prometheus server to a K8s cluster](deploy_prometheus_k8s.md) | **Yes** | - | - | - | - |
| [Deploy a Redis container with standalone Symphony](deploy_redis_no_k8s.md)| - | - | - | **Yes** | - |
| [Deploy a simulated temperature sensor Solution to an Azure IoT Edge device](deploy_solution_to_azure_iot_edge.md) | **Yes** | **Yes** | **Yes** | - | - |
| [Manage RTSP cameras connected to a gateway](symphony-book/get-started/manage_rtsp_cameras.md) | **Yes** | **Yes** | - | - | **Yes** |
