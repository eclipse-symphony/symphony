# Local environment

The local environment is a minikube cluster that deploys symphony for testing purposes.

See all commands with `mage -l`

```
> mage -l

Use this tool to quickly get started developing in the symphony ecosystem. The tool provides a set of common commands to make development easier for the team. To get started using Minikube, run 'mage build minikube:start minikube:load deploy'.

Targets:
  acrLogin              Log into the ACR, prompt if az creds are expired
  build:all             Build builds all containers
  build:api             Build api container
  build:k8s             Build k8s container
  buildUp               builds the latest images and starts the local environment
  cluster:deploy        Deploys the symphony ecosystem to your local Minikube cluster.
  cluster:down          Stop the cluster
  cluster:load          Brings the cluster up with all images loaded
  cluster:status        Show the state of the cluster for CI scenarios
  cluster:up            Brings the cluster up and deploys
  destroy               Uninstall all components
  minikube:dashboard    Launch the Minikube Kubernetes dashboard.
  minikube:delete       Deletes the Minikube cluster from you dev box.
  minikube:install      Installs the Minikube binary on your machine.
  minikube:load         Loads symphony component docker images onto the Minikube VM.
  minikube:start        Starts the Minikube cluster w/ select addons.
  minikube:stop         Stops the Minikube cluster.
  pull:all              Pulls all docker images for symphony
  pull:api              Pull symphony-api
  pull:k8s              Pull symphony-k8s
  pullUp                pulls the latest images and starts the local environment
  test:up               Deploys the symphony ecosystem to minikube and waits for all pods to be ready.
  up                    brings the minikube cluster up with symphony deployed
```


# Getting started

Use the `Up` commands to start the local environment and deploy symphony. If it is your first time running the environment and you do not have local images you will need to either pull or build them.

```bash
# Start the cluster and deploy symphony using the images on your dev box
mage Up

# Pull images from ACR and start the cluster
mage PullUp

# Build images from source and start the cluster
mage BuildUp

# View the pods running in the cluster
mage cluster:status
```

For working the cluster use [k9s](https://github.com/derailed/k9s) or `kubectl`
# Local development

For a typical development workflow you can build the image or images you are modifying, then deploy them to the cluster before testing and applying custom resources.

```bash
# Build the image you are working on or use build:all
mage build:k8s

# Deploy to the cluster
mage up
```

You can also run the deployment manually


```bash
# build first
mage build:all

# run the cluster in the background
mage cluster:up

# deploy symphony
mage cluster:deploy
```

To remove symphony from the cluster use

```
mage Destroy all,nowait
```

# Troubleshooting

If you are seeing strange behavior or getting errors the first thing to try is completely deleting minikube and starting over with a fresh cluster. Many commands will recreate minikube for you automatically, but it is worth checking that minikube is actually getting cleaned up.

```bash
minikube delete
```


# Integration tests

CI integration tests can be run locally with the following command:

```
mage test:up
```

See [integration test README](../test/integration/README.md) for more details.