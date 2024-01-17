# Build Symphony containers

_(last edit: 4/12/2023)_

Symphony has two parts: the platform-agnostic API (symphony-api) and Kubernetes binding (symphony-k8s), both are packaged as Docker containers.

## 0.Prerequisites

* [Go](https://golang.org/dl/) (1.19 or higher, latest stable version is recommended)
* [Git](https://git-scm.com/downloads)
* [Docker](https://www.docker.com/products/docker-desktop)

  > **NOTE:** Some tools we use work better with access to docker commands without sudo. Use Docker Desktop version when possible. Otherwise, you need to add your user to docker group (see [instructions](https://www.docker.com/products/docker-desktop)). Please don't use rootless model, which isn't supported by some of the tools.

* A Kubernetes cluster, such as [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)

  >**NOTE:** Symphony should work with any Kubernetes clusters on x86 and x64 CPUs. symphony-api container is compiled for ARM as well but we haven't tested other containers.

* [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/binaries/)

  ```bash
  curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
  sudo mv kustomize /usr/local/bin/
  ```

* [Kubebuilder](https://book.kubebuilder.io/)

  ```bash
  # install kubebuiler
  curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
  chmod +x kubebuilder && mv kubebuilder /usr/local/bin/
  # create a temporary project to get controller-gen
  cd
  mkdir temp
  cd temp
  kubebuilder init --domain my.domain --repo my.domain/guestbook
  make test
  cp bin/controller-gen $GOPATH/bin/
  cp bin/setup-envtest $GOPATH/bin/
  cd ..
  rm -rf temp
  ```

* make and gcc

  ```bash
  sudo apt-get update && sudo apt-get upgrade -y
  sudo apt-get install make gcc -y
  ```

* An Azure subscription
* Visual Studio Code with the following extensions installed:

  * Go extensions (open any Go source file and select **Install All** tools)
  * [Azure IoT Hub](https://marketplace.visualstudio.com/items?itemName=vsciot-vscode.azure-iot-toolkit) extension

* Enough memory (>16G), especially if you run Kind cluster on the same machine.
* Kustomize (optional; for building Helm chart)
* Helm (optional; for building Helm chart)

## 1. Clone the repository

```bash
git clone https://github.com/Azure/symphony.git
```

## 2. Build and run Symphony binary for local dev/test

To build Symphony API binary, use `go build`:

```bash
cd api
go build -o symphony-api
# to build for Windows
GOOS=windows GOARCH=amd64 go build -o symphony-api.exe
# to build for Mac
GOOS=darwin GOARCH=amd64 go build -o symphony-api-mac
```

Then, you can launch the API as a local web service using (default port is `8082`. See `./symphony-api-no-k8s.json` settings):

```bash
./symphony-api -c ./symphony-api-no-k8s.json
```

## 2. Build and run Symphony API container for local dev/test

To build Symphony API container, use `docker build`:

```bash
# if you build from the api folder
docker build -t <API image tag> .. -f Dockerfile
```

Or,

```bash
# if you build from the repo root folder
docker build -t <API image tag> . -f api/Dockerfile
```

To build multi-platform Symphony API container, use `docker buildx` command:

```bash
cd api
docker buildx build --no-cache --platform linux/amd64,linux/arm64,linux/arm/v7 -t <API image tag> --push .. -f Dockerfile
```

> **NOTE:** if you receive an error message saying "multiple platforms feature is not currently not supported...", use `docker buildx create --use` to enable multi-platform builds.

Some tips:

* Change the `--platform` switch to build for different platforms. For example, if you want to just build for amd64, use `--platform linux/amd64` instead.
* Remove the `--push` switch if you want to push the container image later
* Remove the `--no-cache` switch to leverage existing cache, which may speed up builds.

To run Symphony API as a Docker container, use `docker run` with the `symphony-api-no-k8s.json` configuration file:

```bash
docker run --rm -it -v /path/to/my-config.json:/configs -e CONFIG=/configs/symphony-api-no-k8s.json <API Image tag>
```

For example, while under the `api` folder, you can launch latest Symphony API container with `docker run`:

```bash
docker run --rm -it -v ./api:/configs -e CONFIG=/configs/symphony-api-no-k8s.json ghcr.io/azure/symphony/symphony-api:latest
```

You can override the default logging level with a `LOG_LEVEL` environment variable. For example, to launch Symphony API with `Info` log level:

```bash
# running as process
export LOG_LEVEL=Info
./symphony-api -c ./symphony-api-no-k8s.json
# or, you can directly set the log level switch
./symphony-api -c ./symphony-api-no-k8s.json -l Info
# or, running as container in console model
docker run --rm -it -e LOG_LEVEL=Info -v ./api:/configs -e CONFIG=/configs/symphony-api-no-k8s.json ghcr.io/azure/symphony/symphony-api:latest
```

## 3. Build Symphony K8s binding container

To build Symphony K8s binding container, use the following commands:

```bash
cd k8s
make generate
make build
make docker-build IMG=<Symphony-k8s image tag>
docker push <Symphony-k8s image tag>
```

## 4. Push Symphony containers to GitHub container registry (optional)

```bash
# GitHub Container Registry
TOKEN='{YOUR_GITHUB_PAT_TOKEN}'
docker login ghcr.io --username USERNAME --password $TOKEN
docker tag <Symphony-k8s image tag> ghcr.io/azure/symphony/symphony-k8s:latest
docker push ghcr.io/azure/symphony/symphony-k8s:latest
```

## 5. Update Helm chart (optional)

```bash
cd k8s
mage helmTemplate
# Generated startup yaml will be updated in ../packages/helm/symphony/templates/symphony.yaml.
```

> **IMPORTANT**: With current Kustomize, empty `creationTimestamp` properties are inserted into the generated artifacts somehow, causing Helm chart to fail. You'll need to manually remove all occurrence of `creationTimestamp` properties with `null` or `"null"` from the artifacts, until a proper solution is found.

## 5. Package and push Helm chart (optional)

```bash
cd packages/helm
# update helm version
# 1) Edit Chart.yaml and update both version and appVersion to desired version number, like 0.43.1
# 2) Edit values.yaml and update both tags to the desired version number, like 0.43.1
# package
helm package symphony
# log in to registry
export HELM_EXPERIMENTAL_OCI=1
USER_NAME="USERNAME"
TOKEN='{YOUR_GITHUB_PAT_TOKEN}'
helm registry login ghcr.io --username $USER_NAME --password $PASSWORD
# push image
helm push symphony-0.43.1.tgz oci://ghcr.io/azure/symphony/helm
```

## 6. Build Symphony agent container (optional)

If you plan to use Symphony Agent as a container, run the following commands to build the container.

```bash
cd api
docker buildx build --no-cache --platform linux/amd64,linux/arm64,linux/arm/v7 -t <Agent image tag> --file ./Dockerfile.agent --push .
# or to build for single platform
docker build -t <Agent image tag> -f ./Dockerfile.agent .
```

To run an agent locally, use Docker:

```bash
docker run -p 8088:8088 -e SYMPHONY_URL=http://localhost:8080/v1alpha2/agent/references hbai/symphony-agent:0.1.22
```

For more information, see [Symphony agent](../agent/_overview.md).

## 7. Build Symphony Agent binary (optional)

If you plan to use Symphony Agent as a binary, run the following commands to build the binary.

```bash
cd api
go build -o ./symphony-agent #./symphony-agent.exe for Windows
```

To run agent binary, use a sample `symphony-agent.json` file under the `symphony-api` folder:

```bash
./symphony-agent -c ./symphony-agent.json -l Debug
```

For more information, see [Symphony agent](../agent/_overview.md).

## 8. Build Symphony CLI (maestro)

Follow the steps to [Build Maestro CLI](../cli/build_cli.md).

## Next steps

Now that you have the Symphony containers running and the CLI tool built, follow the steps to [prepare Azure resources](./prepare_azure.md) in your Azure environment for running test scenarios.
