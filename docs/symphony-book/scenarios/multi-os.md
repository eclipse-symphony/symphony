# Multi-OS deployment

_(last update: 2/24/2024)_

This scenario deploys an application that spans Kubernetes, Windows, bare-metal Linux and RTOS. The application simulates a smart appliance that has a backend service, a Windows-based frontend, a driver on an ECU as well as a library on an RTOS device. This scenario can also be considered as a simplified SDV system, where core services run on a HPC cluster, an infotainment system runs on a consumer-facing OS like Android or Windows, and some device drivers run on ECUs.
![multi-os](../images/multi-os.png)

## Generic flow

1. Define your application as a Symphony `solution` object. This object contains four components: a Docker container that is to be deployed on Kubernetes, a UWP application package that is to be deployed to Windows, a Web Assembly module that is to be deployed on a Flatcar VM, and a ThreadX binary image that is to be deployed on a MIMXRT1170-EVK board.
2. Define four deployment `target`s: Kubernetes cluster, Windows machine, Flatcar VM and RTOS.    
    In this scenario, the Windows machine is managed by a [MQTT proxy provider](../providers/mqtt_proxy_provider.md) deployed on the Windows machine.
3. Launch a local Symphony with the proper configuration on the Windows machine:
    ```bash
    ./symphony-api -c ./symphony-agent.json -l Debug
    ```
    > **NOTE**: See [Setup local Symphony on Windows](#setup-local-symphony-on-windows) for more details.

3. Define an `instance` object that maps the above three components to corresponding targets.

## Sample artifacts
You can find sample artifacts in this repository under the `docs/samples/multi-os` folder:
| Artifact | Purpose |
|--------|--------|
| [instance.yaml](../../samples/multi-os/instance.yaml) | Instance definition |
| [mosquitto.yaml](../../samples/multi-os/mosquitto.yaml) | Mosquitto MQTT broker definition<sup>1</sup> |
| [scripts/apply.ps1](../../samples/multi-os/scripts/apply.ps1) | PowerShell script to drive Windows command script to flash NXP board |
| [scripts/bubble_peripheral.cmd](../../samples/multi-os/scripts/bubble_peripheral.cmd) | Script to flash a bubble scale app, see instructions [here](../../samples/scenarios/nxp1170/nxp1170.md) |
| [scripts/get.ps1](../../samples/multi-os/scripts/get.ps1) | PowerShell script to get NXP board state<sup>2</sup> |
| [scripts/hello_world.cmd](../../samples/multi-os/scripts/hello_world.cmd) | Script to flash the 'hello, world' app, see instructions [here](../../samples/scenarios/nxp1170/nxp1170.md) |
| [scripts/remove.ps1](../../samples/multi-os/scripts/remove.ps1) | PowerShell script to revert NXP board to the 'hello, world' app.|
| [solution.yaml](../../samples/multi-os/solution.yaml) | Solution definition |
| [symphony-agent.json](../../samples/multi-os/symphony-agent.json) | Sample Symphony config file to be used on Windows
| [target-ecu.yaml](../../samples/multi-os/target-ecu.yaml) | ECU Target definition |
| [target-k8s.yaml](../../samples/multi-os/target-k8s.yaml) | K8s Target definition |
| [target-pc.yaml](../../samples/multi-os/target-pc.yaml) | PC Target definition |
| [target-rtos.yaml](../../samples/multi-os/target-rtos.yaml) | RTOS Target definition |

<sup>1</sup>: You can use the `mosquitto.yaml` to provision a Mosquitto MQTT broker on your K8s cluster. You can choose to use another existing MQTT broker as needed. You can also embed the `mosquitto.yaml` contents into your `target-k8s.yaml` as a Target component if you want this deployment is automated.

<sup>2</sup>: This demo scenario uses a text file on the Windows machine, named after the component, to track if an app has been flashed to the NXP board.

## Build sample packages

To complete the demo setup, you'll also need to prepare a few packages:

### Build Flatcar image with updated Piccolo
Read instructions [here](../agent/piccolo-wasm-e2e.md) for details on how to build the Flatcar image and the Web Assembly module.

### Build NXP images
Read instructions [here](../../samples/scenarios/nxp1170/nxp1170.md) for more details on building the images and generating the installation script. In this demo, we'll flash a "Hello, World" app as the removal script.
### Build UWP sample app
1. Open `docs/samples/scenarios/homehub/uwp-app/HomeHub.sln` with Visual Studio 2022.
2. Build the solution for `x64` architecture, `Debug` profile.
3. Copy the `HomeHub.Package/AppPackages/HomeHub.Package_1.0.9.0_Debug_Test` folder to a folder that is accessible by the proxy provider (see below).

### Setup local Symphony on Windows
You can use the [`symphony-agent.json`](../../samples/multi-os/symphony-agent.json) as a template for your local Symphony configuration. A few things to notice:

1. You can define multiple MQTT bindings, each corresponds to a Target object on the control plane side.
2. You can increase default MQTT timeout by modifying the `timeoutSeconds` value. Because provider deployments are blocking, you need to make sure the deployment can finish in time to respond before this time window expires.
3. On the `solution-manager` configuration, you should define local target providers you want to use. **Note in current version, the provider name needs to match with the target name**.

### Launch the Flatcar VM
See instructions [here](../agent/piccolo-wasm-e2e.md) for more details. This demo scenario deploys a web site hosted in a Web Assembly module to the Flatcar. In order to access the website from a browser on the hosting PC, you need to mark sure port 8050 is exposed:
```powershell
.\qemu-system-x86_64.exe -m 2G -netdev user,id=net0,hostfwd=tcp::8085-:8085 -device virtio-net-pci,netdev=net0 -fw_cfg name=opt/org.flatcar-linux/config,file=c:\demo\ignition.json -drive if=virtio,file=c:\demo\flatcar_production_qemu_image.img
```

## Deployment steps

1. Deploy targets and solution:

   ```bash   
   kubectl apply -f solution.yaml
   kubectl apply -f target-ecu.yaml
   kubectl apply -f target-k8s.yaml
   kubectl apply -f target-pc.yaml
   kubectl apply -f target-rtos.yaml   
   ```

2. Examine the current state of all targets:

    * **PC:** search for "HomeHub" app on Windows search - no results should be returned.
    * **ROTS:** attach a Putty terminal to the board and observe the default 'Hello, World' application is deployed.
    * **K8s:** use `kubectl get pods` to see no application pods are deployed
    * **Flatcar:** use a browser to access `http://localhost:8050` and the website shouldn't be accessible.

3. Trigger deployment:

    ```bash
    kubectl apply -f instance.yaml
    ```
4. Check deployment status:

    * **PC:** search for "HomeHub" app on Windows search - the app should be found. Launch the app.
    * **ROTS:** observe the data stream on the Putty terminal - move the board to see accelerators work.
    * **K8s:** use `kubectl get pods` to see application pod.
    * **Flatcar:** use a browser to access `http://localhost:8050` and the website should show up.

5. Remove the deployment:

    ```bash
    kubectl delete instance multi-os-instance
    ```
6. Examine the final state:

    * **PC:** search for "HomeHub" app on Windows search - no results should be returned.
    * **ROTS:** attach a Putty terminal to the board and observe the default 'Hello, World' application is deployed.
    * **K8s:** use `kubectl get pods` to see no application pods are deployed
    * **Flatcar:** use a browser to access `http://localhost:8050` and the website shouldn't be accessible.

> **NOTE**: In version 0.48.2 or lower, the Web Assembly is not removed, so the website will still be accessible. 

