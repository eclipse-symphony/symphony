# Activations

An activation is a workflow trigger for a campaign. An activation can trigger a campaign from a middle stage by defining its initial stage in the spec and also with customized input values.

For more information about how Symphony approaches workflows, see [Workflows](../workflows.md).

## Activation cleanup
There is a background job in Symphony to cleanup activations finished for a long time. The default cleanup duration is 180 days. Config can be modified to change the cleanup duration or even disable the background job.

Below is the configuration for backgroundjob vendor. Change the RetentionDuration in the configuration if you want to change the cleanup duration. The supported units are "ns", "us" (or "µs"), "ms", "s", "m", "h". And you can simply remove the manager from the vendor definition if you want to disable the cleanup job.
``` json
{
        "type": "vendors.backgroundjob",
        "route": "backgroundjob",
        "loopInterval": 3600,
        "managers": [
          {
            "name": "activations-cleanup-manager",
            "type": "managers.symphony.activationscleanup",
            "properties": {
              "providers.persistentstate": "k8s-state",
              "singleton": "true",
              "RetentionDuration": "4320h"
            },
            "providers": {
              "k8s-state": {
                "type": "providers.state.memory",
                "config": {}
              }
            }
          }
        ]
      }
```