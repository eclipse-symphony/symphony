# Device

In Symphony, a managed device, which is commonly a sensor or actuator like a camera or an ECU, is registered as a `device` object. A device object can be either manually maintained, or automatically populated by an auto-discovery mechanism like [Akri]( https://github.com/project-akri/akri). The state management of a device is proxied through a target object. In the case where a Symphony agent runs on a target, the agent periodically probes all attached devices and reports their states as part of the target's state update.

> **NOTE**: A target hosts Symphony payloads, while a device usually doesn't

## A typical device management flow

The following workflow depicts a typical process of camera management as an example of a device management workflow.

1. A device object is created (manually or through auto-discovery mechanisms). The following is a simplified device object that describes a RTSP camera on the local network. When a device is attached to a target, a label `<target-name>: "true"` is added to the device. The following example shows a `camera-1 device` that is attached to `gateway-1 target`.

   ```yaml
   apiVersion: fabric.symphony/v1
   kind: Device
   metadata:
     name: camera-1
     labels:
       gateway-1: "true"
   spec:
     properties:
       rtsp: "rstp://192.168.0.11"      
       user: "admin"
       password: "admin"
   ```

1. The target has a `topology` element defined that uses a label selector to select all devices marked to be connected to it.

   ```yaml
   apiVersion: fabric.symphony/v1
   kind: Target
   metadata:
     name: gateway-1
   spec:
     topologies:
     - selector: 
       label.gateway-1: "true"        
   ```

1. The agent running on the target periodically reaches out to all registered devices in its topology and reports device states back to the control plane. As part of the state report, the latest thumbnail URL is attached as a `snapshot` property in the device status, which can be used to display a preview image on an UX.

   ```yaml
   apiVersion: fabric.symphony/v1
   kind: Device
   metadata:
     ...
   spec:
     properties:
       ...
   status:
     properties:
       snapshot: https://voestore.blob.core.windows.net/snapshots/camera-1-snapshot.jpg # <------------
   ```

## Schema

`device.fabric.symphony` represents a non-computational device, such as a sensor (camera, microphone, vibration, pressure, etc.) or an actuator (motor, controller, I/O device, etc.).

| Field | Type | Description |
|-------|------|-------------|
| `Bindings`| `[]BindingSpec` | A list of bindings that represent actions allowed on the AI skill |
| `DisplayName` | `string` | A user friendly name |
| `Properties` | `map[string]string` | A property bag |
