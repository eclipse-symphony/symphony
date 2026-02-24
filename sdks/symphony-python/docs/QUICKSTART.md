# Symphony Python SDK - Quick Start Guide

Get started with the Symphony Python SDK in minutes.

## Installation

```bash
pip install symphony-sdk
```

Or install from source:

```bash
git clone https://github.com/eclipse-symphony/symphony.git
cd symphony/sdks/symphony-python
pip install -e .
```

## Basic Usage

### 1. Connect to Symphony

```python
from symphony_sdk import SymphonyAPI

# Initialize the client
client = SymphonyAPI(
    base_url="https://symphony.example.com",
    username="your-username",
    password="your-password"
)

# Use context manager (recommended)
with SymphonyAPI(base_url, username, password) as client:
    # Your code here
    pass
```

### 2. Work with Targets

Targets represent devices, edge nodes, or compute resources.

```python
# Register a new target
target_spec = {
    "displayName": "Edge Gateway 001",
    "scope": "default",
    "properties": {
        "location": "datacenter-1",
        "os": "linux",
        "arch": "amd64"
    }
}

client.register_target("gateway-001", target_spec)

# List all targets
targets = client.list_targets()
for target in targets.get('items', []):
    print(target['metadata']['name'])

# Get target details
target = client.get_target("gateway-001")

# Send heartbeat
client.ping_target("gateway-001")

# Unregister target
client.unregister_target("gateway-001")
```

### 3. Deploy Solutions

Solutions define application components to be deployed.

```python
import yaml

# Define a solution
solution = {
    "displayName": "Web Application",
    "scope": "default",
    "components": [
        {
            "name": "nginx",
            "type": "container",
            "properties": {
                "container.image": "nginx:latest",
                "container.ports": "80:8080"
            }
        }
    ]
}

# Create the solution
solution_yaml = yaml.dump(solution)
client.create_solution("web-app", solution_yaml)

# Create an instance (deployment)
instance_spec = {
    "solution": "web-app",
    "target": {"name": "gateway-001"},
    "scope": "default",
    "displayName": "Production Web App"
}

client.create_instance("web-app-prod", instance_spec)

# Check deployment status
status = client.get_instance_status("web-app-prod")
print(f"Status: {status.get('status')}")
```

### 4. Use COA (Cloud Object API)

COA provides a unified interface for provider operations.

```python
from symphony_sdk import COARequest, COAResponse, State

# Create a request
request = COARequest(
    method="GET",
    route="/components/list",
    content_type="application/json"
)
request.set_body({"filter": "active"})

# Create a response
response = COAResponse.success({
    "components": ["nginx", "redis"]
})

# Handle errors
error_response = COAResponse.error(
    "Component not found",
    state=State.NOT_FOUND
)
```

### 5. Track Deployment Status

```python
from symphony_sdk import (
    SummaryResult, SummarySpec, SummaryState,
    create_success_component_result,
    create_target_result
)

# Create a deployment summary
summary = SummarySpec(
    target_count=1,
    success_count=1,
    planned_deployment=2,
    current_deployed=2,
    all_assigned_deployed=True
)

# Add target results
target_result = create_target_result(
    status="OK",
    message="All components deployed",
    component_results={
        "nginx": create_success_component_result("Deployed"),
        "redis": create_success_component_result("Deployed")
    }
)
summary.update_target_result("gateway-001", target_result)

# Create summary result
result = SummaryResult(
    summary=summary,
    state=SummaryState.DONE
)

if result.is_deployment_finished():
    print("Deployment complete!")
```

## Common Patterns

### Error Handling

```python
from symphony_sdk import SymphonyAPIError

try:
    client.register_target("device-001", spec)
except SymphonyAPIError as e:
    print(f"API Error: {e}")
    if e.status_code:
        print(f"Status Code: {e.status_code}")
    if e.response_text:
        print(f"Details: {e.response_text}")
```

### Polling for Completion

```python
import time

def wait_for_deployment(client, instance_name, timeout=300):
    """Wait for deployment to complete."""
    start_time = time.time()

    while time.time() - start_time < timeout:
        try:
            status = client.get_instance_status(instance_name)
            state = status.get('status', 'Unknown')

            if state == 'Succeeded':
                return True
            elif state == 'Failed':
                return False

            time.sleep(5)  # Check every 5 seconds
        except SymphonyAPIError:
            time.sleep(5)

    return False  # Timeout

# Use it
if wait_for_deployment(client, "web-app-prod"):
    print("Deployment succeeded!")
else:
    print("Deployment failed or timed out")
```

### Using Dataclasses for Type Safety

```python
from symphony_sdk import (
    TargetSpec, ComponentSpec, TargetState,
    ObjectMeta, InstanceSpec, TargetSelector
)

# Create target with dataclasses
metadata = ObjectMeta(
    name="gateway-001",
    namespace="default",
    labels={"env": "prod"}
)

component = ComponentSpec(
    name="nginx",
    type="container",
    properties={"image": "nginx:latest"}
)

target_spec = TargetSpec(
    displayName="Edge Gateway",
    components=[component],
    properties={"location": "dc1"}
)

target = TargetState(
    metadata=metadata,
    spec=target_spec
)

# Create instance with dataclasses
instance = InstanceSpec(
    name="my-app",
    solution="web-app",
    target=TargetSelector(name="gateway-001"),
    parameters={"replicas": "3"}
)
```

### Batch Operations

```python
# Register multiple targets
targets = [
    ("device-001", {"displayName": "Device 1", ...}),
    ("device-002", {"displayName": "Device 2", ...}),
    ("device-003", {"displayName": "Device 3", ...}),
]

for name, spec in targets:
    try:
        client.register_target(name, spec)
        print(f"✓ Registered {name}")
    except SymphonyAPIError as e:
        print(f"✗ Failed to register {name}: {e}")
```

## Configuration

### Custom Timeout

```python
# Set custom timeout (default is 30 seconds)
client = SymphonyAPI(
    base_url=base_url,
    username=username,
    password=password,
    timeout=60.0  # 60 seconds
)
```

### Custom Logger

```python
import logging

# Create custom logger
logger = logging.getLogger("my-app")
logger.setLevel(logging.DEBUG)
handler = logging.StreamHandler()
handler.setFormatter(logging.Formatter(
    '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
))
logger.addHandler(handler)

# Use with client
client = SymphonyAPI(
    base_url=base_url,
    username=username,
    password=password,
    logger=logger
)
```

## Next Steps

- Explore the [examples directory](../examples/) for more detailed examples
- Read the [API Reference](API.md) for complete API documentation
- Check the [Symphony documentation](https://github.com/eclipse-symphony/symphony) for architecture concepts
- Review the [CHANGELOG](../CHANGELOG.md) for version history

## Common Issues

### Authentication Errors

```python
# Ensure credentials are correct
try:
    token = client.authenticate()
    print("Authentication successful!")
except SymphonyAPIError as e:
    if e.status_code == 401:
        print("Invalid credentials")
    else:
        print(f"Auth error: {e}")
```

### Connection Errors

```python
# Check if Symphony is reachable
if client.health_check():
    print("Connected to Symphony")
else:
    print("Cannot reach Symphony API")
    print(f"URL: {client.base_url}")
```

### Timeout Issues

```python
# Increase timeout for slow operations
client = SymphonyAPI(
    base_url=base_url,
    username=username,
    password=password,
    timeout=120.0  # 2 minutes
)
```

## Support

- Report issues on [GitHub](https://github.com/eclipse-symphony/symphony/issues)
- Check the [main documentation](../README.md)
- Review [test files](../tests/) for more examples
