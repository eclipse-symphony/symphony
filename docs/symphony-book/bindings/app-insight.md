# Application Insight middleware

The Application Insight middleware allows Symphony API consumption events to be collected to an Azure Monitor vtenant, which is identified by an instrumentation key defined in a `APP_INSIGHT_KEY` environment variable. For more information, see [Application Insights overview](https://learn.microsoft.com/eclipse-azure-monitor/app/app-insights-overview).

The Application Insight middleware is plugged into an [HTTP binding](../bindings/http-binding.md) via the binding’s [pipeline](../bindings/http-binding.md#pipeline) configuration, for example:

```json
"pipeline": [
  {
    "type": "middleware.http.telemetry",
    "properties": {
      "enabled": true,
      "maxBatchSize": 8192,
      "maxBatchIntervalSeconds": 2,
      "client": "my-dev-machine"
    }
  }
]
```

> **NOTE**: `client` is an optional string identifier for the current customer/installation. This could be set during installation/provision.

The middleware intercepts all Symphony API calls and records request path, method, and status code as a custom event.

![Application Insight](../images/app-insight.png)

Based on these events, you can collect usage information for questions like:

* How many solutions were created/updated in the past month?
* How many targets were deleted today?

Because the API uses upsert to handle both creation and update, there's currently no easy way to distinguish between creation and update. However, given wide adoption of immutable infrastructure, the assumption is that update to solution/instance is less common. Regardless, this should be improved in future versions.
