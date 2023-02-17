# Build Symphony Containers

Symphony has two parts: the platform-agnostic API (symphony-api) and Kubernetes binding (symphony-k8s), both are packaged as Docker containers.

## 0.Prerequisites
* [Go](https://golang.org/dl/) (1.18 or higher)
* [Git](https://git-scm.com/downloads)
* [Docker](https://www.docker.com/products/docker-desktop)
  > **NOTE:** Some tools we use work better with access to docker commands without sudo. Use Docker Desktop version when possible. Otherwise, you need to add your user to docker group (see [instructions](https://www.docker.com/products/docker-desktop)). Please don't use rootless model, which isn't supported by some of the tools.
* A Kubernetes cluster, such as [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
  >**NOTE:** Symphony should work with any Kubernetes clusters on x86 and x64 CPUs. symphony-api container is compiled for ARM as well but we haven't tested other containers.
* [Kubebuilder](https://book.kubebuilder.io/) (optional)
* make and gcc
  ```bash
  sudo apt-get update && sudo apt-get upgrade -y
  sudo apt-get install make gcc -y
  ```
* An Azure subscription
* Visual Studio Code with
  * Go extensions (open any Go source file and select **Install All** tools)
  * [Azure IoT Tools](https://marketplace.visualstudio.com/items?itemName=vsciot-vscode.azure-iot-tools) extension
* Enough memory (>16G), especially if you run Kind cluster on the same machine.
* Kustomize (optional for buidling Helm chart)
* Helm (optional for building Helm chart)

## 1. Clone the repository
You need to clone the following repositories under the **same parent folder**:

> **NOTE:** It's important to keep the repositories under the same parent folder. These go packages are not published, so we are using replace directives to redirect reference locations.

* https://github.com/Azure/coa
* https://github.com/Azure/symphony-api
* https://github.com/Azure/symphony-k8s

## 2. Build Symphony API container
To build multi-platform Symphony API container, use ```docker buildx``` command:

```bash
cd symphony-api
go mod vendor    
docker buildx build --no-cache --platform linux/amd64,linux/arm64,linux/arm/v7 -t <API image tag> --push .
# or to build for single platform
docker build -t <API image tag> .
```
> **NOTE:** if you receive an error message saying "multiple platforms featue is not currently not supported...", use ```docker buildx create --use``` to enable multiplatform builds

Some tips:

* Change the ```--platform``` switch to build for different platforms. For example, if you want to just build for amd64, use ```--platform linux/amd64``` instead.
* Remove the ```--push``` switch if you want to push the container image later
* Remove the ```--no-cache``` switch to leverage existing cache, which may speed up builds.

If you just want to build the Symphony API binary, use:
```bash
go mod vendor
go build -o symphony-api
```
Then, you can launch the API as a local web server using (default port is ```8080```. See ```./symphony-api.json``` settings):
```bash
./symphony-api -c ./symphony-api.json
```
> **NOTE**: To use Kubernetes reference provider outside of a Kubernetes cluster, you need to change the reference's ```inCluster``` setting to ```false``` (see [Reference Provider](../providers/reference_provider.md)).


You can override the default logging level with a ```LOG_LEVEL``` environment variable. For example, to launch Symphony API with ```Info``` log level:
```bash
# running as process
export LOG_LEVEL=Info
./symphony-api -c ./symphony-api.json
# or, running as container in console model
docker run --rm -it -e LOG_LEVEL=Info possprod.azurecr.io/symphony-api:0.41.32 
```

When running Symphony API as a container, you can use a ```CONFIG``` environment variable to override config file location:
```
docker run --rm -it -v /path/to/my-config.json:/configs -e CONFIG=/configs/my-config.json possprod.azurecr.io/symphony-api:0.41.32
```

## 3. Build Symphony K8s binding container
To build Symphony K8s binding container, use the following commands:
```bash
cd symphony-k8s
go mod vendor
make generate
make build
make docker-build IMG=<Symphony-k8s image tag>
docker push <Symphony-k8s image tag>
```

## 4. Push Symphony containers to Azure container registry (optional)

```bash
az login
TOKEN=$(az acr login --name possprod --expose-token --output tsv --query accessToken)
docker login possprod.azurecr.io --username 00000000-0000-0000-0000-000000000000 --password $TOKEN
docker tag <Symphony-k8s image tag> possprod.azurecr.io/symphony-k8s:latest
docker push possprod.azurecr.io/symphony-k8s:latest
```

## 5. Update Helm chart (optional)
```bash
cd symphony-k8s
cd config/manager
kustomize edit set image controller=possprod.azurecr.io/symphony-k8s:0.38.0
cd ../..
kustomize build ./config/default/ -o ./helm/symphony/templates/symphony.yaml
```
## 5. Package and push Helm chart (optional)
```bash
cd symphony-k8s/helm
# update helm version
# 1) Edit Chart.yaml and update both version and appVersion to desired version number, like 0.1.22
# 2) Edit values.yaml and update both tags to the desired version number, like 0.1.22
# package
helm package symphony
# log in to registry
export HELM_EXPERIMENTAL_OCI=1
USER_NAME="00000000-0000-0000-0000-000000000000"
PASSWORD=$(az acr login --name possprod --expose-token --output tsv --query accessToken)
helm registry login possprod.azurecr.io   --username $USER_NAME --password $PASSWORD
# push image
helm push symphony-0.1.22.tgz oci://possprod.azurecr.io/helm
```
## 6. Build Symphony Agent container (optional, if you plan to use Symphony Agent as a container)
```bash
cd symphony-api
go mod vendor    
docker buildx build --no-cache --platform linux/amd64,linux/arm64,linux/arm/v7 -t <Agent image tag> --file ./Dockerfile.agent --push .
# or to build for single platform
docker build -t <Agent image tag> -f ./Dockerfile.agent .
```
To run an agent locally, use Docker:
```
docker run -p 8088:8088 -e SYMPHONY_URL=http://localhost:8080/v1alpha2/agent/references hbai/symphony-agent:0.1.22
```
> **NOTE:** See [Agent](../agent/agent.md) doc for more details.
## 7. Build Symphony Agent binary (optional, if you plan to use Symphony Agent as a binary)
```bash
cd symphony-api
go mod vendor
go build -o ./symphony-agent #./symphony-agent.exe for Windows
```
To run agent binary, use a sample ```symphony-agent.json``` file under the ```symphony-api``` folder:
```bash
./symphony-agent -c ./symphony-agent.json -l Debug
```
> **NOTE:** See [Agent](../agent/agent.md) doc for more details.