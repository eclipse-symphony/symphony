# Workflow

Symphony has a built-in workflow engine that captures complex operational workflows into declarative models. In addition to basic workflow engine features like branches, conditions, loops and stateful stages, Symphony also supports a set of unique workflow features to enable large-scale, automated operations, including:
* Remote execution of stages. 
* Scheduled execution.
* Isolated execution environment per workflow stage.
* Fan-out execution of stages (map-reduce).

## Fundamentals

* [Defining and running a workflow](./define-campaigns.md)
* Inputs and outputs
* States
* Branches and loops

## Stage Processors 

* [Counter Stage Provider](./counter-provider.md)
* [Delay Stage Provider](./delay-provider.md)
* [Mock Stage Provider](./mock-provider.md)

## Advanced

* Scheduling
* [Error handling and retries](./error-handling.md)
* [Stage isolation with provider proxy](./provider-proxy.md)
* Remote execution
* Fan-out execution

## Scenarios

* [Canary Deployment](../scenarios/canary-deployment.md)
* [Approval with Logic Apps](../scenarios/gated-deployment-logic-app.md)
* [Approval with a custom script](../scenarios/gated-deployment-script.md)