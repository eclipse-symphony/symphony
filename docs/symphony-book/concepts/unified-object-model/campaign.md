# Campaigns

A campaign describes a workflow of multiple stages. When a stage finishes execution, the stage runs a `StageSelector` to select the next stage. The campaign stops execution if no next stage is selected.

For more information about how Symphony approaches workflows, see [Workflows](../concepts/workflows.md).

## Stages

In the simplest format, a stage is defined by a name, a provider, and a stage selector.

Actions in each stage are carried out by a stage provider. Symphony ships with a few providers out-of-box, and it can be extended to include additional stage providers. In the current version, Symphony ships with the following stage providers:

| provider | description |
|--------|--------|
| `providers.stage.counter` | Keeps track of multiple variables. For more information, see [Counter stage provider](../../providers/stage-providers/counter.md). |
| `providers.stage.create` | Creates a Symphony object like `Solutions` and `Instances`. |
| `providers.stage.delay` | Delay execution. For more information, see [Delay stage provider](../../providers/stage-providers/delay.md). |
| `providers.stage.http` | Sends a HTTP request and wait for a response. |
| `providers.stage.list` | Lists objects like `Instances` and sites. |
| `providers.stage.materialize` | Materializes a `Catalog` as a Symphony object. |
| `providers.stage.mock` | A mock provider for testing purposes. |
| `providers.stage.patch` | Patches an existing Symphony object. |
| `providers.stage.remote` | Executes an action on a remote Symphony control plane. |
| `providers.stage.script` | Executes a shell script or a PowerShell script. |
| `providers.stage.wait` | Waits for a Symphony object to be created. |

## Stage interface

Each stage is carried out by a stage provider. The stage provider is defined by a simple `IProcessor` interface:

```go
type IStageProvider interface {
	// Return values: map[string]interface{} - outputs, bool - should the activation be paused (wait for a remote event), error
	Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error)
}
```

Essentially, a stage provider takes the inputs, performs any actions, and returns the outputs. When a stage is invoked, activation inputs are provided through the `inputs` parameter, plus any input parameter declared on the stage itself. For example, if a campaign is activated with inputs `foo` and `bar`, and the stage definition contains an input `baz`, the `inputs` parameter will contain all the three values. 

> **NOTE**: In current version, inputs defined in stage definition override activation inputs. If you do want to keep the activation input, you can use expression `$input(<field name>)` in your stage definition to carry over activation input.

## Stage selectors

When a stage finishes execution, its stage selector is evaluated to decide the next stage to be executed. A stage selector is a Symphony expression that evaluates to a stage name (string). For example, the expression "`next-stage`" selects a stage with the name "`next-stage`".

A stage selector can be used to construct **branches** and **loops** in a workflow. For example, the following expression selects either a "`success`" stage or a "`failed`" stage based on the value of stage output["`status`"]:

`${{$if($equal($output(my-stage,status),200),success,failed)}}`

And the following expression creates a loop based on a `foo` counter (assuming the counter is incremented by the stage provider):

`${{$if($lt($output(foo), 5), mock, '')}}`

A workflow stops when no next stages are selected.

## Stage contexts

Stage contexts allow you to define simple **map-reduce** activities in your workflow. For example, after you enumerate a list of sites, you can fan out a deployment to all these sites from your HQ. The deployments are carried out on individual sites and the results are aggregated back to the HQ. If you attach a `contexts` list to a stage, the stage will be triggered for each of the elements defined in the list and run in parallel. Symphony waits for all the elements to finish execution, aggregates the results, and then evaluates the stage selector to select the next stage.

For example, the following stage takes the `items` list from a list stage, and invokes the remote provider for each item in the list:

```yaml
deploy:
  name: deploy
    provider: providers.stage.remote
    stageSelector: ""
    contexts: "${{$output(list,items)}}"
    inputs:
      operation: materialize
      names:
      - site-app
      - site-instance
```
