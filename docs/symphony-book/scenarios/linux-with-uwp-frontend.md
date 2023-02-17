# Scenario: Linux container with a UWP frontend

In this scenario, you’ll deploy a **Solution** with a Linux container backend running on a Linux **Target** and a UWP frontend running on a Windows 10 machine.

![uwp](../images/uwp.png)

## Prerequisites

* A Windows 10 Client machine
* A Windows gateway machine, which can be another Windows 10 machine. It can also be the same machine as your target Windows 10 machine. You'll be running a Python version provider on this machine.
* Windows SDK 1803 or higher is installed on the gateway machine.
* Visual Studio 2022 for building the sample UWP app package.
* Python 3.10 or higher.

## 1. Prepare the UWP application package

You can use any UWP application package of your choice. The signing certificate must be trusted on the target machine.

Or, you can build the sample **Home Hub** UWP app from this repository. You can find the Visual Studio solution under the ```samples/scenarios/homehub/uwp-app``` folder. Simply open the solution and rebuild the application package. Note you may need to switch to your own signing certificate.

Once the pakage is created, you should have somehting like ```HomeHub.Package_1.0.x.0_x64_Debug.appxbundle``` under the build directory. When you configure your Python provider, you need to enter this path to your Python code.

## 2. Prepare the Python provider
You can find the source code of the sample Python provider under the ```samples/scenario/homehub/python-provider``` folder. Before launching it, you need to update the ```winappinstall.py``` file and change the ```WINAPPINSTALL``` constant and the ```TARGET_IP``` constant to match with your environment. 
Once the file is changed, you cna launch the provider with Python:
```cmd
py winappinstall.py
```

## 3. Define the Target
Your target definition should have a proxy provier enabled. The ```serverUrl``` should be pointing to the gateway machine where you are running the Python provider:
```yaml
apiVersion: fabric.symphony/v1
kind: Target
metadata:
  name: mac-book
spec:
  topologies:
  - bindings:
    - role: instance
      provider: providers.target.k8s
      config:
        inCluster: "true"
    - role: win.uwp
      provider: providers.target.proxy
      config:
        name: proxy
        serverUrl: "http://<your gateway where Python provider is running>:8090/"
```
## 4. Define the solution
The following solution defines a nginx container and a UWP app. The ```app.package.pth``` should point to the full path to the```appxbundle``` file.
```yaml
apiVersion: solution.symphony/v1
kind: Solution
metadata:
  name: home-hub
spec:
  components:
  - name: backend
    type: container
    properties:
      container.image: "nginx"
  - name: HomeHub_1.0.7.0_x64
    type: win.uwp
    properties:
      app.package.path: "E:\\HomeHub.Package_1.0.7.0_Debug_Test\\HomeHub.Package_1.0.7.0_x64_Debug.appxbundle"
```

## 5. Define the instance
The instance simply maps the solution to your target. Note the Windows 10 target is not explictly specified as it's proxed. 
```yaml
apiVersion: solution.symphony/v1
kind: Instance
metadata:
  name: homehub-instance
spec:
  solution: home-hub
  target:
    name: mac-book
```

## 6. Deploy
```
kubectl create -f mac-target.yaml
kubectl create -f homehub-app.yaml
kubectl create -f homehub-instance.yaml
```

Observe the app getting deployed to the target Windows 10 machine.

## 7. Remove Instance
```
kubectl delete instance homehub-instance
```

Observe the app is removed from the target Winows 10 machine.

## 8. Clean up
```
kubectl delete solution home-hub
kubectl delete target mac-book
```