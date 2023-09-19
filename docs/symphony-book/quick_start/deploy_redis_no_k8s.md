# Symphony Quick Start - Deploying a Redis container with standalone Symphony

_(last edit: 9/18/2023)_

This quick start walks you through the steps of setting up a new Symphony control plane in standalone mode and deploying a new Symphony solution instance to your local machine using Docker.

> **NOTE**: The following steps are tested under a Ubuntu 20.04.4 TLS WSL system on Windows 11. However, they should work for Linux, Windows, and MacOS systems as well.

## Build and launch Symphony API binary
### 1. Clone the Symphony repo
```bash
git clone https://github.com/azure/symphony
```
### 2. Build the Symphony API binary
```bash
cd api
go build -o symphony-api
```
### 3. Launch the Symphony API in standalone mode
```bash
./symphony-api -c ./symphony-api-no-k8s.json -l Debug
```

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
Next, you define your current machine as a [Target](../uom/target.md) with a Docker [target provider](../providers/target_provider.md):

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
Next, you define a [Solution](../uom/solution.md) with a single Redis container as a component:

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
Now, you define an [Instance](../uom/instance.md), which will trigger the Docker provider to deploy the Redis container to you location machine:

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
You should see a ```sample-redis``` container running after a few seconds.
To test state reconciliation, manually remove the container:
```bash
docker rm -f sample-redis
```
You should see the container relaunched after a few seconds.

## Delete
To delete the container, send a ```DELETE`` request:

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
If you run ```docker ps``` again, you should see the container has been terminated.

> **NOTE**: The standalone Symphony API uses a in-memory state store by default. If you shut down the ```symphony-api``` process, all states will be purged.