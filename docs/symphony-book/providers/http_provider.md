# providers.target.http

This provider triggers a HTTP web hook. It’s commonly used in a [gated deployment](../scenarios/gated-deployment.md).

Deployment is considered successful if the web hook returns a ```200``` response.

**ComponentSpec** Properties are mapped as the following:

| ComponentSpec Properties| HTTP Provider|
|--------|--------|
| ```Type``` | ```http```|
| ```Properties[http.url]``` | HTTP URL |
| ```Properties[http.body]``` | HTTP body<sup>1</sup> |
| ```Properties[http.method]``` | HTTP method, default is ```POST``` |

<sup>1</sup>: You can use a few replacement functions in the body string, including ```$instance()```, ```$solution()``` and ```$target()```, which correspond to the current [Instance](../uom/instance.md) name, the current [Solution](../uom/solution.md) name and the current [Target](../uom/target.md) name.
>**NOTE:** This provider can’t reconstruct the current state, hence it always current state as null when asked. This means the http web hook will be periodically invoked (because the current state remains unknown). Hence, the corresponding web hook is required to be **idempotent** to avoid unwanted side effects.

