# Local environment

The local environment is a minikube cluster that deploys symphony for testing purposes.

See all commands with `mage -l`


# Deploying to local env

```
mage minikube:start
mage Deploy
```

# Integration tests

CI integration tests can be run locally with the following command:

```
mage SetupIntegrationTests
```