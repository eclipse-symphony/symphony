---
type: docs
title: "Counter Campaign"
linkTitle: "Counter Campaign"
description: ""
weight: 100
---

This is a simple campaign that counts from a given value up to 20. It contains a single `providers.stage.counter` stage, which takes an input value, `val`, increments it by 1, and passes the updated value through the trigger for the next stage.

A conditional stage selector selects the next stage. If `val` is less than 20, the selector routes execution back to the same counter stage, creating a loop. Once val reaches 20, the selector chooses an empty stage instead, terminating the campaign.

## Part 1: Run through the sample using `maestro`

> **NOTE:** This sample assumes you've already installed and configured `maestro`.

1. Point `maestro` to the Symphony installation:

    ```bash
    maestro config use-context <context name>
    ```
    You can use `maestro config get-context` to check for available contexts and the current context.


2. Run the counter campaign sample, which is one of the packaged samples:

    ```bash
    maestro samples run counter-campaign
    ```
    You should see outputs like:
    ```bash
    Creating campaign counter-campaign ... done
    Creating campaignversion counter-campaign-v-version1 ... done
    Creating activation counter-activation ... done
    ```
3. Check the activation status:
 
    ```bash
    ./maestro get activation -n counter-activation --json-path @status.stageHistory[0].outputs
    ```
    You should see the activation status is 200 (OK), and the output `val` is 20:
    ```bash
    STATUS  VAL
    200     20
    ```
## Part 2: Understand the campaign 

1. Open `~/.symphony/campaigns/counter/campaign-version.yaml` to examine the campaign definition.
2. The counter stage uses a `providers.stage.counter` that takes an input `val` and increment its value by 1. It then writes the value to its output, which is made avaialble to `stageSelector`:
    ```yaml
    counter:
      name: "counter"
      provider: "providers.stage.counter"
      stageSelector: "${{$if($lt($output(counter,val), 20), counter, '')}}"
      inputs:
        val: "${{$trigger(val, 0)}}"
    ```
3. The `stageSelector` selects the next stage to be executed. In this case, it uses a `if` statement in Symphony expression that checks if the stage output `val` is larger than 20. If so, it selects `counter` as the next stage. Otherwise, it selects an empty stage, which causes the activation to stop. This means the counter stage will be repeatedly called until `val` reaches 20.

## Part 3. Clean up

1. If you run Symphony as a process, simply shut down the process. 
2. If you run Symphony on Kubernetes, delete the Symphony objects:

    ```bash
    kubectl delete activation counter-activation
    kubectl delete campaignversion counter-campaign-v-version1
    kubectl delete campaign counter-campaign
    ```