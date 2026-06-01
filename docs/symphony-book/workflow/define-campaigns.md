# Defining and running a workflow

In Symphony, a workflow is described by a `CampaignVersion` object. A campaignversion contains one or more stages. Each stage is processed by a stage processor.  After each stage, a stage selector is evaluated to select the next stage. When no next stages are selected, the campaignversion finishes. 

The following example shows a simple Symphony campaignversion with a single stage. The stage is handled by a mock processor that simply generates some outputs in Symphony logs.
```yaml
apiVersion: workflow.symphony/v1
kind: CampaignVersion
metadata:
  name: hello-world
spec:
  firstStage: "mock"
  selfDriving: true
  stages:
    mock:
      name: "mock"
      provider: "providers.stage.mock"      
```
A campaignversion defines a workflow. To execute a workflow, you create an `Activation` object. A campaignversion can be activated multiple times. Activations are retained for 24 hours by default. To activate the above workflow, create a new Activation object like the following:
```yaml
apiVersion: workflow.symphony/v1
kind: Activation
metadata:
  name: hello-world-activation
spec:
  campaignversion: "hello-world"
  name: "hello-world-activation"
  inputs:
        foo: "bar"       
```
If you observe Symphony API logs, you can find the outputs from the mock stage:
```txt
====================================================
MOCK STAGE PROVIDER IS PROCESSING INPUTS:
__activation:   hello-world-activation
__stage:        mock
__activationGeneration:         1
__previousStage:        mock
__site:         hq
foo:    bar
__campaignversion:     hello-world
__namespace:    <nil>
----------------------------------------
TIME (UTC)  : 2024-04-07T02:02:32Z
TIME (Local): 2024-04-07T02:02:32Z
----------------------------------------
MOCK STAGE PROVIDER IS DONE PROCESSING WITH OUTPUTS:
__activation:   hello-world-activation
__stage:        mock
__activationGeneration:         1
__previousStage:        mock
__site:         hq
foo:    bar
__campaignversion:     hello-world
__namespace:    <nil>
====================================================
```