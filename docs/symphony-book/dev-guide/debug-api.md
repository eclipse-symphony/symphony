# Debug Symphony API locally

## Use unit tests

The best way to test a specific Symphony API component is to write/use unit test cases in the project to test the component in isolation. With VS Code extensions like [Code Debugger](https://marketplace.visualstudio.com/items?itemName=wowbox.code-debuger), you can set up break points and trace through the code, as shown in the following screenshot:

![debug](../images/debug.png)

> **NOTE**: If you use VS Code with a WSL folder, please make sure you have the [WSL extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl) installed.

## Use Postman

Symphony exposes a REST API, which you can call using tools like [Postman](https://www.postman.com/).

1. Build and launch Symphony API as a local process:

   ```bash
   # folder: symphony/api
   go build -o symphony-api
   ./symphony-api -c ./symphony-api-no-k8s.json -l Debug
   ```

1. Use Postman to send API calls to `http://localhost:8080/v1alpha2`.

1. Add temporary log statements to trace what's happening, rebuild and re-launch, and use the console to observe logs.

## Sample requests

1. Deploy a Redis server as a new `instance`.

   > **NOTE**: This request assumes that your `kubectl` is configured to use your target K8s cluster as the default context. If you want to use a different cluster, either update your `kubectl` settings to use the new cluster as the default context (using `kubectl config use-context <context-name>`), or modify the Symphony API configuration file (like `symphony-api-dev.json`) and update the corresponding target provider settings.

   * Method: POST
   * Path: http://localhost:8080/v1alpha2/solution/instances
   * Body:

     ```json
     {
         "instance": {
             "scope": "default",
             "name": "redis-instance",
             "solution": "my-solution",
             "target": {
                 "name": "my-k8s"
             }                 
         },
         "solution": {
             "name": "my-solution",
             "components": [
                 {
                     "name": "redis-server",
                     "properties": {
                         "type": "docker",
                         "container.image": "docker.io/redis:6.0.5"
                     }
                 }
             ]
         },
         "targets": {
             "my-k8s": {
                 "topologies": [
                     {
                         "bindings": [
                             {
                                 "role": "instance",
                                 "provider": "providers.target.k8s",
                                 "config": {
                                     "configType": "path"
                                 }
                             }
                         ]
                     }
                 ]
             }
         },
         "assignments": {
             "my-k8s": "{redis-server}"
         },
         "componentStartIndex": 0,
         "componentEndIndex": 0                
     }

2. Delete the above instance:

    * Method: DELETE
    * Path: http://localhost:8080/v1alpha2/solution/instances?name=redis-instance
    * Body: Same as above.

    > **NOTE**: Symphony requires the deployment object to be posted during deletion because it aims to make providers stateless. The state is played back to the provider to make decisions.
