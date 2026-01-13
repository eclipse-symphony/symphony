# Symphony Python SDK - Features Overview

This document provides a comprehensive overview of all features available in the Symphony Python SDK.

## Table of Contents

- [Client Features](#client-features)
- [Authentication & Security](#authentication--security)
- [Target Management](#target-management)
- [Solution Management](#solution-management)
- [Instance & Deployment](#instance--deployment)
- [COA (Cloud Object API)](#coa-cloud-object-api)
- [Status & Summary Tracking](#status--summary-tracking)
- [Data Models & Type Safety](#data-models--type-safety)
- [Error Handling](#error-handling)
- [Advanced Features](#advanced-features)

---

## Client Features

### SymphonyAPI Client

The main client for all Symphony operations.

**Features:**
- ✅ Context manager support for automatic cleanup
- ✅ Configurable timeout per request
- ✅ Custom logger integration
- ✅ Automatic session management and connection pooling
- ✅ Thread-safe request handling

**Example:**
```python
with SymphonyAPI(base_url, username, password, timeout=60) as client:
    # Auto cleanup when done
    pass
```

### Connection Management

**Features:**
- ✅ Automatic connection pooling via requests.Session
- ✅ Health check endpoint support
- ✅ Graceful connection closure
- ✅ Network error handling

**Example:**
```python
if client.health_check():
    print("Symphony is reachable")
```

---

## Authentication & Security

### Token-Based Authentication

**Features:**
- ✅ Automatic token acquisition
- ✅ Token caching with expiry tracking
- ✅ Automatic token refresh
- ✅ Force refresh capability
- ✅ Secure credential handling

**Example:**
```python
# Authentication happens automatically
client.list_targets()  # Token acquired if needed

# Or force refresh
token = client.authenticate(force_refresh=True)
```

### Security

**Features:**
- ✅ Bearer token authentication
- ✅ HTTPS support
- ✅ No credential storage on disk
- ✅ Session-based credential management

---

## Target Management

Comprehensive target lifecycle management.

### Registration

**Features:**
- ✅ Register new targets with full spec
- ✅ Support for properties, components, metadata
- ✅ Topology and constraint definitions
- ✅ Force redeploy option

**Example:**
```python
target_spec = {
    "displayName": "Edge Gateway",
    "properties": {"location": "dc1", "os": "linux"},
    "components": [...],
    "metadata": {"owner": "team-a"}
}
client.register_target("gateway-001", target_spec)
```

### Querying

**Features:**
- ✅ List all targets
- ✅ Get specific target details
- ✅ Support for YAML and JSON formats
- ✅ JSONPath filtering

**Example:**
```python
# List all
targets = client.list_targets()

# Get specific target
target = client.get_target("gateway-001")

```

### Status Updates

**Features:**
- ✅ Update target status
- ✅ Send heartbeat pings
- ✅ Track last seen time
- ✅ Custom status properties

**Example:**
```python
# Heartbeat
client.ping_target("gateway-001")

# Update status
client.update_target_status("gateway-001", {
    "properties": {"health": "healthy"}
})
```

### Unregistration

**Features:**
- ✅ Graceful unregistration
- ✅ Direct delete option
- ✅ Cleanup of associated resources

**Example:**
```python
client.unregister_target("gateway-001", direct=True)
```

---

## Solution Management

Define and manage application solutions.

### Creation

**Features:**
- ✅ YAML-based solution definitions
- ✅ Multiple component support
- ✅ Embedded specifications
- ✅ Component properties and routing

**Example:**
```python
solution_yaml = """
displayName: Web Application
components:
  - name: nginx
    type: container
    properties:
      image: nginx:latest
"""
client.create_solution("web-app", solution_yaml)
```

### Components

**Features:**
- ✅ Container components
- ✅ Config components
- ✅ Custom component types
- ✅ Component dependencies
- ✅ Component routing and filtering

**Example:**
```python
component = ComponentSpec(
    name="nginx",
    type="container",
    properties={"image": "nginx:latest"},
    dependencies=["config-service"],
    routes=[...]
)
```

### Querying

**Features:**
- ✅ List all solutions
- ✅ Get solution details
- ✅ YAML/JSON format support
- ✅ Component filtering

**Example:**
```python
solutions = client.list_solutions()
solution = client.get_solution("web-app")
```

---

## Instance & Deployment

Deploy solutions to targets.

### Instance Creation

**Features:**
- ✅ Create instances from solutions
- ✅ Target selection with selectors
- ✅ Custom parameters
- ✅ Topology definitions
- ✅ Pipeline configurations
- ✅ Version management

**Example:**
```python
instance_spec = {
    "solution": "web-app",
    "target": {"name": "gateway-001"},
    "parameters": {"replicas": "3"},
    "topologies": [...],
    "pipelines": [...]
}
client.create_instance("web-app-prod", instance_spec)
```

### Deployment Operations

**Features:**
- ✅ Apply deployments
- ✅ Get deployment components
- ✅ Direct reconciliation
- ✅ Delete deployments

**Example:**
```python
# Apply deployment
deployment_spec = {...}
client.apply_deployment(deployment_spec)

# Reconcile
client.reconcile_solution(deployment_spec, delete=False)

# Get components
components = client.get_deployment_components()
```

### Status Tracking

**Features:**
- ✅ Real-time instance status
- ✅ Deployment progress tracking
- ✅ Success/failure detection

**Example:**
```python
status = client.get_instance_status("web-app-prod")
if status.get('status') == 'Succeeded':
    print("Deployment successful!")
```

---

## COA (Cloud Object API)

Unified interface for provider operations.

### COA Requests

**Features:**
- ✅ Multiple HTTP methods (GET, POST, PUT, DELETE)
- ✅ Flexible routing
- ✅ Custom parameters and metadata
- ✅ Multiple content types:
  - JSON (application/json)
  - Plain text (text/plain)
  - Binary (application/octet-stream)
- ✅ Automatic base64 encoding/decoding

**Example:**
```python
request = COARequest(
    method="POST",
    route="/components/deploy",
    content_type="application/json"
)
request.set_body({"component": "nginx", "version": "latest"})

# Get decoded body
data = request.get_body()
```

### COA Responses

**Features:**
- ✅ Rich state enum (200+ states)
- ✅ Convenience methods for common responses
- ✅ Metadata support
- ✅ Redirect URI support
- ✅ Body encoding/decoding

**Example:**
```python
# Success response
response = COAResponse.success({"status": "deployed"})

# Error responses
error = COAResponse.error("Failed", State.INTERNAL_ERROR)
not_found = COAResponse.not_found("Component not found")
bad_request = COAResponse.bad_request("Invalid input")
```

### State Management

**Features:**
- ✅ Comprehensive state enum with 100+ values
- ✅ HTTP status code mapping
- ✅ Human-readable state strings
- ✅ Custom state codes for Symphony operations

**Example:**
```python
state = State.from_http_status(404)  # NOT_FOUND
print(str(state))  # "Not Found"

if state == State.NOT_FOUND:
    print("Resource not found")
```

### Serialization

**Features:**
- ✅ JSON serialization/deserialization
- ✅ Content type handling
- ✅ Base64 encoding for binary data

**Example:**
```python
# Serialize
json_str = serialize_coa_request(request)

# Deserialize
request = deserialize_coa_request(json_str)
```

---

## Status & Summary Tracking

Track deployment status and generate summaries.

### Summary Specifications

**Features:**
- ✅ Target-level tracking
- ✅ Component-level tracking
- ✅ Success/failure counters
- ✅ Deployment progress metrics
- ✅ Custom messages

**Example:**
```python
summary = SummarySpec(
    target_count=3,
    success_count=3,
    planned_deployment=10,
    current_deployed=10,
    all_assigned_deployed=True
)
```

### Component Results

**Features:**
- ✅ Per-component status
- ✅ Success/failure tracking
- ✅ Error messages
- ✅ State codes

**Example:**
```python
# Success
comp_result = create_success_component_result("Deployed")

# Failure
comp_result = create_failed_component_result(
    "Image pull failed",
    State.INTERNAL_ERROR
)
```

### Target Results

**Features:**
- ✅ Aggregate target status
- ✅ Component result collection
- ✅ Target-level messages
- ✅ Result merging

**Example:**
```python
target_result = create_target_result(
    status="OK",
    message="All components deployed",
    component_results={
        "nginx": success_result,
        "redis": success_result
    }
)

summary.update_target_result("target-1", target_result)
```

### Summary Results

**Features:**
- ✅ Complete deployment summary
- ✅ State tracking (PENDING, RUNNING, DONE)
- ✅ Timestamp tracking
- ✅ Deployment hash
- ✅ Detailed status message generation
- ✅ Serialization support

**Example:**
```python
result = SummaryResult(
    summary=summary,
    summary_id="deploy-001",
    state=SummaryState.DONE,
    deployment_hash="abc123"
)

if result.is_deployment_finished():
    message = result.summary.generate_status_message()
```

---

## Data Models & Type Safety

Comprehensive data models with full type hints.

### Resource Metadata

**Features:**
- ✅ Kubernetes-style metadata
- ✅ Namespace support
- ✅ Labels and annotations
- ✅ Resource naming

**Example:**
```python
metadata = ObjectMeta(
    name="my-resource",
    namespace="default",
    labels={"env": "prod", "app": "web"},
    annotations={"owner": "team-a"}
)
```

### Specifications

**Features:**
- ✅ TargetSpec with full target definition
- ✅ SolutionSpec with components
- ✅ InstanceSpec with deployment config
- ✅ DeploymentSpec with complete state
- ✅ ComponentSpec with properties and routing

### Type Conversion

**Features:**
- ✅ to_dict() for serialization
- ✅ from_dict() for deserialization
- ✅ Nested object support
- ✅ List and dict handling
- ✅ Enum support

**Example:**
```python
# Convert to dict
spec_dict = to_dict(target_spec)

# Convert from dict
target_spec = from_dict(spec_dict, TargetSpec)
```

### Component Serialization

**Features:**
- ✅ JSON serialization
- ✅ Component list handling
- ✅ Solution state serialization
- ✅ Deployment spec serialization

**Example:**
```python
# Serialize components
json_str = serialize_components([comp1, comp2])

# Deserialize
components = deserialize_components(json_str)
```

---

## Error Handling

Comprehensive error handling and reporting.

### SymphonyAPIError

**Features:**
- ✅ Custom exception class
- ✅ HTTP status code capture
- ✅ Response text capture
- ✅ Detailed error messages

**Example:**
```python
try:
    client.register_target("device", spec)
except SymphonyAPIError as e:
    print(f"Error: {e}")
    print(f"Status: {e.status_code}")
    print(f"Details: {e.response_text}")
```

### Error Categories

**Features:**
- ✅ Network errors (connection, timeout)
- ✅ Authentication errors (401)
- ✅ Authorization errors (403)
- ✅ Not found errors (404)
- ✅ Server errors (500+)
- ✅ Validation errors (400)

### Retry Support

**Features:**
- ✅ Retry logic patterns in examples
- ✅ Exponential backoff support
- ✅ Configurable retry attempts
- ✅ Retryable error detection

---

## Advanced Features

### Logging Integration

**Features:**
- ✅ Custom logger support
- ✅ Request/response logging
- ✅ Debug mode support
- ✅ Configurable log levels

**Example:**
```python
logger = logging.getLogger("my-app")
client = SymphonyAPI(url, user, pass, logger=logger)
```

### Timeout Configuration

**Features:**
- ✅ Per-client timeout configuration
- ✅ Configurable default timeout
- ✅ Per-request timeout override

**Example:**
```python
client = SymphonyAPI(url, user, pass, timeout=120.0)
```

### Session Management

**Features:**
- ✅ Persistent HTTP sessions
- ✅ Connection pooling
- ✅ Automatic cleanup
- ✅ Thread safety

### Content Negotiation

**Features:**
- ✅ YAML response support
- ✅ JSON response support
- ✅ Plain text support
- ✅ Content-Type handling

### Advanced Querying

**Features:**
- ✅ JSONPath support for filtering
- ✅ Document type selection
- ✅ Path-based extraction
- ✅ Query parameters

**Example:**
```python
# Extract specific field
properties = client.get_target(
    "device-001",
    doc_type="json",
    path="$.spec.properties"
)
```

---

## Performance Features

### Efficiency

- ✅ Connection pooling reduces overhead
- ✅ Token caching minimizes auth requests
- ✅ Minimal dependencies (only `requests`)
- ✅ Efficient JSON parsing
- ✅ Base64 encoding only when needed

### Scalability

- ✅ Thread-safe client
- ✅ Support for multiple concurrent requests
- ✅ No global state
- ✅ Resource cleanup

---

## Development Features

### Testing Support

- ✅ Comprehensive test suite
- ✅ Mock-friendly design
- ✅ Example unit tests
- ✅ High code coverage

### Documentation

- ✅ Detailed docstrings
- ✅ Type hints throughout
- ✅ Usage examples
- ✅ API reference
- ✅ Quick start guide

### IDE Support

- ✅ Full type hints for autocomplete
- ✅ Dataclass support
- ✅ Clear error messages
- ✅ Comprehensive documentation strings

---

## Supported Operations

### Target Operations
✅ Register, ✅ Unregister, ✅ List, ✅ Get, ✅ Ping, ✅ Update Status

### Solution Operations
✅ Create, ✅ Delete, ✅ List, ✅ Get

### Instance Operations
✅ Create, ✅ Delete, ✅ List, ✅ Get, ✅ Get Status

### Deployment Operations
✅ Apply, ✅ Reconcile, ✅ Get Components, ✅ Delete Components

### Utility Operations
✅ Authenticate, ✅ Health Check, ✅ Get Config

---

## Platform Support

- ✅ Linux
- ✅ macOS
- ✅ Windows
- ✅ Docker/Containers
- ✅ Kubernetes pods

## Python Version Support

- ✅ Python 3.9
- ✅ Python 3.10
- ✅ Python 3.11
- ✅ Python 3.12
- ✅ Python 3.13

---

## See Also

- [Quick Start Guide](QUICKSTART.md)
- [API Reference](API.md)
- [Examples](../examples/)
- [Main README](../README.md)
