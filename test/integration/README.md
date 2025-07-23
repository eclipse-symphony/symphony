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


## Running tests

Run all tests with

```
mage Test
```

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

## Remote Agent Tests

Remote agent tests are located in `test/e2e/remote-agent-integration/` and test the communication between Symphony and remote agents using different protocols.

### Running HTTP Communication Tests

To run the HTTP communication tests:

```bash
cd test/e2e/remote-agent-integration
go test -v ./scenarios/http-communication/ -timeout 30m
```

### Running MQTT Communication Tests

To run the MQTT communication tests:

```bash
cd test/e2e/remote-agent-integration
go test -v ./scenarios/mqtt-communication -timeout 30m
```

These tests verify:

- Remote agent bootstrap process
- HTTP/HTTPS and MQTT communication protocols
- Certificate-based authentication
- Target registration and topology updates
- End-to-end deployment

## Adding tests

Copy an existing test such as `scenarios/basic` and modify it to test your scenario.
Make sure the folder is under `scenarios/` so it will get picked up automatically.

Tests MUST have:

* A `magefile.go` with a `Test` target that deploys and runs the entire test.

Tests SHOULD have:

* `Setup`, `Verify`, and `Cleanup` targets to make it possible to get the cluster into the same state using console commands.