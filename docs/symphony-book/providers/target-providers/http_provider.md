# providers.target.http

This provider triggers a HTTP web hook. It’s commonly used in a [gated deployment](../scenarios/gated-deployment.md).

Deployment is considered successful if the web hook returns a `200` response.

**ComponentSpec** properties are mapped as the following:

| ComponentSpec Properties| HTTP Provider|
|--------|--------|
| `Type` | `http`|
| `Properties[http.url]` | HTTP URL |
| `Properties[http.body]` | HTTP body<sup>1</sup> |
| `Properties[http.method]` | HTTP method, default is `POST` |

1: You can use a few replacement functions in the body string, including `$instance()`, `$solution()` and `$target()`, which correspond to the current [Instance](../concepts/unified-object-model/instance.md) name, the current [Solution](../concepts/unified-object-model/solution.md) name and the current [Target](../concepts/unified-object-model/target.md) name.

The HTTP provider can’t reconstruct the current state, so it always reports its current state as null when asked. This means that the http web hook will be periodically invoked (because the current state remains unknown). Hence, the corresponding web hook is required to be **idempotent** to avoid unwanted side effects.
