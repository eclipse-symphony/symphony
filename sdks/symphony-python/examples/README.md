# Symphony Python SDK Examples

Example programs to demonstrate what the SDK is capable of.

## Prerequisites

```bash
# Install from source
pip install -e .
```

## Examples Overview

### 1. Basic Client Usage ([01_basic_client.py](01_basic_client.py))

Learn the fundamentals of using the Symphony API client:
- Initializing the client
- Authentication
- Health checks
- Context manager usage
- Session management

**Key concepts:**
```python
from symphony_sdk import SymphonyAPI

with SymphonyAPI(base_url, username, password) as client:
    # Authentication happens automatically
    if client.health_check():
        print("Connected!")
```

### 2. Target Management ([02_target_management.py](02_target_management.py))

Working with Symphony targets (devices, edge nodes, etc.):
- Registering new targets
- Listing all targets
- Getting target details
- Updating target status
- Sending heartbeat pings
- Unregistering targets
- Using dataclasses for type safety

**Key concepts:**
```python
# Register a target
target_spec = {
    "displayName": "IoT Device",
    "properties": {"location": "datacenter-1"}
}
client.register_target("device-001", target_spec)

# Or use dataclasses
from symphony_sdk import TargetSpec, TargetState
target = TargetSpec(displayName="IoT Device", ...)
```

### 3. Solution and Instance Management ([03_solution_deployment.py](03_solution_deployment.py))

Managing solutions and deploying instances:
- Creating solutions with YAML specs
- Creating instances from solutions
- Applying deployments
- Checking deployment status
- Listing solutions and instances
- Resource cleanup
- Using InstanceSpec dataclass

**Key concepts:**
```python
# Create a solution
solution_yaml = yaml.dump({...})
client.create_solution("web-app", solution_yaml)

# Create an instance
instance_spec = {
    "solution": "web-app",
    "target": {"name": "device-001"}
}
client.create_instance("web-app-prod", instance_spec)
```

### 4. COA Provider ([04_coa_provider.py](04_coa_provider.py))

Working with Cloud Object API (COA) requests and responses:
- Creating COA requests and responses
- Handling different content types (JSON, text, binary)
- Body encoding/decoding
- Building providers
- Working with DeploymentSpec

**Key concepts:**
```python
from symphony_sdk import COARequest, COAResponse, State

# Create a request
request = COARequest(method="GET", route="/components")
request.set_body({"filter": "active"})

# Create a response
response = COAResponse.success({"components": [...]})
response = COAResponse.error("Not found", State.NOT_FOUND)
```

### 5. Summary and Status Tracking ([05_summary_tracking.py](05_summary_tracking.py))

Tracking deployment progress and status:
- Creating summary results
- Tracking component status
- Handling successful/failed deployments
- Incremental status updates
- Generating status messages
- Serialization

**Key concepts:**
```python
from symphony_sdk import (
    SummaryResult, SummarySpec, SummaryState,
    create_success_component_result,
    create_failed_component_result
)

# Track deployment
summary = SummarySpec(target_count=2, planned_deployment=4)
summary.update_target_result("target-1", target_result)

result = SummaryResult(summary=summary, state=SummaryState.DONE)
if result.is_deployment_finished():
    print("Deployment complete!")
```

### 6. Error Handling and Best Practices ([06_error_handling.py](06_error_handling.py))

Production-ready error handling patterns:
- Catching and handling SymphonyAPIError
- Implementing retry logic with exponential backoff
- Connection and authentication validation
- Timeout handling
- Input validation
- Comprehensive error handling wrappers
- Graceful degradation strategies

**Key concepts:**
```python
from symphony_sdk import SymphonyAPIError

# Basic error handling
try:
    client.register_target("device", spec)
except SymphonyAPIError as e:
    if e.status_code == 404:
        print("Not found")
    elif e.status_code == 500:
        print("Server error - retry")

# Retry with exponential backoff
def retry_with_backoff(func, max_retries=3):
    for attempt in range(max_retries):
        try:
            return func()
        except SymphonyAPIError as e:
            if attempt < max_retries - 1:
                time.sleep(2 ** attempt)
    raise
```

## Running the Examples

### Update Credentials

Most examples require you to update the credentials before running:

```python
# Update these values
base_url = "https://your-symphony-instance.com"
username = "your-username"
password = "your-password"
```

### Run an Example

```bash
# Make the script executable
chmod +x examples/01_basic_client.py

# Run it
python examples/01_basic_client.py
```

### COA Provider Example

The COA provider example can be run directly without credentials:

```bash
python examples/04_coa_provider.py
```

### Summary Tracking Example

The summary tracking example demonstrates data structures and can also be run directly:

```bash
python examples/05_summary_tracking.py
```

## Common Patterns

### Error Handling

```python
from symphony_sdk import SymphonyAPIError

try:
    client.register_target(name, spec)
except SymphonyAPIError as e:
    print(f"Error: {e}")
    print(f"Status code: {e.status_code}")
    print(f"Response: {e.response_text}")
```

### Using Context Managers

```python
# Recommended: Ensures proper cleanup
with SymphonyAPI(base_url, username, password) as client:
    # Your code here
    pass
# Session automatically closed
```

### Working with Dataclasses

```python
from symphony_sdk import TargetSpec, ComponentSpec, InstanceSpec

# Type-safe approach
component = ComponentSpec(
    name="nginx",
    type="container",
    properties={"image": "nginx:latest"}
)

target = TargetSpec(
    displayName="My Device",
    components=[component]
)
```

### Checking Deployment Status

```python
import time

# Poll for completion
for attempt in range(10):
    status = client.get_instance_status(instance_name)
    if status.get('status') == 'Succeeded':
        break
    time.sleep(2)
```

## Data Models

The SDK provides comprehensive data models matching Symphony's COA specification:

- **Target Models**: `TargetSpec`, `TargetState`, `TargetStatus`
- **Solution Models**: `SolutionSpec`, `SolutionState`, `ComponentSpec`
- **Instance Models**: `InstanceSpec`, `DeploymentSpec`
- **COA Models**: `COARequest`, `COAResponse`, `State`
- **Summary Models**: `SummaryResult`, `SummarySpec`, `ComponentResultSpec`

See [models.py](../src/symphony_sdk/models.py), [types.py](../src/symphony_sdk/types.py), and [summary.py](../src/symphony_sdk/summary.py) for full details.

## Next Steps

1. Review the [main README](../README.md) for installation and setup
2. Check the [API documentation](../docs/API.md) for detailed API reference
3. Explore the [test files](../tests/) for more usage examples
4. Read the Symphony documentation for architecture concepts

## Need Help?

- Check the [Symphony documentation](https://github.com/eclipse-symphony/symphony)
- Review the source code in [src/symphony_sdk/](../src/symphony_sdk/)
- Look at the unit tests in [tests/](../tests/)
- Open an issue on the Symphony GitHub repository

## Contributing

Found a bug or want to add an example? Contributions are welcome! Please follow the Symphony contribution guidelines.
