# Tracing middleware

The Tracing middleware configures a distributed tracing pipeline leveraging [OpenTelemetry](https://opentelemetry.io/).

To add a tracing middleware, configure your [HTTP binding](../bindings/http-binding.md) to include a tracing middleware with a trace exporter. Symphony currently supports:

* Console exporter
* Zipkin exporter

The following example shows how to define a tracing middleware with a Zipkin exporter:

```json
{
  "type": "middleware.http.tracing",
  "properties": {
    "pipeline": [
      {
        "exporter" : {
          "type": "tracing.exporters.zipkin",
          "backendUrl": "http://localhost:9411/api/v2/spans",
          "sampler": {
            "sampleRate": "always"
          }
        }
      }
    ]
  }
}
```
