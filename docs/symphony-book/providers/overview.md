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
  
## Developing Providers

* [Provider Interface](./provider_interface.md)
* [Writing a Python-based Provider](./python_provider.md)