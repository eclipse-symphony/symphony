# Multi-site distribution

Selected Symphony Catalog objects are automatically synchronized from a parent Symphony instance to all its children. This allows a headquarters to centrally control some artifacts such as application templates and standardized configurations, while each connected child performs autonomous state seeking. Symphony can scale out to a large scale with this kind of cascade deployment topology. Comparing to a single control plane layer approach, this topology enables:

* Both online and offline operations. This allows business continuity on factory floors when a child instance is disconnected from the cloud.
* Break-glass scenarios. For emergencies, local operators can directly attach to the local control plane to override global settings, or to temporarily cut off the system from the cloud.
* Delegated subsystems. When a subsystem is delegated to a vendor, the vendor can set up self-contained Symphony instance to manage local assets, and report statuses back to the global Symphony instance. In this setting, the vendor has complete control of the desired state of local systems. And states are aggregated to the headquarters.

## Configuration distribution process

1.	A HQ user defines a standardized configuration Catalog object
2.	This object is synchronized to all connected child Symphony instances. There’s a continuous synchronization background process. And when a Campaign requires specific Catalog objects, the priority of these objects are raised so that they are synchronized earlier.
3.	Once the object is synchronized, it can be “materialized” into solid object types like Solutions or local Catalog objects.

See the [multi-site example](../scenarios/multisite-deployment.md) for more details.
