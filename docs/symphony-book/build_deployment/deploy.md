# Deploy Symphony

_(last edit: 4/12/2023)_

## Option 1: Using maestro

Maestro is Symphony’s CLI that allows you to bootstrap Symphony with your Kubernetes clusters, or to run latest Symphony build in standalone mode. It also allows you to quickly deploy sample scenarios using prebuilt samples. It’s a great tool to get started with Symphony quickly! Please see maestro instructions [here](../cli/cli.md).

## Option 2: Using Helm
You can also deploy Symphony to a Kubernetes cluster using Helm 3:
```bash
helm install symphony oci://possprod.azurecr.io/helm/symphony --version 0.43.1
```
Or, to update your existing Symphony release to a new version:
```bash
helm upgrade --install symphony oci://possprod.azurecr.io/helm/symphony --version 0.43.1
```

## Option 3: Using Docker
You can run Symphony API in standalone mode as a Docker container 
```bash
docker run --rm -it  -v /configuration/file/path/on/host:/config -e CONFIG=/config/symphony-api-no-k8s.json possprod.azurecr.io/symphony-api:latest
```
> **NOTE**: You can find various Symphony API configuration files under the ```api``` folder of the Symphony repo. Please see [this doc](../hosts/overview.md) for details on different configurations you can use.
