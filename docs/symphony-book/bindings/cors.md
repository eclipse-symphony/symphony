# CORS middleware

The CORS middleware allows you to control the behavior of [cross-origin resource sharing (CORS)](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing).

The CORS middleware is plugged into an [HTTP binding](../bindings/http-binding.md) via the binding’s [pipeline](../bindings/http-binding.md#pipeline) configuration, for example:

```json
"pipeline": [
  {
    "type": "middleware.http.cors",
    "properties": {
      "Access-Control-Allow-Headers": "authorization,Content-Type",
      "Access-Control-Allow-Credentials": "true",
      "Access-Control-Allow-Methods": "HEAD,GET,POST,PUT,DELETE,OPTIONS",
      "Access-Control-Allow-Origin": "*"
    }
  }
]
```
