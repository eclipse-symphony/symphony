<!--
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
SPDX-License-Identifier: MIT
-->
# Overview

This is a framework for symphony fault tests. 

# How to add a new fault test
1. Add failpoints in codes

Add a comment line like example blow. Gofail package will translate it to the failpoint code when compiled with failpoint enabled.
```
// DO NOT REMOVE THIS COMMENT
// gofail: var beforeProviders string
```

> **Note**: every failpoint referenced by a test case must have a matching marker in the code. A missing marker does not fail the scenario — the injection silently no-ops (`failpoint does not exist` in the logs) and the test passes without testing anything. After adding a case, check the scenario logs for `failpoint does not exist` to confirm the injection is real.

2. Add a new fault test case

There are already two fault tests - solutionversion upgrade and workflow materialize in the faultTests folder. You can add a new case in the [constants.go](./utils/constants.go) you want to use existing fault tests with new failpoint.

For example, you can specify the test, the pod to inject failure, the failpoint name and fault types in the below structure. The most common fault type is `100.0%panic`. And you can also use other faults like sleep, error following [Gofail term](https://github.com/etcd-io/gofail/blob/master/doc/design.md#gofail-term)
```json
{
    testCase:  TestCases["solutionversionUpdate"],
    podLabel:  PodLabels["api"],
    fault:     "onQueueError",
    faultType: DefaultFaultType,
}
```

You can also add new fault tests under faultTests if the existing ones don't hit the new failpoints.

# How injection works

For every probe and injection, the test process dials a fresh port-forward to the current live pod for the case's label and talks to the failpoint server (`localhost:22381` inside the pod) over plain HTTP. Retries re-resolve the pod each time, so the channel recovers when a `panic` fault restarts the pod. Nothing needs to be installed in the target image (no `curl`).

# Run tests

## Local
First build fault images and setup cluster 
```
cd test/localenv
mage build:apifault
mage build:k8sfault
mage cluster:up
```
Trigger fault test.
```
cd test/integration/scenarios/faultTests
mage faulttests
```

## Github Action

Here is a fault [github action](../../../../.github/workflows/fault.yml) and you can trigger the workflow in the Action page.

# Diagnostic

## Local
All the test logs are collected under `/tmp/symphony-integration-test-logs/`

## Github Action
All the test logs are collected in the artifacts