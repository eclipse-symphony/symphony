# Symphony Multi-target Control Plane
## Table of Content

### Part I: Fundamentals

* [**Chapter 0**: Quick start](symphony-book/quick_start/quick_start.md)
    * [Deploy to K8s](symphony-book/quick_start/deploy_prometheus_k8s.md)
    * [Deploy to IoT Edge](symphony-book/quick_start/deploy_solution_to_azure_iot_edge.md)
* [**Chapter 1**: Introduction](symphony-book/introduction/introduction.md)
* [**Chapter 2**: Build & Deployment](symphony-book/build_deployment/build_deployment.md)
    * [Build Symphony](symphony-book/build_deployment/build.md)
    * [Deploy Symphony](symphony-book/build_deployment/deploy.md)
    * [Deploy Symphony Behind API Management](symphony-book/build_deployment/apim.md)
* [**Chapter 3**: Symphony Object Model](symphony-book/uom/uom.md)
    * [Device](symphony-book/uom/device.md)    
    * [Instance](symphony-book/uom/instance.md)
    * [Model](symphony-book/uom/ai-model.md)
    * [Skill](symphony-book/uom/ai-skill.md)    
    * Skill Node
    * Skill Package
    * [Solution](symphony-book/uom/solution.md)    
    * [Target](symphony-book/uom/target.md)
* [**Chapter 4**: Target Management](symphony-book/target-management/target-management.md)    
* [**Chapter 5**: Device Management](symphony-book/device-management/device-management.md)
* [**Chapter 6**: Solution Management](symphony-book/solution-management/solution-management.md)
* [**Chapter 7**: Instance Management](symphony-book/instance-management/instance-management.md)
* **Chapter 8**: AI Model Management
* [**Chapter 9**: AI Skill Management](symphony-book/skill-management/skill-management.md)
* **Chapter 10**: Policy Management
    * GateKeeper
    * Kyverno
    * Azure Policy
* **Chapter 11**: Security
    * [Authentication](symphony-book/security/authentication.md)
    * [Authorization](symphony-book/security/authorization.md)
    * [CORS Control](symphony-book/bindings/cors.md)
* [**Chapter 12**: Providers](symphony-book/providers/overview.md)
    * [Azure IoT Edge Provider](symphony-book/providers/iot_provider.md)
    * [Helm Provider](symphony-book/providers/helm_provider.md)
    * [HTTP Provider](symphony-book/providers/http_provider.md)    
    * [Kubernetes Provider](symphony-book/providers/k8s_provider.md)
    * [HTTP Proxy Provider](symphony-book/providers/http_proxy_provider.md)
    * [MQTT Proxy Provider](symphony-book/providers/mqtt_proxy_provider.md)
    * Radius Provider
    * [Script Provider](symphony-book/providers/script_provider.md)
    * [Staging Provider](symphony-book/providers/staging_provider.md)
    * Terraform Provider
    * [Provider Interface](symphony-book/providers/provider_interface.md)
    * [Write Custom Providers](symphony-book/providers/python_provider.md)
* [**Chapter 13**: Managers](symphony-book/managers/overview.md)
* [**Chapter 14**: Vendors](symphony-book/vendors/overview.md)
* [**Chapter 15**: Hosts](symphony-book/hosts/overview.md)
* [**Chapter 16**: Bindings](symphony-book/bindings/overview.md)
    * [HTTP Binding](symphony-book/bindings/http-binding.md)
        * [Application Insight middleware](symphony-book/bindings/app-insight.md)
        * [CORS middleware](symphony-book/bindings/cors.md)
        * [Distributed Tracing middleware](symphony-book/bindings/tracing.md)
        * [JWT Token middleware](symphony-book/bindings/jwt-handler.md)
    * [MQTT Binding](symphony-book/bindings/mqtt-binding.md)
* **Chapter 17**: Capabilities and Capability Matching
* [**Chapter 18**: Agent](symphony-book/agent/agent.md)
    * Symphony Agent
    * Polling Agent
* [**Chapter 19**: Integrations](symphony-book/integrations/overview.md)
    * [Akri](symphony-book/akri/akri.md)
    * AKS
    * AKS Edge Essentials
    * AKS Fleet Management
    * Arc
    * Azure API Management
    * Azure Logic Apps
    * Azure Resource Manager
    
* [**Chapter 20**: Deployment Scenarios](symphony-book/scenarios/development-scenarios.md)    
    * [Adaptive deployments](symphony-book/scenarios/adaptive-deployment.md) - _deploys the same artifact to multiple target types_
    * [Canary deployments](symphony-book/instance-management/instance-management.md#canary-deployment) - _adjustable canary deployments_
    * Cascaded deployments - _deploys along cluster hierarchies_
    * Dynamic deployments - _dynamically reconfiguration based on discovered devices_
    * Fan-out deployments - _deploys to matching targets_
    * [Gated deployments](symphony-book/scenarios/gated-deployment.md) - _adds control points to your deployments_
    * [Human approval](symphony-book/scenarios/human-approval.md) - _requires manual approval_
    * Hybrid deployments - _deploys to both edge and cloud_
    * [Scheduled deployments](symphony-book/instance-management/instance-management.md#scheduled-deployment) - _deploys at given schedule_
    * [Split deployments](symphony-book/scenarios/linux-with-uwp-frontend.md) - _distributes components to matching targets (such as Windows frontend and Linux backend)_
    * Staged deployments - _deploys app across rings_
* [**Chapter 21**: CLI](symphony-book/cli/cli.md)
  * [Usage](symphony-book/cli/cli.md)
  * [Build](symphony-book/cli/build_cli.md)
* [**Chapter 22**: REST API](symphony-book/api/api.md)
    * [Instances API](symphony-book/api/instances-api.md)
    * [Solutions API](symphony-book/api/solutions-api.md)
    * [Targets API](symphony-book/api/targets-api.md)
    * [Query Projection](symphony-book/api/projection.md)
* **Chapter 23**: SDKS
    * Rust SDK
    * Python SDK
* **Chapter 24**: Observability
    * Events
    * [Distributed Tracing](symphony-book/observability/distributed-tracing.md)
    * Logs
    * Metrics
* **Chapter 25**: Federation
    * Component Name Resolutions
    * Federated Queries
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
