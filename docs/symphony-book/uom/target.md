# Target

A target is a computational resource that hosts Symphony payloads. A target can be an Azure IoT Edge device, a Kubernetes cluster, a PC or server, or any other endpoints that support the Symphony [provider interface](../providers/provider_interface.md).

Symphony ships a number of providers out-of-box to support various target types. And Symphony is extensible with either [native providers](../providers/overview.md#provider-types) (that come with Symphony builds), [script providers](../providers/script_provider.md) or [standalone providers](../providers/standalone_providers.md) via a proxy.
