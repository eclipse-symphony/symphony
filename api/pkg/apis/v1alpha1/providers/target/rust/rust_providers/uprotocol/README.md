# Proxy Target Provider for Eclipse uProtocol

This crate provides a (proxy) target provider which allows (standalone) target providers to be implemented as Eclipse uProtocol entities (uEntity).
For this purpose, the `ITargetProvider` trait defined by [Symphony's Rust binding](../../symphony/README.md) has been _translated_ into a corresponding [AsyncAPI definition](./uservice/asyncapi.yaml) which needs to be implemented by the uEntity.

## Using the Proxy Target Provider

The provider works similarly to Symphony's standard MQTT proxy provider. It supports using both [Eclipse Zenoh&trade;](https://zenoh.io) as well as MQTT 5 as the uProtocol transport for communicating with the target uEntity. The provider supports the following configuration properties:

| Property  | Mandatory | Description |
|-----------|-----------|-------------|
| `name`    | yes | A logical name that should reflect the semantics of the proxied remote target. |
| `libFile` | yes | The absolute path to the shared library that contains the provider code. |
| `libHash` | yes | The hex representation of the hash code to use for verifying the integrity of the shared library.<br>This value will be compared to the SHA256 hash value computed for the given library file.<br>A value of `any` disables the check. This is useful during development, when updating the expected hash value for each build cycle seems undesirable/unnecessary. |
| `validationRule` | no | If not specified the default validation rule will be used. |
| `localEntity` | yes | A uProtocol URI that defines the address of the proxy provider within the uProtocol network. |
| `getMethodUri` | yes | The uProtocol URI of the (proxied) target provider's _Get_ method. The URIs for the _Update_ and _Delete_ methods are derived from this URI. |
| `getMethodTimeoutMillis` | no | The number of milliseconds that the Symphony control plane should wait for a _Get_ method response from the target provider before considering the method invocation to have failed. |
| `applyMethodTimeoutMillis` | no | The number of milliseconds that the Symphony control plane should wait for an _Update_ or _Delete_ method response from the target provider before considering the method invocation to have failed. |
| `zenohConfig` | no | The absolute path to a Zenoh configuration file.<br> Either this property or `brokerAddress` must be set to configure a usable transport. |
| `brokerAddress` | no | The address of the MQTT 5 broker to use for communicating with the proxied target.<br>Either this property or `zenohConfig` must be set to configure a usable transport.|
| `clientId` | no | The client ID to use when connecting to the MQTT 5 broker. |

## Example Configuration

The following JSON object is an example Symphony _Target_ definition. It can be used to install _Engine Controller_ and _Telematics Unit_ firmware to a remote vehicle's ECUs by means of the uProtocol target provider. The target provider is configured with the MQTT 5 broker address that is used for the uProtocol based communication with the remote target uEntity identified by the `getMethodUri` address.

```json
{
    "spec": {
        "displayName": "Vehicle_ECUs",
        "components": [
            {
                "name": "Engine Controller",
                "metadata": {
                    "manufacturer": "ACME Inc.",
                    "model": "EC-14"
                },
                "type": "ecu.fw",
                "properties": {
                    "fw-image": "https://acme.io/fw/engine-control-1.45.img"
                }
            },
            {
                "name": "Telematics Unit",
                "type": "ecu.fw",
                "properties": {
                    "fw-image": "https://non-existent.io/fw/telematics-unit-2.0.img"
                },
                "parameters": {
                    "back-end-url": "https://non-existent.io/telematics"
                }
            }
        ],
        "topologies": [
            {
                "bindings": [
                    {
                        "role": "ecu.fw",
                        "provider": "providers.target.rust",
                        "config": {
                            "name": "ECU Firmware Updater",
                            "libFile": "/absolute/path/to/libuprotocol.so",
                            "libHash": "any",
                            "localEntity": "up://symphony/DA00/1/0",
                            "getMethodUri": "//ecu-updater.app/A100/1/1",
                            "brokerAddress": "mqtt://localhost:1883"
                        }
                    }
                ]
            }
        ]
    }
}
```

The path to the library needs to be adapted to the concrete location of the target folder configured for your Rust toolchain.
