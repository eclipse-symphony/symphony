# The Symphony Book

_(last edit: 4/25/2023)_

## Table of Content

### Part I: Fundamentals

* [**Chapter 0**: Quick start](symphony-book/quick_start/quick_start.md)
    * [Deploy apps to K8s](symphony-book/quick_start/deploy_prometheus_k8s.md)
    * [Deploy apps locally in standalone mode](symphony-book/quick_start/deploy_redis_no_k8s.md)
    * [Deploy apps to IoT Edge](symphony-book/quick_start/deploy_solution_to_azure_iot_edge.md)
* [**Chapter 1**: Introduction](symphony-book/introduction/introduction.md)
* [**Chapter 2**: Build & Deployment](symphony-book/build_deployment/build_deployment.md)
    * [Build Symphony](symphony-book/build_deployment/build.md)
    * [Deploy Symphony](symphony-book/build_deployment/deploy.md)
    * [Running Symphony in standalone mode](symphony-book/build_deployment/standalone.md)
    * [Deploy Symphony Behind API Management](symphony-book/build_deployment/apim.md)
* [**Chapter 3**: Symphony Object Model](symphony-book/uom/uom.md)
    * Compaign
    * [Device](symphony-book/uom/device.md)    
    * [Instance](symphony-book/uom/instance.md)
    * [Model](symphony-book/uom/ai-model.md)
    * [Skill](symphony-book/uom/ai-skill.md)    
    * Skill Node
    * Skill Package
    * [Solution](symphony-book/uom/solution.md)    
    * [Target](symphony-book/uom/target.md)
    * [Modeling Language](./symphony-book/uom/property-expressions.md)
* [**Chapter 4**: Target Management](symphony-book/target-management/target-management.md)    
* [**Chapter 5**: Device Management](symphony-book/device-management/device-management.md)
* [**Chapter 6**: Solution Management](symphony-book/solution-management/solution-management.md)
* [**Chapter 7**: Instance Management]
* [**Chapter 8**: Campaign Management]
* **Chapter 9**: AI Model Management
* [**Chapter 10**: AI Skill Management](symphony-book/skill-management/skill-management.md)
* **Chapter 11**: Policy Management
    * GateKeeper
    * Kyverno
    * Azure Policy
* **Chapter 12**: Security
    * [Authentication](symphony-book/security/authentication.md)
    * [Authorization](symphony-book/security/authorization.md)
    * [CORS Control](symphony-book/bindings/cors.md)
* [**Chapter 13**: Providers](symphony-book/providers/overview.md)
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
    * [Provider Conformance](symphony-book/providers/conformance.md)
    * [Write Custom Providers](symphony-book/providers/python_provider.md)
* [**Chapter 14**: Managers](symphony-book/managers/overview.md)
    * [Solution Manager](symphony-book/managers/solution-manager.md)
* [**Chapter 15**: Vendors](symphony-book/vendors/overview.md)
    * [Jobs Vendor](symphony-book/vendors/job.md)
* [**Chapter 16**: Hosts](symphony-book/hosts/overview.md)
* [**Chapter 17**: Bindings](symphony-book/bindings/overview.md)
    * [HTTP Binding](symphony-book/bindings/http-binding.md)
        * [Application Insight middleware](symphony-book/bindings/app-insight.md)
        * [CORS middleware](symphony-book/bindings/cors.md)
        * [Distributed Tracing middleware](symphony-book/bindings/tracing.md)
        * [JWT Token middleware](symphony-book/bindings/jwt-handler.md)
    * [MQTT Binding](symphony-book/bindings/mqtt-binding.md)
* **Chapter 18**: Capabilities and Capability Matching
* [**Chapter 19**: Agent](symphony-book/agent/agent.md)
    * Symphony Agent
    * Polling Agent
* [**Chapter 20**: Integrations](symphony-book/integrations/overview.md)
    * [Akri](symphony-book/akri/akri.md)
    * AKS
    * AKS Edge Essentials
    * AKS Fleet Management
    * Arc
    * Azure API Management
    * Azure Logic Apps
    * Azure Resource Manager
    
* [**Chapter 21**: Deployment Scenarios](symphony-book/scenarios/development-scenarios.md)    
    * [Adaptive deployments](symphony-book/scenarios/adaptive-deployment.md) - _deploys the same artifact to multiple target types_
    * [Canary deployments](symphony-book/instance-management/instance-management.md#canary-deployment) - _adjustable canary deployments_
    * Cascaded deployments - _deploys along cluster hierarchies_
    * Dynamic deployments - _dynamically reconfiguration based on discovered devices_
    * Fan-out deployments - _deploys to matching targets_
    * [Gated deployments](symphony-book/scenarios/gated-deployment.md) - _adds control points to your deployments_
    * [Human approval](symphony-book/scenarios/human-approval.md) - _requires manual approval_
    * Hybrid deployments - _deploys to both edge and cloud_
    * [Remote deployments over MQTT](symphony-book/scenarios/remote-deployment-over-mqtt.md) - _drive remote PowerShell deployment over MQTT_
    * [Scheduled deployments](symphony-book/instance-management/instance-management.md#scheduled-deployment) - _deploys at given schedule_
    * [Split deployments](symphony-book/scenarios/linux-with-uwp-frontend.md) - _distributes components to matching targets (such as Windows frontend and Linux backend)_
    * Staged deployments - _deploys app across rings_
* [**Chapter 22**: CLI](symphony-book/cli/cli.md)
  * [Usage](symphony-book/cli/cli.md)
  * [Build](symphony-book/cli/build_cli.md)
* [**Chapter 23**: REST API](symphony-book/api/api.md)
    * [Instances API](symphony-book/api/instances-api.md)
    * [Solutions API](symphony-book/api/solutions-api.md)
    * [Targets API](symphony-book/api/targets-api.md)
    * [Query Projection](symphony-book/api/projection.md)
* **Chapter 24**: SDKS
    * Rust SDK
    * Python SDK
* **Chapter 25**: Observability
    * Events
    * [Distributed Tracing](symphony-book/observability/distributed-tracing.md)
    * Logs
    * Metrics
* **Chapter 26**: Federation
    * Component Name Resolutions
    * Federated Queries
* **Chapter 27**: Developer Guide
    * [Getting Started](./symphony-book/dev-guide/getting-started.md)
    * [Debugging](./symphony-book/dev-guide/debugging-api.md)