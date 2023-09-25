# Campaigns

A Campaign describes a workflow of multiple Stages. When a Stage finishes execution, it runs a ```StageSelector``` to select the next stage. The Campaign stops execution if no next stage is selected.

## Stages

In the simplest format, a Stage is defined by a ```name```, a ```provider```, and a ```stageSelector```.

Actions in each stage are carried out by a stage provider. Symphony ships with a few providers out-of-box, and it can be extended to include additional stage providers. In the current version, Symphony ships with the following stage providers:

| provider | description |
|--------|--------|
| ```providers.stage.create``` | Creates a Symphony object like ```Solutions``` and ```Instances``` |
| ```providers.stage.http``` | Sends a HTTP request and wait for a response |
| ```providers.stage.list``` | Lists objects like ```Instances``` and sites |
| ```providers.stage.materialize``` | Materializes a ```Catalog``` as a Symphony object |
| ```providers.stage.mock``` | A mock provider for testing purposes |
|```providers.stage.patch``` | Patches an existing Symphony object|
| ```providers.stage.remote``` | Executes an action on a remote Symphony control plane |
| ```providers.stage.wait``` | Wait for a Symphony object to be created |

## Stage Selectors
When a stage finishes execution, its stage selector is evaluated to decide the next stage to be executed. A stage selector is a Symphony expression that evaluates to a stage name (string). For example, the expression "```next-stage```" selects a stage with the name "```next-stage```".

A stage selector can be used to contruct **branches** and **loops** in a workflow. For example, the following expression selects either a "```success```" stage or a "```failed```" stage based on the value of stage output["```status```"]:

```$if($equal($output(my-stage,status),200),success,failed)```

And the follwing expression creates a loop based on a ```foo``` counter (assuming the counter is incremented by the stage provider):

```$if($lt($output(foo), 5), mock, '')```

A workflow stops when no next stages are selected.

## Stage Contexts
