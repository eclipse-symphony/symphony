# Symphony quickstart - Deploy a Redis container with standalone Symphony

_(last edit: 9/18/2023)_

This quick start walks you through the steps of setting up a new Symphony control plane in standalone mode and deploying a new Symphony solution instance to your local machine using Docker.

> **NOTE**: The following steps are tested under a Ubuntu 20.04.4 TLS WSL system on Windows 11. However, they should work for Linux, Windows, and MacOS systems as well.

## Prerequisites

* Build and launch Symphony API binary on your development machine. For more information, see [Use Symphony as a binary](./quick_start_binary.md).
* [Docker](https://www.docker.com/) on your development machine.

## Authentication

Using a web client, send the following request:

* **ADDRESS**: http://localhost:8082/v1alpha2/users/auth
* **METHOD**: POST
* **BODY**:

  ```json
  {
    "username": "admin",
    "password": ""
  }
  ```

The response body contains an access token, which you need to attach to the subsequent requests as a **bearer token**.

```json
{
  "accessToken": "eyJhbGci...",
  "tokenType": "Bearer"
}
```

## Define a target

Define your current machine as a [target](../concepts/unified-object-model/target.md) with a Docker [target provider](../providers/target_provider.md):

* **ADDRESS**: http://localhost:8082/v1alpha2/targets/registry/sample-docker-target
* **METHOD**: POST
* **BODY**:

  ```json
  {
    "displayName": "sample-docker-target",
    "forceRedploy": true,
    "topologies": [
      {
        "bindings": [
          {
            "role": "instance",
            "provider": "providers.target.docker",
            "config": {}
          }
        ]
      }
    ]
  }
  ```

## Define a solution

Define a [solution](../concepts/unified-object-model/solution.md) with a single Redis container as a component:

* **ADDRESS**: http://localhost:8082/v1alpha2/solutions/sample-redis
* **METHOD**: POST
* **BODY**:

  ```json
  {
    "displayName": "sample-redis",
    "components": [
      {
        "name": "sample-redis",
        "type": "container",
        "properties": {
          "container.image": "redis:latest"
        }
      }
    ]
  }
  ```

## Define an instance

Define an [instance](../concepts/unified-object-model/instance.md), which triggers the Docker *provider* to deploy the Redis container *solution* to your location machine *target*:

* **ADDRESS**: http://localhost:8082/v1alpha2/instances/redis-server
* **METHOD**: POST
* **BODY**:

  ```json
  {
    "displayName": "redis-server",
    "name": "default",
    "solution": "sample-redis",
    "target": {
      "name": "sample-docker-target"
    }        
  }
  ```

## Validate

Use Docker CLI to check running containers:

```bash
docker ps
```

You should see a `sample-redis` container running after a few seconds.

To test state reconciliation, manually remove the container:

```bash
docker rm -f sample-redis
```

You should see the container relaunched after a few seconds.

## Delete

To delete the container, send a `DELETE` request:

* **ADDRESS**: http://localhost:8082/v1alpha2/instances/redis-server
* **METHOD**: DELETE
* **BODY**:

  ```json
  {
    "displayName": "redis-server",
    "solution": "sample-redis",
    "target": {
      "name": "sample-docker-target"
    }        
  }
  ```

If you run `docker ps` again, you should see that the container has been terminated.

> **NOTE**: The standalone Symphony API uses a in-memory state store by default. If you shut down the `symphony-api` process, all states will be purged.
