# Symphony

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Symphony is a powerful service orchestration engine that enables the organization of multiple intelligent edge services into a seamless, end-to-end experience. Its primary purpose is to address the inherent complexity of edge deployment by providing a set of technology-agnostic workflow APIs, which are designed to deliver a streamlined experience for users across all device profiles.

Symphony is uniquely capable of providing consistency across the entire software stack, from drivers to containers to configurations and policies. This comprehensive approach ensures that all aspects of your intelligent edge projects are effectively managed and optimized. Moreover, Symphony provides full support for the entire lifecycle of your edge computing initiatives, spanning from the initial deployment to ongoing updates and maintenance.

With Symphony, users can benefit from a powerful and versatile platform that streamlines edge service orchestration and management, while also ensuring seamless integration with existing technology stacks. Whether you are a small business owner or a large enterprise, Symphony is an ideal solution for enhancing the efficiency and effectiveness of your edge computing initiatives.

## Symphony Characteristics

* Standard-based

    Symphony is a versatile and standards-based solution that delivers exceptional flexibility and extensibility. It natively runs on Kubernetes, which means users can leverage all existing Kubernetes tooling to interact with Symphony. Moreover, Symphony supports a wide range of popular industrial standards, protocols, and frameworks, including [OpenTelemetry](https://opentelemetry.io/), [Distributed Application Runtime (Dapr)](https://dapr.io/), [Message Queuing Telemetry Transport (MQTT)](https://mqtt.org/), [Open Neural Network Exchange (ONNX)](https://onnx.ai/), [Akri](https://github.com/project-akri/akri), [kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/), [Helm](https://helm.sh/), and many others. This broad range of support makes Symphony an ideal solution for organizations seeking to build and deploy edge services that meet their specific needs.

* Meet customers where they are

    Another key advantage of Symphony is its extensibility. It supports the integration of first-party and third-party services, and all Symphony capabilities, including device registry, device updates, and solution deployment, can be replaced with custom implementations. This means that Symphony can be tailored to meet the specific needs of any organization, regardless of their size or complexity.

* Zero-frction adoption

    Symphony's zero-friction adoption approach is another key feature that sets it apart from other solutions. Users can get started with Symphony using a single computer, and there is no need for special hardware, an Azure subscription, or Kubernetes to start experimenting with the solution. Additionally, the same Symphony artifacts used during testing and development can be carried over into production deployments, ensuring a smooth transition and reducing overall deployment time and costs.

- **Azure powered and platform agnostic**

    Symphony is powered by Azure and is platform-agnostic, making it an ideal solution for organizations that already use Azure Edge and AI services like [Azure IoT Hub](https://docs.microsoft.com/azure/iot-hub/), [Azure IoT Edge](https://azure.microsoft.com/services/iot-edge/), [Azure Cognitive Services](https://azure.microsoft.com/services/cognitive-services/), [Azure Storage](https://azure.microsoft.com/products/category/storage/), [Azure ML](https://azure.microsoft.com/services/machine-learning/), Mon[Azure Monitor](https://docs.microsoft.com/azure/azure-monitor/)itor, and [Azure Arc](https://learn.microsoft.com/azure/azure-arc/overview). However, Symphony is also fully compatible with other non-Azure services or open-source software tools, allowing organizations to modify the solution to meet their specific needs. This flexibility ensures that Symphony meets customers where they are, making it an ideal solution for organizations of all sizes and complexities.

## Getting Started
There are several ways to get started with Symphony, including using the CLI tool, Helm, Docker, or the symphony-api binary.

### Using Symphony CLI
The easiest way to get started with Symphony is by using Symphony's CLI tool, called maestro. The CLI tool can be installed on **Linux**, **WSL**, and **Mac** using the following command:

```Bash
wget -q https://raw.githubusercontent.com/Haishi2016/Vault818/master/cli/install/install.sh -O - | /bin/bash
```
For **Windows**, the following PowerShell command can be used:
```PowerShell
powershell -Command "iwr -useb https://raw.githubusercontent.com/Haishi2016/Vault818/master/cli/install/install.ps1 | iex"
```
After Symphony is installed, you can use `maestro` to try out sample scenarios.

```bash
maestro up
```
You can also install Symphony using Helm by running the following command:
### Using Helm
After Symphony is installed, you can use maestro to try out sample scenarios.
```Bash
helm install symphony oci://symphonyk8s.azurecr.io/helm/symphony --version 0.41.42
```
### Using Docker
You can also install Symphony using Helm by running the following command:
```Bash
docker run -d --name symphony-api -p 8080:8080 possprod.azurecr.io/symphony-api:0.41.42 
```
### Using symphony-api binary
You can also run Symphony in standalone mode as a single process by running the following command:
```Bash
./symphony-api -c ./symphony-api-dev.json -l Debug
```
## What's Next

* [Quickstart scenarios](./docs/symphony-book/quick_start/quick_start.md)
* [Symphony Docs](./docs/README.md)

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft 
trademarks or logos is subject to and must follow 
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
