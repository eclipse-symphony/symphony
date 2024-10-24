# Write a Python-based provider

Under the `samples/scenarios/homehub/python-provider` folder, you can find a sample implementation of a Python-based provider. The provider implements an HTTP server that implements the [provider interface](./provider_interface.md), which can be invoked by a proxy provider running on the same machine or from another machine.

> **NOTE**: This sample shows a standalone provider over HTTP. You can also write a standalone provider over MQTT.

At the time of writing, we don’t have a Python Symphony SDK yet. Instead, there’s an `sdk_poc.py` file under the same folder that serves as a temporary helper before the SDK is offered. The file defines the data structures needed to implement the provider interface.

An out-of-process provider is expected to handle the following routes:

| Route | Method | Comment |
|--------|--------|--------|
| `/instances` | GET | Get currently installed components |
| `/instances` | POST | Apply components |
| `/instances` | DELETE | Remove components |
| `/needsupdate` | GET | If an update is needed (returns `200`) or not (returns `500`) |
| `/needsremove`| GET | if a deletion is needed (returns `200`) or not (returns `500`)|

The `NeedsUpdate()` method and the `NeedsRemove()` method take two component arrays as parameters. In the REST interface, the two arrays are combined into a single JSON object:

```json
{
    "desired": [],
    "current": []
}
```
