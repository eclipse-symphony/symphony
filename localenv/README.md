# Local environment

The local environment is a minikube cluster that deploys symphony for testing purposes.

See all commands with `mage -l`


# Getting started

Use the `mage UpClean` command to start the local environment. Any existing
minikube cluster will be removed.

Symphony images will be built locally with the `local` tag and deployed to the cluster.

```
mage UpClean
```

Use `k9s` or `kubectl` to inspect the cluster once it is running.

# Local development

Bring up the cluster with symphony deployed with:

```
mage Up
```

Building and deploying can be controlled separately with:

```
mage Build
mage Deploy
```

To remove symphony from the cluster use

```
mage Destroy
```


# Integration tests

CI integration tests can be run locally with the following command:

```
mage SetupIntegrationTests
```

See [integration test README](../test/integration/README.md) for more details.