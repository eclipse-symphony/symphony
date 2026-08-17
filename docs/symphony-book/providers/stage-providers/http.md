# Http stage provider

Http stage provider sends an HTTP request to a given URL and captures the response into stage outputs. Request headers and a JSON body can be supplied through inputs. The provider can also poll a second URL until the polled operation reaches a desired state, which is useful for waiting on long-running operations such as Logic Apps workflows.

All configuration fields can be overridden per execution by providing inputs with the same names.

## Config

| Field | Value |
|-------|-------|
| `url` | Request URL. |
| `method` | HTTP method, such as `GET` or `POST`. |
| `successCodes` | (optional) List of status codes considered successful, such as `[200, 202]`. If set, any other status code fails the stage. |
| `wait.url` | (optional) URL to poll (with `GET`) after the initial request. |
| `wait.start` | (optional) Status codes of the initial response that trigger polling. |
| `wait.success` | (optional) Status codes of the poll response that indicate success. |
| `wait.fail` | (optional) Status codes of the poll response that indicate failure. |
| `wait.interval` | (optional) Seconds to wait between polls. |
| `wait.count` | (optional) Maximum number of polls. `0` means poll until success or failure. |
| `wait.expression` | (optional) Expression evaluated against the poll response body. Polling stops only when the expression succeeds. |
| `wait.expressionType` | (optional) `symphony` (default) or `jsonpath` — the expression language of `wait.expression`. |

## Inputs

| Field | Value |
|-------|-------|
| `<config field>` | (optional) Any config field listed above, overriding the configured value for this execution. |
| `header.<name>` | (optional) HTTP request header, e.g. `header.Content-Type`. |
| `body` | (optional) Request body, serialized as JSON. |

## Outputs

| Field | Value |
|-------|-------|
| `status` | Response status code, e.g. `200`. |
| `body` | Response body as a string. |
| `header.<name>` | Response headers. |
| `waitResult` | Result of `wait.expression`, when polling is used. |
| `waitBody` | Body of the successful poll response, when polling is used. |

## Sample

Call a Logic Apps workflow and continue to the `deploy` stage when the call succeeds:

```yaml
approval:
  name: "approval"
  provider: "providers.stage.http"
  config:
    url: "<Logic Apps Workflow URL>"
    method: "GET"
    successCodes: [200]
  stageSelector: ${{$if($equal($output(approval,status), 200),'deploy','end')}}
```
