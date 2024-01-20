# Symphony REST API

Although you can interact with Symphony through standard Kubernetes tools like `kubectl`, you can also use Symphony through its REST API with any web clients. For instance, Symphony’s CLI tool, `maestro`, uses Symphony REST API.

With Symphony’s [binding](../bindings/_overview.md) concept, you can also consume Symphony REST API through other protocols like MQTT. This is useful when you try to set up a local Symphony installation and want to enable remote access to the API without a public IP address.

Symphony API is configurable to use different identity providers. By default, it offers a [built-in provider](../security/authentication.md) that can be configured with predefined usernames and passwords, so that you can run Symphony as a self-contained service.

Symphony API can be hosted and scaled on cloud behind a load balancer or an API gateway. You can also run Symphony API as a container or a standalone process. Symphony API can also be compiled into a Web Assembly with in-memory state stores. This allows you to host a Symphony API server natively in modern browsers.

For more details, see:

* [Instances API](./instances-api.md)
* [Solutions API](./solutions-api.md)
* [Targets API](./targets-api.md)

You can find an Open API definition of Symphony API in [Sypmhony.openapi.yaml](./Symphony.openapi.yaml).
