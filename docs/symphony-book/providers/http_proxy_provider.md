# HTTP Proxy Provider

HTTP Proxy provider delegates provider operations to a different process/machine. This provider enables you to write your own provider implementation in any programming languages, and to host your [standalone provider](./standalone_providers.md) on any machines that are reachable by the Symphony control plane via HTTP.

For example, you can proxy provider operations to a Windows machine, and your provider on the Windows machine can use PowerShell to implement its logics.

## Provider Configuration

| Field | Comment |
|--------|--------|
| serverUrl | HTTP/HTTPS URL of the provider implemenation|


## Related Topics

* [Provider Interface](./provider_interface.md)
* [Writing a Python-based Provider](./python_provider.md)
* [Scenario:Deploying Linux container with a WUP frontend](../scenarios/linux-with-uwp-frontend.md)