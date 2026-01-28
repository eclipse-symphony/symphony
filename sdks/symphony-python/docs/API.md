# Symphony Python SDK - API Reference

Complete API reference for the Symphony Python SDK.

## Table of Contents

- [SymphonyAPI Client](#symphonyapi-client)
- [Data Models](#data-models)
- [COA Types](#coa-types)
- [Summary Models](#summary-models)
- [Utility Functions](#utility-functions)
- [Exceptions](#exceptions)

---

## SymphonyAPI Client

The main client for interacting with the Symphony REST API.

### Class: `SymphonyAPI`

```python
SymphonyAPI(
    base_url: str,
    username: str,
    password: str,
    timeout: float = 30.0,
    logger: Optional[logging.Logger] = None
)
```

**Parameters:**
- `base_url`: Base URL of the Symphony API (e.g., 'https://symphony.example.com')
- `username`: Symphony username for authentication
- `password`: Symphony password for authentication
- `timeout`: Request timeout in seconds (default: 30.0)
- `logger`: Optional logger instance

**Context Manager:**
Supports context manager protocol for automatic session cleanup.

```python
with SymphonyAPI(base_url, username, password) as client:
    # Use client
    pass
```

### Authentication Methods

#### `authenticate(force_refresh: bool = False) -> str`

Authenticate with Symphony API and return access token.

**Parameters:**
- `force_refresh`: Force token refresh even if current token is valid

**Returns:** Access token string

**Raises:** `SymphonyAPIError` if authentication fails

### Target Management Methods

#### `register_target(target_name: str, target_spec: Dict[str, Any]) -> Dict[str, Any]`

Register a target with Symphony.

**Parameters:**
- `target_name`: Name of the target to register
- `target_spec`: Target specification dictionary

**Returns:** API response data

#### `unregister_target(target_name: str, direct: bool = False) -> Dict[str, Any]`

Unregister a target from Symphony.

**Parameters:**
- `target_name`: Name of the target to unregister
- `direct`: Whether to use direct delete

**Returns:** API response data

#### `get_target(target_name: str, doc_type: str = 'yaml', path: str = '$.spec') -> Dict[str, Any]`

Get target specification.

**Parameters:**
- `target_name`: Name of the target
- `doc_type`: Document type ('yaml' or 'json')
- `path`: JSONPath to extract from response

**Returns:** Target specification data

#### `list_targets() -> Dict[str, Any]`

List all registered targets.

**Returns:** List of targets

#### `ping_target(target_name: str) -> Dict[str, Any]`

Send heartbeat ping to target.

**Parameters:**
- `target_name`: Name of the target to ping

**Returns:** Ping response data

#### `update_target_status(target_name: str, status_data: Dict[str, Any]) -> Dict[str, Any]`

Update target status.

**Parameters:**
- `target_name`: Name of the target
- `status_data`: Status information dictionary

**Returns:** API response data

### Solution Management Methods

#### `create_solution(solution_name: str, solution_spec: str, embed_type: Optional[str] = None, embed_component: Optional[str] = None, embed_property: Optional[str] = None) -> Dict[str, Any]`

Create a solution with embedded specification.

**Parameters:**
- `solution_name`: Name of the solution
- `solution_spec`: Solution specification as text (usually YAML)
- `embed_type`: Optional embed type
- `embed_component`: Optional embed component
- `embed_property`: Optional embed property

**Returns:** API response data

#### `get_solution(solution_name: str, doc_type: str = 'yaml', path: str = '$.spec') -> Dict[str, Any]`

Get solution specification.

**Parameters:**
- `solution_name`: Name of the solution
- `doc_type`: Document type ('yaml' or 'json')
- `path`: JSONPath to extract from response

**Returns:** Solution specification data

#### `delete_solution(solution_name: str) -> Dict[str, Any]`

Delete a solution.

**Parameters:**
- `solution_name`: Name of the solution to delete

**Returns:** API response data

#### `list_solutions() -> Dict[str, Any]`

List all solutions.

**Returns:** List of solutions

### Instance Management Methods

#### `create_instance(instance_name: str, instance_spec: Dict[str, Any]) -> Dict[str, Any]`

Create an instance.

**Parameters:**
- `instance_name`: Name of the instance
- `instance_spec`: Instance specification dictionary

**Returns:** API response data

#### `get_instance(instance_name: str, doc_type: str = 'yaml', path: str = '$.spec') -> Dict[str, Any]`

Get instance specification.

**Parameters:**
- `instance_name`: Name of the instance
- `doc_type`: Document type ('yaml' or 'json')
- `path`: JSONPath to extract from response

**Returns:** Instance specification data

#### `delete_instance(instance_name: str) -> Dict[str, Any]`

Delete an instance.

**Parameters:**
- `instance_name`: Name of the instance to delete

**Returns:** API response data

#### `list_instances() -> Dict[str, Any]`

List all instances.

**Returns:** List of instances

### Deployment Methods

#### `apply_deployment(deployment_spec: Dict[str, Any]) -> Dict[str, Any]`

Apply a deployment.

**Parameters:**
- `deployment_spec`: Deployment specification dictionary

**Returns:** API response data

#### `get_deployment_components() -> Dict[str, Any]`

Get deployment components.

**Returns:** Components data

#### `delete_deployment_components() -> Dict[str, Any]`

Delete deployment components.

**Returns:** API response data

#### `reconcile_solution(deployment_spec: Dict[str, Any], delete: bool = False) -> Dict[str, Any]`

Direct reconcile/delete deployment.

**Parameters:**
- `deployment_spec`: Deployment specification dictionary
- `delete`: Whether this is a delete operation

**Returns:** API response data

#### `get_instance_status(instance_name: str) -> Dict[str, Any]`

Get instance status.

**Parameters:**
- `instance_name`: Name of the instance

**Returns:** Instance status data

### Utility Methods

#### `health_check() -> bool`

Perform a basic health check of the Symphony API.

**Returns:** True if API is accessible, False otherwise

#### `close()`

Close the HTTP session.

---

## Data Models

### ObjectMeta

Kubernetes-style metadata for resources.

```python
@dataclass
class ObjectMeta:
    namespace: str = ""
    name: str = ""
    labels: Optional[Dict[str, str]] = None
    annotations: Optional[Dict[str, str]] = None
```

### TargetSpec

Specification for a Symphony target.

```python
@dataclass
class TargetSpec:
    properties: Dict[str, str] = None
    components: List[ComponentSpec] = None
    constraints: str = ""
    topologies: List[TopologySpec] = None
    scope: str = ""
    displayName: str = ""
    metadata: Dict[str, str] = None
    forceRedeploy: bool = False
```

### ComponentSpec

Specification for a component.

```python
@dataclass
class ComponentSpec:
    name: str = ""
    type: str = ""
    routes: List[RouteSpec] = None
    constraints: str = ""
    properties: Dict[str, str] = None
    dependencies: List[str] = None
    skills: List[str] = None
    metadata: Dict[str, str] = None
    parameters: Dict[str, str] = None
```

### SolutionSpec

Specification for a solution.

```python
@dataclass
class SolutionSpec:
    components: List[ComponentSpec] = None
    scope: str = ""
    displayName: str = ""
    metadata: Dict[str, str] = None
```

### InstanceSpec

Specification for an instance deployment.

```python
@dataclass
class InstanceSpec:
    name: str = ""
    parameters: Optional[Dict[str, str]] = None
    solution: str = ""
    target: Optional[TargetSelector] = None
    topologies: Optional[List[TopologySpec]] = None
    pipelines: Optional[List[PipelineSpec]] = None
    scope: str = ""
    display_name: str = ""
    metadata: Optional[Dict[str, str]] = None
    versions: Optional[List[VersionSpec]] = None
    arguments: Optional[Dict[str, Dict[str, str]]] = None
    opt_out_reconciliation: bool = False
```

### DeploymentSpec

Complete deployment specification.

```python
@dataclass
class DeploymentSpec:
    solutionName: str = ""
    solution: SolutionState = None
    instance: InstanceSpec = None
    targets: Dict[str, TargetState] = None
    devices: List[DeviceSpec] = None
    assignments: Dict[str, str] = None
    componentStartIndex: int = -1
    componentEndIndex: int = -1
    activeTarget: str = ""

    def get_components_slice(self) -> List[ComponentSpec]:
        """Get slice of components for deployment."""
```

---

## COA Types

### State

Enumeration of COA response states.

```python
class State(IntEnum):
    # HTTP states
    OK = 200
    ACCEPTED = 202
    BAD_REQUEST = 400
    UNAUTHORIZED = 401
    FORBIDDEN = 403
    NOT_FOUND = 404
    INTERNAL_ERROR = 500

    # Custom states
    BAD_CONFIG = 1000
    INVALID_ARGUMENT = 2000
    # ... and many more

    def __str__(self) -> str:
        """Human-readable string representation."""

    @classmethod
    def from_http_status(cls, code: int) -> 'State':
        """Get State from HTTP status code."""
```

### COARequest

COA request structure.

```python
@dataclass
class COARequest(COABodyMixin):
    method: str = "GET"
    route: str = ""
    metadata: Optional[Dict[str, str]] = field(default_factory=dict)
    parameters: Optional[Dict[str, str]] = field(default_factory=dict)
    content_type: str = "application/json"
    body: str = ""  # Base64 encoded

    def set_body(self, data: Any, content_type: Optional[str] = None) -> None:
        """Set body data with content type."""

    def get_body(self) -> Any:
        """Get decoded body data."""

    def to_json_dict(self) -> Dict[str, Any]:
        """Convert to JSON-serializable dictionary."""
```

### COAResponse

COA response structure.

```python
@dataclass
class COAResponse(COABodyMixin):
    state: State = State.OK
    metadata: Optional[Dict[str, str]] = field(default_factory=dict)
    redirect_uri: Optional[str] = None
    content_type: str = "application/json"
    body: str = ""  # Base64 encoded

    def set_body(self, data: Any, content_type: Optional[str] = None) -> None:
        """Set body data with content type."""

    def get_body(self) -> Any:
        """Get decoded body data."""

    def to_json_dict(self) -> Dict[str, Any]:
        """Convert to JSON-serializable dictionary."""

    @classmethod
    def success(cls, data: Any = None, content_type: str = "application/json") -> 'COAResponse':
        """Create a success response."""

    @classmethod
    def error(cls, message: str, state: State = State.INTERNAL_ERROR,
              content_type: str = "application/json") -> 'COAResponse':
        """Create an error response."""

    @classmethod
    def not_found(cls, message: str = "Resource not found") -> 'COAResponse':
        """Create a not found response."""

    @classmethod
    def bad_request(cls, message: str = "Bad request") -> 'COAResponse':
        """Create a bad request response."""
```

---

## Summary Models

### SummaryState

State enumeration for summary operations.

```python
class SummaryState(IntEnum):
    PENDING = 0   # Currently unused
    RUNNING = 1   # Reconcile operation in progress
    DONE = 2      # Reconcile operation completed
```

### ComponentResultSpec

Result specification for a component.

```python
@dataclass
class ComponentResultSpec:
    status: State = State.OK
    message: str = ""

    def to_dict(self) -> Dict[str, any]:
        """Convert to dictionary."""

    @classmethod
    def from_dict(cls, data: Dict[str, any]) -> 'ComponentResultSpec':
        """Create from dictionary."""
```

### TargetResultSpec

Result specification for a target.

```python
@dataclass
class TargetResultSpec:
    status: str = "OK"
    message: str = ""
    component_results: Dict[str, ComponentResultSpec] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, any]:
        """Convert to dictionary."""

    @classmethod
    def from_dict(cls, data: Dict[str, any]) -> 'TargetResultSpec':
        """Create from dictionary."""
```

### SummarySpec

Deployment summary specification.

```python
@dataclass
class SummarySpec:
    target_count: int = 0
    success_count: int = 0
    planned_deployment: int = 0
    current_deployed: int = 0
    target_results: Dict[str, TargetResultSpec] = field(default_factory=dict)
    summary_message: str = ""
    job_id: str = ""
    skipped: bool = False
    is_removal: bool = False
    all_assigned_deployed: bool = False
    removed: bool = False

    def update_target_result(self, target: str, spec: TargetResultSpec) -> None:
        """Update target result, merging with existing."""

    def generate_status_message(self) -> str:
        """Generate detailed status message."""

    def to_dict(self) -> Dict[str, any]:
        """Convert to dictionary."""

    @classmethod
    def from_dict(cls, data: Dict[str, any]) -> 'SummarySpec':
        """Create from dictionary."""
```

### SummaryResult

Complete summary result for a deployment.

```python
@dataclass
class SummaryResult:
    summary: SummarySpec = field(default_factory=SummarySpec)
    summary_id: str = ""
    generation: str = ""
    time: datetime = field(default_factory=datetime.now)
    state: SummaryState = SummaryState.PENDING
    deployment_hash: str = ""

    def is_deployment_finished(self) -> bool:
        """Check if deployment is finished."""

    def to_dict(self) -> Dict[str, any]:
        """Convert to dictionary."""

    @classmethod
    def from_dict(cls, data: Dict[str, any]) -> 'SummaryResult':
        """Create from dictionary."""
```

---

## Utility Functions

### Serialization Functions

```python
def to_dict(obj: Any) -> Dict[str, Any]:
    """Convert dataclass object to dictionary."""

def from_dict(data: Dict[str, Any], cls: type) -> Any:
    """Convert dictionary to dataclass object."""

def serialize_components(components: List[ComponentSpec]) -> str:
    """Serialize components to JSON string."""

def deserialize_components(json_str: str) -> List[ComponentSpec]:
    """Deserialize JSON string to components list."""

def serialize_coa_request(coa_request: COARequest) -> str:
    """Serialize COARequest to JSON string."""

def deserialize_coa_request(json_str: str) -> COARequest:
    """Deserialize JSON string to COARequest."""

def serialize_coa_response(coa_response: COAResponse) -> str:
    """Serialize COAResponse to JSON string."""

def deserialize_coa_response(json_str: str) -> COAResponse:
    """Deserialize JSON string to COAResponse."""
```

### Summary Helper Functions

```python
def create_success_component_result(message: str = "") -> ComponentResultSpec:
    """Create a successful component result."""

def create_failed_component_result(message: str, status: State = None) -> ComponentResultSpec:
    """Create a failed component result."""

def create_target_result(status: str = "OK", message: str = "",
                        component_results: Dict[str, ComponentResultSpec] = None) -> TargetResultSpec:
    """Create a target result specification."""
```

---

## Exceptions

### SymphonyAPIError

Custom exception for Symphony API errors.

```python
class SymphonyAPIError(Exception):
    def __init__(self, message: str,
                 status_code: Optional[int] = None,
                 response_text: Optional[str] = None):
        """
        Initialize Symphony API error.

        Args:
            message: Error message
            status_code: HTTP status code if available
            response_text: Raw response text if available
        """
```

**Attributes:**
- `message`: Error message (from `str(exception)`)
- `status_code`: HTTP status code (if available)
- `response_text`: Raw response text (if available)

**Usage:**

```python
try:
    client.register_target("device", spec)
except SymphonyAPIError as e:
    print(f"Error: {e}")
    print(f"Status: {e.status_code}")
    print(f"Details: {e.response_text}")
```

---

## Constants

### COAConstants

Constants used in Symphony COA operations.

```python
class COAConstants:
    # Header constants
    COA_META_HEADER = "COA_META_HEADER"

    # Tracing and monitoring
    TRACING_EXPORTER_CONSOLE = "tracing.exporters.console"
    METRICS_EXPORTER_OTLP_GRPC = "metrics.exporters.otlpgrpc"
    # ... more constants

    # Provider constants
    PROVIDERS_PERSISTENT_STATE = "providers.persistentstate"
    PROVIDERS_VOLATILE_STATE = "providers.volatilestate"
    # ... more constants

    # Output constants
    STATUS_OUTPUT = "status"
    ERROR_OUTPUT = "error"
    STATE_OUTPUT = "__state"
```

---

## See Also

- [Quick Start Guide](QUICKSTART.md)
- [Examples](../examples/)
- [Main README](../README.md)
- [Symphony Documentation](https://github.com/eclipse-symphony/symphony)
