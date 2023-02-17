# Providers

Symphony uses providers to encapsulate platform-specific knowledge to a smallest scope. Providers are also the main extension points of Symphony. A provider implements a specific capability interface, such as state management, pub-sub, and state seeking. A provider is stateless and single-threaded. Some provider interfaces (such as state seeking) require a provider to be idempotent.

## Provider Types
* Certificate
* Probe
* [Proxy](./proxy_provider.md)
* Pub-Sub
* [Reference](./reference_provider.md)
* Reporter
* State  
* [Target](./target_provider.md)
* Uploader
  Manages components on a [```Target```](../uom/target.md). A target provider reads the **current state** from a target and reconcile it with the **desired state** specified by a user.
  |Provider Name| Description|
  |--------|--------|
  | ```providers.target.azure.adu```| Coordinates [Azure Device Update](https://learn.microsoft.com/en-us/azure/iot-hub-device-update/understand-device-update) updates|
  | ```providers.target.azure.iotedge```|Coordinates [Azure IoT Edge](https://azure.microsoft.com/en-us/products/iot-edge/) updates (read [more](./iot_provider.md))|
  | ```providers.target.helm```|Coordinates [Helm](https://helm.sh/) chart updates (read [more](./helm_provider.md))|
  | ```providers.target.http```| Invoke a HTTP endpoint (read [more](./http_provider.md))|
  | ```providers.target.k8s```|Native Kubernetes provider (read [more](./k8s_provider.md))|
  | ```providers.target.kubectl```|Coordinates Kubernetes Yaml file updates|
  | ```providers.target.mqtt```| Invoke a standalone provider over MQTT (read [more](./mqtt_proxy_provider.md))|
  | ```providers.target.proxy```| Proxies provider operations to a different machine (read [more](./http_proxy_provider.md))|
  | ```providers.target.script``` | Uses Bash/PowerShell script to deploy updates (read [more](./script_provider.md))|
  |```providers.target.staging```| Stages artifacts on the target object for a polling agent to pick up (read [more](./staging_provider.md)) |
  | ```providers.target.win10.sideload```| [Sideload apps](https://learn.microsoft.com/en-us/windows/application-management/sideload-apps-in-windows-10) with Windows client devices and XBOX (read [more](./win10_sideload_provider.md))|
  

## Developing Providers

* [Provider Interface](./provider_interface.md)
* [Writing a Python-based Provider](./python_provider.md)