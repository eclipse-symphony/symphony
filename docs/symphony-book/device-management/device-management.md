# Device Management
In Symphony, a managed device, which is commonly a sensor or actuator like a camera or an ECU, is registered as a ```Device``` object. A ```Device``` object can be either manually maintained, or automatically populated by an auto-discovery mechanism like [Akri]( https://github.com/project-akri/akri). The state management of a ```Device``` is proxied through a ```Target``` object. In the case where a Symphony Agent runs on a ```Target```, the agent periodically probe all attached ```Devices``` and report their states as part of ```Target``` state update.

> **NOTE**: A **Target** hosts Symphony pyaloads, while a **device** usually doesn't

## A Typical Device Management Flow
The following workflow depicts a typical process of camera management as an example of a device management workflow.

1. A ```Device``` object is created (manually or through auto-discovery mechanisms). The following is a simplified ```Device``` object that describes a RTSP camera on the local network. When a ```Device``` is attached to a ```Target```, a label ```<target-name>: "true"``` is added to the device. The following example shows a ```camera-1 Device``` that is attached to ```gateway-1 Target```. 
    ```bash
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
2. The ```Target``` has a ```topology``` element defined that uses a label selector to select all ```Devices``` marked to be connected to it.
    ```bash
    apiVersion: fabric.symphony/v1
    kind: Target
    metadata:
      name: gateway-1
    spec:
      topologies:
      - selector: 
          label.gateway-1: "true"        
    ```
3. The agent running on the ```Target``` periodically reaches out to all registered devices in its topology and reports device states back to the control plane. As part of the state report, the latest thumbnail URL is attached as a ```snapshot``` property in ```Device``` status, which can be used to display a preview image on an UX.
    ```bash
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