# Retail Scenario

Before start, please [set up your own kubernetes cluster](https://kubernetes.io/docs/setup/) OR follow the [instruction](../../../test/localenv/README.md) to set up minikube to run symphony. **We recommend the second method since it's easier.**

## Build UWP sample app (in Windows)
0. Search for "For developer" in windows search bar, and then turn on the Developer Mode and Device discovery. Restart you PC to ensure it takes affect.
1. Open `ContosoCafeteriaKiosk/ContosoCafeteriaKiosk.sln` with Visual Studio 2022.
2. Build the solution for `x64` and `x86` architecture, `Debug` profile.
3. Here you should create a new certificate to sign it. Please keep it and trust it on your PC. **(IMPORTANT)**
4. Copy the `ContosoCafeteriaKiosk_1.0.0.0_Debug_Test` folder to a folder that is accessible by the proxy provider (see below).
Reference: [Create an App Installer file with Visual Studio](https://learn.microsoft.com/en-us/windows/msix/app-installer/create-appinstallerfile-vs)

## Deploy a MQTT broker (in WSL)

Please refer to the [instruction](../../../test/localenv/README.md) to set up minikube to run symphony. Here are some command that can be useful:

  ```bash
  cd ~/symphony/test/localenv
  mage build:all
  mage cluster:up
  ```

We'll use a MQTT broker to facilitate communication between Symphony and the remote agent, which you'll run from your machine or on your target device. 
We offer a sample deployment file at `templates/mosquitto.yaml`, which you can use to deploy a [mosquitto](https://mosquitto.org/) test MQTT broker with anoymous access enabled.

  ```bash
  kubectl apply -f ./templates/mosquitto.yaml
  ```

Once deployment is complete, you should see a `mosquitto-service` service in your service list. This will the broker your agents connect to.

  ```bash
  kubectl get svc
  NAME                TYPE           CLUSTER-IP     EXTERNAL-IP       PORT(S)
  ...
  mosquitto-service   LoadBalancer   10.98.133.25   172.179.118.100   1883:32450/TCP
  ...
  ```

If you are using Minikube, the `EXTERNAL-IP` might show as `<pending>` (finally it will be 127.0.0.1). You'll need to use K8s port forwarding to expose the service to your local machine. Then, you'll be able to access the MQTT broker through `tcp://localhost:1883`.

  ```bash
  kubectl port-forward svc/mosquitto-service 1883:1883 &
  ```

If you are using MiniKube, please run `minikube tunnel` in a single terminal windows and keep it open for the rest steps.

## Setup local Symphony on Windows (in Windows)
You can use the [`templates/symphony-agent.json`](./templates/symphony-agent.json) as a template for your local Symphony configuration. A few things to notice:

1. You can define multiple MQTT bindings, each corresponds to a Target object on the control plane side.
2. You can increase default MQTT timeout by modifying the `timeoutSeconds` value. Because provider deployments are blocking, you need to make sure the deployment can finish in time to respond before this time window expires.
3. On the `solution-manager` configuration, you should define local target providers you want to use. The provider name needs to match with the target name.

## Launching the agent (in Windows)

1. Create a `C:\demo` folder and a `C:\demo\staging` folder on your Windows machine.

2. Build the `symphony-api.exe`:

   ```powershell
   # under the api folder of symphony repository
   $env:GOOS="windows"
   $env:GOARCH="amd64"
   go build -o symphony-api.exe
   ```

3. Copy `symphony-api.exe` to the `C:\demo` folder. The Symphony agent and Symphony API share the same binary, driven by different configuration files, which you'll copy next.

4. Copy the `symphony-agent.json` file under the `api` folder to the `C:\demo` folder. This is the configuration file that you'll use to launch the Symphony agent. In a production environment, the Symphony agent can be configured as a Windows service that is automatically launched upon start.


5. Once you've finished previous configuration steps, you can launch a new instance of Symphony Agent through command line (under `C:\demo` folder):

    ```powershell
    .\symphony-api.exe -c symphony-agent.json -l Debug
    ```

## Deploying sample applications (in WSL)

1. Deploy targets and solution:

    ```bash
    kubectl apply -f ./templates/solution.yaml
    kubectl apply -f ./templates/windows-target.yaml
    kubectl apply -f ./templates/k8s-target.yaml
    ```

The docker image used for deploy the backend service is described in the `solution.yaml` file. If you want to change it to another image, you can edit the link.

2. Examine the current state of all targets:

PC: search for "kiosk" app on Windows search - no results should be returned.

K8s: use kubectl get pods to see no application pods are deployed

3. Trigger deployment:

    ```bash
    kubectl apply -f ./templates/instance.yaml
    ```

4. Check deployment status and how those applications look like:

PC: search for "kiosk" app on Windows search - the app should be found. Launch the app.

K8s: use kubectl get pods to see application pod.

If you are using Minikube, you'll need to use K8s port forwarding to expose the service to your local machine.

  ```bash
  kubectl port-forward svc/nginx-ingress-ingress-nginx-controller 5000:80 &
  ```

Then you can use `http://localhost:5000` in web browser to have a look of the backend service. 

5. [optional] Remove the deployment:

    ```bash
    kubectl delete instance retail-instance
    ```

Examine the final state:

PC: search for "kiosk" app on Windows search - no results should be returned.

K8s: use kubectl get pods to see no application pods are deployed
