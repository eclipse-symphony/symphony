<!--
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
SPDX-License-Identifier: MIT
-->
# Integration tests

## Starting the cluster


From in /localenv run:

```
mage test:up
```

This command will delete minikube if it is running, start a new cluster. Then load up the images.

To speed things up for development you can re-use the cluster if it is already running and manually run localenv commands as needed.

The Redis pub/sub provider retries its startup connection for up to 30 seconds, using exponential backoff from 100 milliseconds capped at 2 seconds. This tolerates Redis Service startup delays; persistent failures still abort API initialization with an error. Images must be rebuilt to include this behavior.

## Running tests

Run all tests with

```
mage Test
```

The suite runs `04.workflow` with labeling disabled (`mage test false`). To test labeling explicitly, run `mage test true` from that scenario's directory.

The workflow scenario collects logs and cleans up on both success and failure. The `05.catalogNconfigmap` verification checks its named targets, instances, and activations in the namespace under test, rather than assuming the namespace contains no other resources. ConfigMaps are checked in `default`, matching the instance's explicitly configured deployment scope.

The `03.basicWithNsDelete` scenario stops before applying manifests or running verification if cluster deployment fails; its test entry point still collects logs and cleans up.

The `00.unit` scenario runs provider tests from their directories in the API module so Go uses the API module's dependency manifest and checksums, rather than those of the integration-test module.

The scenario also creates a private temporary home and a flattened kubeconfig for the current context, embedding the local certificate/key/CA data before running the provider tests. Only the test subprocesses use this home and kubeconfig; your original kubeconfig and Go caches are unchanged, and the temporary credentials are removed afterward. This lets minikube's file-based credentials work with the providers' kubeconfig validation without relaxing that validation.

## Running a single test

To run a single test you can provide the path a test folder.

```
mage RunTest scenarios/basic
```

## Debugging tests

Steps to debug a test:
```
# Start the cluster
cd test/localenv
mage test:up

# Go to the test folder
cd ../test/integration/scenarios/basic

# Run setup to prepare the cluster
mage Setup

# Run the tests
mage Verify
```

### Debugging tests in VS Code
If a test is failing you can debug it in VS Code.
1. open the `test/integration` folder in VS Code.
1. set breakpoints in the test
1. Click debug test above the test function name

Compare the test with the state of the cluster using k9s of kubectl to track down the problem.

## Adding tests

Copy an existing test such as `scenarios/basic` and modify it to test your scenario.
Make sure the folder is under `scenarios/` so it will get picked up automatically.

Tests MUST have:
* A `magefile.go` with a `Test` target that deploys and runs the entire test.

Tests SHOULD have:
* `Setup`, `Verify`, and `Cleanup` targets to make it possible to get the cluster into the same state using console commands.