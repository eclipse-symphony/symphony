# MQTT proxy provider

The MQTT proxy provider delegates provider operations to a different process/machine through an MQTT broker. This provider enables you to write your own provider implementation in any programming language, and to host your [standalone provider](./standalone_providers.md) on any machines that are reachable by the Symphony control plane via MQTT.

For example, you can proxy provider operations to a Windows machine, and your provider on the Windows machine can use PowerShell to implement its logic.

## Provider configuration

| Field | Comment |
|--------|--------|
| `brokerAddress` | broker address, like tcp://localhost:1883 |
| `clientID` | client ID for your choice |
| `keepAliveSeconds` | MQTT client keep-alive seconds |
| `pingTimeoutSeconds` | MQTT client ping timeout |
| `requestTopic` | topic for sending API requests |
| `responseTopic` | topic for getting API responses |
| `timeoutSeconds` | time limit on when a response is received<sup>1</sup> |

1: Messaging through pub/sub is an asynchronous communication pattern. However, Symphony requires all providers to operate in a synchronous manor. Once the request is sent, the MQTT proxy provider blocks to wait for a response, or until the timeout limit is reached, in which case the provider operation is considered failed.

## Related topics

* [Provider interface](./provider_interface.md)
* [Write a Python-based provider](./python_provider.md)
* [Scenario: Deploy a Linux container with a WUP frontend](../scenarios/linux-with-uwp-frontend.md)
