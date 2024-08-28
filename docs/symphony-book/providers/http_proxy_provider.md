# HTTP proxy provider

The HTTP proxy provider delegates provider operations to a different process/machine. This provider enables you to write your own provider implementation in any programming language, and to host your [standalone provider](./standalone_providers.md) on any machines that are reachable by the Symphony control plane via HTTP.

For example, you can proxy provider operations to a Windows machine, and your provider on the Windows machine can use PowerShell to implement its logics.

## Provider configuration

| Field | Comment |
|--------|--------|
| `serverUrl` | HTTP/HTTPS URL of the provider implementation|

## Related topics

* [Provider interface](./provider_interface.md)
* [Write a Python-based Provider](./python_provider.md)
* [Scenario: Deploy a Linux container with a WUP frontend](../scenarios/linux-with-uwp-frontend.md)
