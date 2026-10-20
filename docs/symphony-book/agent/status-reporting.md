# Agent status reporting

An agent can load `vendors.status` alongside its deployment vendors to periodically
report local observations to a Target. Reporting does not depend on deployment
reconciliation, `Get()`, or `forceRedeploy`.

The vendor schedules a `managers.symphony.status` manager. The manager evaluates an
agent-local property bag on every tick and passes the result to an `IReporter`.
`providers.reporter.mqtt` transports reports over MQTT; the existing
`providers.reporter.http` can also be selected without changing the manager.

## Agent configuration

Add the following vendor to the agent's `api.vendors` array:

```json
{
  "type": "vendors.status",
  "loopInterval": 30,
  "managers": [
    {
      "name": "status-manager",
      "type": "managers.symphony.status",
      "properties": {
        "poll.enabled": "true",
        "target.name": "$env:TARGET_NAME",
        "target.namespace": "default",
        "providers.reporter": "mqtt-reporter",
        "report.properties": "{\"agent.site\":\"$env:SITE_NAME\",\"agent.hostname\":\"$env:HOSTNAME\",\"agent.description\":\"host-${{ $property('agent.hostname') }}\",\"agent.connected\":\"true\"}"
      },
      "providers": {
        "mqtt-reporter": {
          "type": "providers.reporter.mqtt",
          "config": {
            "name": "status-reporter",
            "brokerAddress": "$env:MQTT_BROKER_ADDRESS",
            "clientID": "$env:STATUS_MQTT_CLIENT_ID",
            "requestTopic": "symphony/status/edge-01/request",
            "responseTopic": "symphony/status/edge-01/response",
            "timeoutSeconds": 8
          }
        }
      }
    }
  ]
}
```

Set `TARGET_NAME` to the name of an existing Target, `SITE_NAME` and `HOSTNAME` to
the observations to report, and `STATUS_MQTT_CLIENT_ID` to a unique MQTT client ID.
On systems without `HOSTNAME`, select an environment variable available to the
agent process instead. The example topics identify `edge-01`; use topics
appropriate to the actual agent and configure the receiver to match.

The vendor requires a positive `loopInterval` in seconds. The first report is sent
after that interval, followed by one report per tick. Set `poll.enabled` to
`"false"` to disable reporting. Reporting is sequential; a slow report can delay
the next tick. To report multiple Targets, configure a manager for each Target.
Use distinct reporter client IDs for separate MQTT reporter instances.

### Property bag

`report.properties` is a JSON-encoded object of string values, because manager
configuration properties are strings. Supported values include:

* Literal strings, such as `"true"`.
* `$env:NAME`, resolved from the **agent process environment on every tick**.
* Symphony expressions, such as `${{ 6 * 7 }}` or
  `host-${{ $property('agent.hostname') }}`. Property references use this local
  bag. Expressions requiring configuration or secret providers require the
  corresponding evaluation context to be available on the agent; deployment-only
  context is not supplied by the status manager.

Unset environment variables, null expression results, and expression failures
fail the entire report and are logged. No partial report is sent. Empty but
explicitly set environment variables are permitted.

The manager adds the reserved `agent.reportedAt` property as a UTC RFC 3339
timestamp. An empty bag (`"{}"`) can therefore be used for timestamp-only
reporting. The bag cannot redefine `agent.reportedAt`.

The bag is local configuration, not part of the Target's desired state. Changing
it requires updating/restarting the agent. Configure only observations that are
safe to publish: there is no automatic secret redaction or environment allowlist
beyond the locally configured bag. Do not include credentials or tokens.

## Control-plane receiver

For a complete CLI-driven local demo, including API/agent binary overrides, a
managed test broker, Target registration, and status watching, see the
[Maestro local demo](../../../cli/README.md#local-agent-status-demo).

Reports use the existing `vendors.agent` reporting endpoint backed by a reporter
such as `providers.reporter.k8s`. The standard API configuration already includes
this vendor. Add an MQTT binding with the same report topics and a different,
unique client ID:

```json
{
  "type": "bindings.mqtt",
  "config": {
    "brokerAddress": "tcp://localhost:1883",
    "clientID": "status-control-plane",
    "requestTopic": "symphony/status/edge-01/request",
    "responseTopic": "symphony/status/edge-01/response",
    "timeoutSeconds": 8
  }
}
```

Replace the binding's example `brokerAddress` with the same broker URL used by
the agent. Unlike the reporter, the MQTT binding does not resolve `$env:` values
itself.

Alternatively, run a dedicated receiver process with just this vendor in its
`api.vendors` array and the binding above:

```json
{
  "type": "vendors.agent",
  "route": "agent",
  "managers": [
    {
      "name": "reference-manager",
      "type": "managers.symphony.reference",
      "properties": {
        "providers.volatilestate": "memory",
        "providers.reporter": "k8s-reporter"
      },
      "providers": {
        "memory": {
          "type": "providers.state.memory",
          "config": {}
        },
        "k8s-reporter": {
          "type": "providers.reporter.k8s",
          "config": {
            "inCluster": true
          }
        }
      }
    }
  ]
}
```

For a receiver outside Kubernetes, set `inCluster` to `false` and configure
`configPath` for the Kubernetes reporter. The receiver needs permission to get
Targets and update their status subresource. It does not create missing Targets.

The MQTT reporter sends a COA `POST` request to route `references`, including
Target identity, the evaluated bag, and a unique `request-id`. The binding
dispatches it to `agent/references` and echoes the request ID in the response.
The reporter waits for that response: a successful broker publication alone is
**not** considered a successful status update.

Reported properties are merged into `Target.status.properties` with
`overwrite=false`, preserving other properties and provisioning status. Removing
a property from the local bag does not delete its previously reported value.

```yaml
status:
  properties:
    agent.site: factory-a
    agent.hostname: edge-01
    agent.description: host-edge-01
    agent.connected: "true"
    agent.reportedAt: "2026-10-20T18:00:00Z"
```

`agent.connected` in this example is a configured observation, not a built-in
connectivity controller. Reports stopping does **not** automatically change it
to `"false"`. Consumers must treat old timestamps as stale; clock-sensitive
decisions should account for agent clock skew. This feature does not add an
expiry controller or an MQTT Last Will.

### Standalone Target receiver

When the control plane uses standalone state storage, `vendors.targets` exposes
`POST targets/report` (MQTT route key `report`). This accepts the same flat
property bag and resource identity parameters used by reporter providers, then
merges the report through the existing Targets manager. It works with the
Target state provider configured on that manager, including standalone memory
storage; a Kubernetes reporter is not required.

Set the MQTT reporter's `route` to `"report"` and configure a control-plane MQTT
binding on the report topics. The endpoint requires `group=fabric.symphony`,
`kind=targets`, `version=v1`, and a nonempty `id`; `namespace` defaults to
`default`. It supports merging only (`overwrite=false`) and does not create a
missing Target. Maestro configures this route and binding automatically for
`up --no-k8s --with-mqtt-broker`.

As with the existing agent endpoint, MQTT access requires trusted publishers and
appropriate broker controls. This endpoint does not bind an MQTT identity to a
specific Target.

## MQTT options and failures

Required reporter options are `brokerAddress`, `clientID`, `requestTopic`, and
`responseTopic`. Topics must be distinct and must not contain MQTT wildcards.
The optional `route` defaults to `references`, the route key used by the MQTT
binding.

`timeoutSeconds` defaults to 8, `keepAliveSeconds` to 2, and
`pingTimeoutSeconds` to 1. Timeouts apply separately to connection, subscription,
publication, and status acknowledgement. The reporter connects lazily and
reconnects/resubscribes on a subsequent tick after a lost connection. A failed
report is logged; the next tick evaluates fresh observations rather than
replaying an old report. Reports are not retained.

The reporter supports the binding's certificate settings:

```json
{
  "brokerAddress": "ssl://mqtt.example.com:8883",
  "useTLS": "true",
  "caCertPath": "/etc/symphony/mqtt/ca.pem",
  "clientCertPath": "/etc/symphony/mqtt/client.pem",
  "clientKeyPath": "/etc/symphony/mqtt/client.key",
  "username": "$env:MQTT_USERNAME",
  "password": "$env:MQTT_PASSWORD"
}
```

Configure equivalent TLS settings on the receiver binding, as appropriate.
`useTLS` and `insecureSkipVerify` are string booleans, matching the MQTT binding.
Certificate paths and authentication settings on the reporter support `$env:NAME`.
Keep certificate verification enabled in production.

**Trust boundary:** MQTT bindings expose the configured vendor endpoints without
the HTTP authentication middleware. The existing reporting endpoint accepts
resource identity from the request; a topic does not itself constrain the
payload to one Target. Only trusted publishers should reach it. Use TLS,
broker authentication/topic ACLs, and least-privilege Kubernetes RBAC. A dedicated
receiver with restricted credentials avoids exposing unrelated API vendors over
MQTT. This feature does not add per-agent identity-to-Target authorization.

## Tests

From the `api` directory:

```bash
go test ./pkg/apis/v1alpha1/managers/status ./pkg/apis/v1alpha1/vendors \
  github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/mqtt \
  -run 'Test(Status|MQTTReporter)'
```

The end-to-end test starts a private Mosquitto broker on a dynamically allocated
loopback port and exercises the real MQTT binding and Kubernetes reporter against
a test Kubernetes HTTP endpoint. It skips when `mosquitto` is not on `PATH`.
