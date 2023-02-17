# HTTP Binding
HTTP Binding binds Symphony API to standard HTTP/HTTPS clients.

## Configure HTTP Binding
To set up HTTP binding, modify your [Symphony host configuration file](../hosts/overview.md). The simpliest HTTP binding configuration looks like this:
```json
"bindings": [
    {
      "type": "bindings.http",
      "config": {
        "port": 8098
      }
    }
  ]
```
And to se up a HTTPS binding with a auto-generated self-signed certificate:

> **NOTE:** You can use multiple bindings at the same time.
```json
"bindings": [
    {
      "type": "bindings.http",
      "config": {
          "port": 8081,
          "tls": true,
          "certProvider": {
          "type": "certs.autogen",
          "config":{}
          }
      }
    }
]
```
Please seee [Cert providers](../providers/cert_providers.md) for details on supported certificate providers and their configurations.

## Pipeline
HTTP binding also allows you to define a pipeline of middleware, such as [CORS](./cors.md), [JWT token handler](./jwt-handler.md) and [distributed tracing using OpenTelemetry](./tracing.md). It's expected that other middleware will be enabled in future versions, such as caching, device attestation, and more.

To define a middleware pipeline, add a ```pipeline``` element to the root of your binding config, and follow the formats of individual middleware configurations.
```json
"bindings": [
    {
        "type": "bindings.http",
        "config": {
            ...
            "pipeline": [
                {
                    //middleware 1 config
                },
                ...
                {
                    //middleware n config
                }
            ]
        }
    }
]
```

Upon requests, middleware are applied in the order they are defined in the pipeline. Unpon responses, middleware are applied in the reversed order. A specific middleware may choose to handle only requests, only responses, or both.

> **NOTE:** Please see a few examples of Symphony host configuration files under the root folder of the ```symphony-api``` repo, as explained [here](../hosts/overview.md).