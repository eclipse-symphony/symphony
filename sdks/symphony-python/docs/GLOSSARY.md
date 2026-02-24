# Symphony Python SDK Glossary

This glossary defines commonly used terms and concepts in the Symphony Python SDK.

## Core Resource Concepts

### Target
A device, edge node, or compute resource that can execute solutions. Examples include IoT devices, edge gateways, or Kubernetes clusters. Targets must be registered with Symphony before deploying instances to them. Managed through lifecycle operations: register, unregister, list, get, and ping.

### Solution
A deployment specification that defines application components to be deployed. Solutions are YAML-based templates containing component definitions, properties, and metadata. They serve as reusable templates for creating multiple instances.

### Instance
A concrete deployment of a solution onto specific targets. Instances bind solutions to targets and track deployment status. The actual deployment orchestration happens at the instance level.

### Component
Individual units within a solution, such as containers, configurations, or scripts. Each component has a name, type, properties, and optional dependencies. Components can be container-based, config-based, or custom types.

### Deployment
The execution context that ties together solutions, instances, targets, and devices. DeploymentSpec contains the complete state including solution definition, instance configuration, target information, and device assignments.

## Configuration and Metadata

### ObjectMeta
Kubernetes-style metadata following standard conventions. Contains name, namespace, labels, and annotations for resource identification and organization.

### TargetSelector
Specifies which targets a component or instance should be deployed to. Supports name-based selection or label selectors for flexible target matching.

### ComponentSpec
Detailed specification for a component including name, type, properties, constraints, dependencies, skills, routes, and parameters.

### SolutionSpec
The specification of a solution containing components, scope, display name, and metadata.

### InstanceSpec
Specification for an instance deployment including solution reference, target selector, parameters, topologies, pipelines, and version information.

### Scope
A namespace or partition for organizing Symphony resources. Allows multi-tenancy and resource isolation (e.g., "default" scope).

## Advanced Deployment Concepts

### Topology
Defines the physical or logical placement of components on devices within targets. Specifies devices, selectors, and bindings for component distribution across a target's infrastructure.

### Binding
Specifies how components are bound to providers or roles. Contains role, provider, and configuration information for component lifecycle management.

### Pipeline
A data processing or orchestration pipeline specification. Can reference skills and parameters for component workflows or data transformations.

### VersionSpec
Manages solution versioning with percentage-based canary deployments. Supports multi-version deployments with traffic/load distribution percentages.

### Device
Physical hardware entities within a target. DeviceSpec contains device properties, bindings, and display information.

## COA (Cloud Object API) Concepts

### COARequest
HTTP-method request abstraction for provider operations. Contains method (GET/POST/PUT/DELETE), route, content type, body (base64 encoded), metadata, and parameters.

### COAResponse
Response abstraction with state, content type, body encoding, metadata, and optional redirect URI. Provides factory methods for common responses: `success()`, `error()`, `not_found()`, `bad_request()`.

### State
Comprehensive enumeration (100+ values) representing operation results. Maps HTTP status codes to operation states, and includes custom Symphony states for configuration errors, async operations, workflow status, and specific provider failures.

### Content Type
COA bodies support multiple content types: "application/json", "text/plain", "application/octet-stream". Bodies are base64 encoded when transmitted.

## Status and Progress Tracking

### SummarySpec
Aggregates deployment results across targets and components. Tracks target counts, success counts, planned deployments, current deployments, and whether all assigned components are deployed.

### ComponentResultSpec
Tracks individual component deployment status. Contains status (State enum) and message for component-level outcome.

### TargetResultSpec
Aggregates component results for a target. Contains target-level status string, message, and map of component results.

### SummaryResult
Complete deployment operation summary including summary specification, state (PENDING/RUNNING/DONE), timestamps, and deployment hash.

### SummaryState
Enumeration for summary operation state: PENDING (unused), RUNNING (reconciliation in progress), DONE (reconciliation completed).

## Provider and Configuration Concepts

### Provider
A backend service that handles specific operations or resources. Referenced by bindings and can include config providers, secret providers, state providers, probe providers, reporters, queue providers, etc.

### Skill
A named capability that a component possesses. Components can declare skills they implement, which are then referenced by pipelines for orchestration.

### Route
Specifies routing rules for a component. Includes route pattern, filters, properties, and type for message/request routing between components.

### Filter
Routing filter specification with direction, parameters, and type for conditional message/request processing.

### Constraint
Deployment constraints for components or targets. Used to express deployment requirements or restrictions (e.g., affinity, anti-affinity rules).

## API and Client Concepts

### SymphonyAPI
The main REST client for Symphony. Manages authentication, session pooling, timeout, and logging. Use as a context manager for automatic cleanup. Supports token-based authentication with automatic refresh.

### Namespace
API scope for resource organization. Most operations default to `namespace="default"`. Used for multi-tenancy isolation.

### Doc Type
Response format parameter. Accepts "yaml" or "json". Affects how specs are returned (defaults to YAML for specs).

### JSONPath
Query parameter for selective field extraction from responses using JSONPath expressions (e.g., `"$.spec.properties"`).

### Authentication Token
Bearer token returned by `authenticate()` method. Automatically managed and cached by SymphonyAPI client with refresh on expiry.

## Lifecycle and Operations

### Reconciliation
The process of ensuring actual state matches desired state. Triggered via `reconcile_solution()` to synchronize deployments with target state.

### Force Redeploy
A target property indicating components should be redeployed even if already present. Part of TargetSpec configuration.

### Provisioning Status
Detailed deployment operation status including operation ID, status string, failure cause, errors, and output. Tracks Kubernetes operations, Helm operations, and infrastructure provisioning.

### Health Check
API connectivity verification. Returns boolean indicating Symphony API accessibility.

### Ping/Heartbeat
Target keep-alive mechanism. Targets send periodic pings to indicate they're active and responsive.

### Opt-out Reconciliation
InstanceSpec property allowing instances to skip automatic reconciliation when enabled.

## Common Symphony Workflows

### Register Target
Initial operation to register a device or edge node with Symphony. Must be done before deploying instances to a target.

### Create Solution
Define a template of components and resources. Solutions are reusable across multiple instances.

### Create Instance
Instantiate a solution on specific target(s). Triggers deployment orchestration.

### Monitor Status
Track instance and deployment progress via `get_instance_status()` and SummaryResult tracking.

### Reconcile
Force synchronization between desired (instance spec) and actual (deployed) state on targets.

### Cleanup
Remove instances, solutions, and unregister targets. Supports direct delete option to bypass cleanup pipelines.
