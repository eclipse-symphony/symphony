# Symphony

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![build](https://github.com/eclipse-symphony/symphony/actions/workflows/go.yml/badge.svg)

_(last edit: 02/02/2024)_

Symphony is a powerful service orchestration engine that enables the organization of multiple intelligent edge services into a seamless, end-to-end experience. Its primary purpose is to address the inherent complexity of edge deployment by providing a set of technology-agnostic workflow APIs, which are designed to deliver a streamlined experience for users across all device profiles.

Symphony is uniquely capable of providing consistency across the entire software stack, from drivers to containers to configurations and policies. This comprehensive approach ensures that all aspects of your intelligent edge projects are effectively managed and optimized. Moreover, Symphony provides full support for the entire lifecycle of your edge computing initiatives, spanning from the initial deployment to ongoing updates and maintenance.

With Symphony, users can benefit from a powerful and versatile platform that streamlines edge service orchestration and management, while also ensuring seamless integration with existing technology stacks. Whether you are a small business owner or a large enterprise, Symphony is an ideal solution for enhancing the efficiency and effectiveness of your edge computing initiatives.

## Symphony Characteristics

* Standard-based

    Symphony is a versatile and standards-based solution that delivers exceptional flexibility and extensibility. It natively runs on Kubernetes, which means users can leverage all existing Kubernetes tooling to interact with Symphony. Moreover, Symphony supports a wide range of popular industrial standards, protocols, and frameworks, including [OpenTelemetry](https://opentelemetry.io/), [Distributed Application Runtime (Dapr)](https://dapr.io/), [Message Queuing Telemetry Transport (MQTT)](https://mqtt.org/), [Open Neural Network Exchange (ONNX)](https://onnx.ai/), [Akri](https://github.com/project-akri/akri), [kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/), [Helm](https://helm.sh/), and many others. This broad range of support makes Symphony an ideal solution for organizations seeking to build and deploy edge services that meet their specific needs.

    Symphony also supports running in a standalone mode independent from Kubernetes. All you need is a single Symphony binary and nothing else!

* Meet customers where they are

    Another key advantage of Symphony is its extensibility. It supports the integration of first-party and third-party services, and all Symphony capabilities, including device registry, device updates, and solution deployment, can be replaced with custom implementations. This means that Symphony can be tailored to meet the specific needs of any organization, regardless of their size or complexity.

* Zero-friction adoption

    Symphony's zero-friction adoption approach is another key feature that sets it apart from other solutions. Users can get started with Symphony using a single computer, and there is no need for special hardware, an Azure subscription, or Kubernetes to start experimenting with the solution. Additionally, the same Symphony artifacts used during testing and development can be carried over into production deployments, ensuring a smooth transition and reducing overall deployment time and costs.

- **Symphony is platform agnostic**

    Symphony was started by Microsoft as a platform-agnostic project, making it an ideal solution for organizations that already use Azure Edge and AI services like [Azure IoT Hub](https://docs.microsoft.com/azure/iot-hub/), [Azure IoT Edge](https://azure.microsoft.com/services/iot-edge/), [Azure Cognitive Services](https://azure.microsoft.com/services/cognitive-services/), [Azure Storage](https://azure.microsoft.com/products/category/storage/), [Azure ML](https://azure.microsoft.com/services/machine-learning/), [Azure Monitor](https://docs.microsoft.com/azure/azure-monitor/), and [Azure Arc](https://learn.microsoft.com/azure/azure-arc/overview). However, Symphony is also fully compatible with other non-Azure services or open-source software tools, allowing organizations to modify the solution to meet their specific needs. This flexibility ensures that Symphony meets customers where they are, making it an ideal solution for organizations of all sizes and complexities.

## Getting Started
There are several ways to get started with Symphony, including using the CLI tool, Helm, Docker, or the symphony-api binary.

### Using Symphony CLI

> **NOTE**: The following GitHub URL is a temporary parking location and is sugject to change.

The easiest way to get started with Symphony is by using Symphony's CLI tool, called maestro. The CLI tool can be installed on **Linux**, **WSL**, and **Mac** using the following command:

```Bash
wget -q https://raw.githubusercontent.com/eclipse-symphony/symphony/master/cli/install/install.sh -O - | /bin/bash
```
For **Windows**, the following PowerShell command can be used:
```PowerShell
powershell -Command "iwr -useb https://raw.githubusercontent.com/eclipse-symphony/symphony/master/cli/install/install.ps1 | iex"
```
After Symphony is installed, you can use `maestro` to try out sample scenarios.

```bash
maestro up
```

### Using Helm
You can also install Symphony using Helm by running the following command:
```Bash
helm install symphony oci://ghcr.io/eclipse-symphony/helm/symphony --version '0.48.28'
```
After Symphony is installed, you can use maestro to try out sample scenarios.

### Using Docker
You can also install Symphony using Docker with the bundled `symphony-api.json` or volume mounting your own & injecting its reference via `CONFIG` env:
```Bash
docker run -d --name symphony-api -p 8080:8080 -e CONFIG=/symphony-api.json ghcr.io/eclipse-symphony/symphony-api:0.48.28
```
### Using symphony-api binary
You can also run Symphony in standalone mode as a single process by running the following command:
```Bash
./symphony-api -c ./symphony-api-dev.json -l Debug
```
## Provider Conformance Test Results
Symphony is an extensible system with the concept of providers. For each provider types, we define one or multiple conformance test suites that ensure provider implementations behaves consistently and predictably.

### Target Providers

| Provider | Basic<sup>1</sup> | 
|--------|--------|
| ```providers.target.adb``` |![](https://byob.yarr.is/Haishi2016/badges/target-adb-app)|
| ```providers.target.azure.adu``` |![](https://byob.yarr.is/Haishi2016/badges/target-adu-app)|
| ```providers.target.azure.iotedge``` |![](https://byob.yarr.is/Haishi2016/badges/target-iotedge-app)|
| ```providers.target.docker```|![](https://byob.yarr.is/Haishi2016/badges/target-docker-app)|
| ```providers.target.helm```|![](https://byob.yarr.is/Haishi2016/badges/target-helm-app)|
| ```providers.target.http```|![](https://byob.yarr.is/Haishi2016/badges/target-http-app)|
| ```providers.target.k8s``` |![](https://byob.yarr.is/Haishi2016/badges/target-k8s-app)|
| ```providers.target.kubectl```|![](https://byob.yarr.is/Haishi2016/badges/target-kubectl-app)|
| ```providers.target.mqtt```|![](https://byob.yarr.is/Haishi2016/badges/target-mqtt-app)|
| ```providers.target.proxy```|![](https://byob.yarr.is/Haishi2016/badges/target-proxy-app)|
| ```providers.target.script```|![](https://byob.yarr.is/Haishi2016/badges/target-script-app)|
| ```providers.target.staging```|![](https://byob.yarr.is/Haishi2016/badges/target-staging-app)|
| ```providers.target.win10```|![](https://byob.yarr.is/Haishi2016/badges/target-win10-app)|

1. **Basic** conformance level requires a provider to properly respond to missing properties

## What's Next

* [The Symphony Book](./docs/README.md)
* [Set up a local environment](./test/localenv/README.md)

## Community

### Communication and Discord

All your contributions and suggestions are greatly appreciated! One of the easiest ways to contribute is to participate in Discord discussions, report issues, or join the monthly community calls.

### Questions and issues

Reach out with any questions you may have and we'll make sure to answer them as soon as possible. Community members, please feel free to jump in to join discussions or answer questions!

| Platform  | Link        |
|:----------|:------------|
| Discord | Join the [Discord server](https://discord.gg/JvY8qBkWbw)

### Email announcements

Want to stay up to date with Symphony releases, community calls, and other announcements? Join the Google Group to stay up to date on the latest Symphony news.

| Group | Link |
|:------|:-----|
| symphonyoss | Join the [symphonyoss Group](https://groups.google.com/g/symphonyoss)

### Community meetings

Every month we host a community call to showcase new features, review upcoming milestones, and engage in a Q&A. For community calls, anyone from the Symphony community can participate or present a topic. All are welcome!


You can always catch up offline by watching the recordings below.

| Asset | Link        |
|:-----------|:------------|
| Meeting Link | [Teams Link](https://teams.microsoft.com/meet/267721771421?p=hyAHXrqsyVdDA0VAZQ)
| Meeting Recordings | [YouTube](https://www.youtube.com/@Eclipse-Symphony/videos)

### Upcoming calls

| Date & time |
|-------------|
| Wednesday March 26 <sup>th</sup>, 2025 8:00am Pacific Time (PST) |

### Previous calls

| Date & time | Link |
|-------------|:-------------|
| 12/11/2024 | [Recording Link](https://www.youtube.com/watch?v=0WEDia5JD-Y)|
| 01/15/2025 | [Recording Link](https://youtu.be/8b4wc21eOjM)|
| 02/26/2025 | [Recording Link](https://youtu.be/VAwGlObx0mQ)|


## Contributing

This project welcomes contributions and suggestions.  

### Eclipse Contributor Agreement

Before your contribution can be accepted by the project team contributors must electronically sign the Eclipse Contributor Agreement (ECA).

http://www.eclipse.org/legal/ECA.php
Commits that are provided by non-committers must have a Signed-off-by field in the footer indicating that the author is aware of the terms by which the contribution has been provided to the project. The non-committer must additionally have an Eclipse Foundation account and must have a signed Eclipse Contributor Agreement (ECA) on file.

For more information, please see the Eclipse Committer Handbook: https://www.eclipse.org/projects/handbook/#resources-commit

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft 
trademarks or logos is subject to and must follow 
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
