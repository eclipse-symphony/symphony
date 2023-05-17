# Target Providers

A target provider interacts with a specific system to carry out state seeking actions on the system, such as getting the current state and applying the new desired state. Target Provider is the most important extension point of Symphony. Symphony itself is platform-agnostic, system-specific knowledge is encapsulated by the corresponding provider. Symphony currently has the following target providers:

| Provider Type | Platform |
|--------|--------|
| ```providers.target.adb``` | Sideload Android apps using [Android Debug Bridge (adb)](https://developer.android.com/tools/adb) | 
| ```providers.target.azure.adu``` | Device updates using [Device Update for IoT Hub (a.k.a. ADU)](https://learn.microsoft.com/en-us/azure/iot-hub-device-update/) |
| ```providers.target.azure.iotedge``` | Deploy solution instances as [Azure IoT Edge](https://learn.microsoft.com/en-us/azure/iot-edge/?view=iotedge-1.4) modules |
| ```providers.target.docker```| Deploy [Docker](https://www.docker.com/) containers |
| ```providers.target.heml```| Deploy [Helm](https://helm.sh/) charts |
| ```providers.target.http```| Send state-seeking actions (such as ```Apply()```) to a HTTP endpoint |
| ```providers.target.k8s``` | Deploy solution instances as K8s [deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) |
| ```providers.target.kubectl```| Deploy K8s ```YAML``` docs using ```kubectl``` |
| ```providers.target.mqtt```| Delegate state-seeking actions to a remote management plane over [MQTT](https://mqtt.org/) |
| ```providers.target.proxy```<sup>1</sup>| Delegate state-seeking actions to a remote management plane over HTTP |
| ```providers.target.script```| Delegate state-seeking actions to external Bash/Powershell scripts |
| ```providers.target.staging```| Stage solution component on the target objects<sup>2</sup>|
| ```providers.target.win10```| Sideload Windows apps using [WinAppDeployCmd](https://learn.microsoft.com/en-us/windows/uwp/packaging/install-universal-windows-apps-with-the-winappdeploycmd-tool). | 

1: Difference between ```providers.target.http``` and ```providers.target.proxy```. The proxy provider expects the target HTTP handler implements the [Target Provider Interface](./provider_interface.md), while the http provider allows any handler to be used. The http provider is commonly used as a webhook to trigger external workflows (such as [human approval](../scenarios/human-approval.md)) instead of doing actual deployment.

2: The staging provider stages solution components onto the selected target object instead of doing any actual deployments. The scenario is to allow an exteranl poll-based agent to pick up components to be deployed from the target object.

