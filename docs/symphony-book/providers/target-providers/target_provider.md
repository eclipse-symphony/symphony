# Target providers

A target provider interacts with a specific system to carry out state seeking actions on the system, such as getting the current state and applying the new desired state. Target Provider is the most important extension point of Symphony. Symphony itself is platform-agnostic, system-specific knowledge is encapsulated by the corresponding provider. Symphony currently has the following target providers:

| Provider type | Platform |
|--------|--------|
| `providers.target.adb` | Sideload Android apps using [Android Debug Bridge](https://developer.android.com/tools/adb) |
|`providers.target.arcextension` | Manage Azure Arc extensions |
| `providers.target.azure.adu` | Update devices using [Device Update for IoT Hub](https://learn.microsoft.com/azure/iot-hub-device-update/) |
| `providers.target.azure.iotedge` | Deploy solution instances as [Azure IoT Edge](https://learn.microsoft.com/azure/iot-edge/?view=iotedge-1.4) modules<br><br>[`IoT Edge provider`](./iot_provider.md) |
| `providers.target.configmap`| Manage kubernetes configMap object |
| `providers.target.docker`| Deploy [Docker](https://www.docker.com/) containers |
| `providers.target.helm`| Deploy [Helm](https://helm.sh/) charts<br><br>[Helm provider](./helm_provider.md) |
| `providers.target.http`| Send state-seeking actions (such as `Apply()`) to an HTTP endpoint<br><br>[HTTP provider](./http_provider.md) |
| `providers.target.ingress`| Manage kubernetes ingress object |
| `providers.target.k8s` | Deploy solution instances as K8s [deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) |
| `providers.target.kubectl`| Deploy K8s YAML docs using `kubectl` |
| `providers.target.mock`| A mock provider to be used in manager unit tests |
| `providers.target.mqtt`| Delegate state-seeking actions to a remote management plane over MQTT |
| `providers.target.proxy`<sup>1</sup>| Delegate state-seeking actions to a remote management plane over HTTP or MQTT<br><br>[HTTP proxy provider](../http_proxy_provider.md)<br>[MQTT proxy provider](../mqtt_proxy_provider.md) |
| `providers.target.script`| Delegate state-seeking actions to external Bash/Powershell scripts<br><br>[Script provider](./script_provider.md) |
| `providers.target.staging`| Stage solution component on the target objects<sup>2</sup>|
| `providers.target.win10`| Sideload Windows apps using [WinAppDeployCmd](https://learn.microsoft.com/windows/uwp/packaging/install-universal-windows-apps-with-the-winappdeploycmd-tool). |

1: The `providers.target.proxy` provider expects the target HTTP or MQTT handler to implement the [target provider interface](./provider_interface.md), unlike the HTTP or MQTT providers that allow any handler to be used. The HTTP provider is commonly used as a webhook to trigger external workflows <!--(such as [human approval](../scenarios/human-approval.md))--> instead of doing actual deployment.

2: The staging provider stages solution components onto the selected target object instead of doing any actual deployments. The scenario is to allow an external poll-based agent to pick up components to be deployed from the target object.

## Related topics

* [Target provider conformance](conformance.md)
* [Target provider interface](./provider_interface.md)
