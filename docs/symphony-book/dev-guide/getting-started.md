# Getting Started with Symphony Development

Welcome to Symphony development! This document walks you through steps to configure your dev environment, get oriented in the code base, run some test cases, and get ready to make some contributions!

## Setting up your dev environment
You can develop Symphony on any PCs, Macs, or Linux systems. A common choice is an Ubuntu 20.04 WSL on Windows 11, for instance. Symphony is written in [Go](https://go.dev/) language. Symphony components are packaged as a few [Docker](https://www.docker.com/products/docker-desktop) containers. And the whole Symphony system is packaged as a Helm chart. Symphony’s K8s integration uses [kubebuilder](https://book.kubebuilder.io/). 

Please see the [build instruction doc](../build_deployment/build.md) for details on setting up your environment and building Symphony components.

## Understanding repo structure
Symphony repo consists a few top-level folders, as summarized below:

| Folder | Content |
|--------|--------|
| ```api``` | Symphony API source code |
| ```cli``` | Symphony CLI (maestro) source code |
| ```coa``` | A microservice framework (HB-MVP) Symphony uses. Some HB-MVP introduction articles: [part 1](https://www.linkedin.com/pulse/hb-mvp-design-pattern-extensible-systems-part-i-haishi-bai/), [part 2](https://www.linkedin.com/pulse/hb-mvp-design-pattern-extensible-systems-part-ii-haishi-bai/), [part 3](https://www.linkedin.com/pulse/hb-mvp-design-pattern-extensible-systems-part-iii-haishi-bai/)|
| ```docs``` | Symphony docs and samples |
| ```k8s``` | Symphony K8s control plane |
| ```sdks```| Symphony SDKs. We currently have a Python SDK PoC |

## Debugging Symphony locally
Please see [this document](./debugging-api.md) for more details on how to run and debug Symphony locally.

## Braches and Forks
You are encouraged to contribute directly to Symphony. Create your own feature branch and create PR to get your contributions reviewed and merged. If you’ve decided to fork Symphony for your specific project, we highly recommend you contribute to upstream as much as possible. 

Symphony repo is lack of automatic CI/CD pipelines, gated check-ins and automated workflows. Please make contributions to make the repo more automated.

## Wrting a provider
A common task of extending Symphony is to write/modify a [provider](../providers/overview.md), especially a [target provider](../providers/target_provider.md). 

A target provider implements the [Target Provider Interface](../providers/provider_interface.md).

To create a new provider:

1. Create a new folder under the ```api/pkg/apis/v1alpha1/providers/target``` folder (such as ```myprovider```).
2. Create two files under the above folder. One file contains the provider implemenation (such as ```myprovider.go```). And the other file contains unit tests for the provider (```myprovider_test.go```).
    > **NOTE**: An easy way to get started with a provider is to copy an existing provider implementation and make modifications.
3. Implement the target provider interface in your provider source code. Generally, a provider defines an associated configuration type, which will be injected to the provider instance during initialization. 
4. Implement relevant unit test cases.



