# Providers

Symphony uses providers to encapsulate platform-specific knowledge to a small scope. Providers are also the main extension points of Symphony. A provider implements a specific capability interface, such as state management, pub-sub, and state seeking. A provider is stateless and single-threaded. Some provider interfaces (such as state seeking) require a provider to be idempotent.

## Provider types

Common provider types include:

* Proxy, like [HTTP proxy](./http_proxy_provider.md) and [MQTT proxy](./mqtt_proxy_provider.md)
* [Reference](./reference_provider.md)
* [Target](./target_provider.md)
* [Staging](./staging_provider.md)
* Certificate
* Probe
* Pub-Sub
* Reporter
* State  
* Uploader
  
## Develop providers

* [Provider interface](./provider_interface.md)
* [Write a Python-based provider](./python_provider.md)