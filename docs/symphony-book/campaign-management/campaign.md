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
| ```providers.stage.materialize``` | Materializes a ```Catalog``` as an Symphony object |
| ```providers.stage.mock``` | A mock provider for testing purposes |
| ```providers.stage.remote``` | Executes an action on a remote Symphony control plane |
| ```providers.stage.wait``` | Wait for a Symphony object to be created |



