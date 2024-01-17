# Job vendor

The job vendor subscribes to a `job` topic, which contains reconciliation requests. The reconciliation request is a simple JSON payload with an `objectType` of either `instance` or `target`, an `id` of the object, and a desired action, which can be either `UPDATE` or `DELETE`. For example:

```json
{
  "metadata": {
    "objectType": "instance"
  }, 
  "body": {
    "id": "my-instance-1",
    "action": "UPDATE"
  }
}
```

Upon receiving a job, the job vendor creates a [deployment](../concepts/unified-object-model/deployment.md) object and sends the object to the Symphony solution vendor.

The job vendor can also be configured to trigger periodical reconciliation jobs by enabling the `poll.enabled` property of the `managers.symphony.jobs`:

```json
{
  "type": "vendors.jobs",
  "route": "jobs",
  "loopInterval": 15,
  "managers": [
    {
      "name": "jobs-manager",
      "type": "managers.symphony.jobs",
      "properties": {
        "providers.state": "mem-state",
        "baseUrl": "http://symphony-service:8080/v1alpha2/",
        "user": "admin",
        "password": "",
        "interval": "#15",
        "poll.enabled": "false"               
      },
      "providers": {
        "mem-state": {
          "type": "providers.state.memory",
          "config": {}
        }
      }
    }
  ]
}
```

## Additional routes

The job vendor also offers the following routes:

| Route | Method | Function |
|--------|--------|--------|
| /jobs | GET | Displays last 20 `trace` events |
| /jobs | POST | Publishes a new `trace` event |

These routes can be used to test pub/sub system configurations. They are not used for any other purposes.
