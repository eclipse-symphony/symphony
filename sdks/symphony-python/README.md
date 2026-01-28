# Symphony SDK for Python

[![Python Version](https://img.shields.io/badge/python-3.9+-blue.svg)](https://www.python.org/downloads/)

`symphony-sdk` is a lightweight, dependency-minimal Python client for interacting with [Eclipse Symphony](https://github.com/eclipse-symphony/symphony). It bundles the REST API client together with the COA (Cloud Object API) data models, type definitions, and summary helpers used by Symphony control planes and providers.

## Features

**Modern REST Client**
- Automatic token management and session handling
- Context manager support for clean resource management
- Comprehensive error handling with `SymphonyAPIError`
- Configurable timeouts and logging

**Complete Data Models**
- COA request/response structures with multiple content types
- Target, Solution, Instance, and Deployment specifications
- Full type hints and dataclass support for IDE autocomplete
- Summary and status tracking models

**Production Ready**
- Dependency-minimal: only requires `requests`
- Comprehensive test suite
- Detailed documentation and examples
- Type-safe with Python dataclasses

## Installation

Install from source:

```bash
git clone https://github.com/eclipse-symphony/symphony.git
cd symphony/sdks/symphony-python
pip install -e .
```

## Quick Start

### Basic Client Usage

```python
from symphony_sdk import SymphonyAPI

# Use context manager for automatic cleanup
with SymphonyAPI(
    base_url="http://localhost:8082",
    username="admin",
    password=""
) as client:
    # Authentication happens automatically
    if client.health_check():
        print("Connected to Symphony!")

    # List all targets
    targets = client.list_targets()
    print(f"Found {len(targets.get('items', []))} targets")
```

### Working with Targets

```python
from symphony_sdk import SymphonyAPI

with SymphonyAPI(base_url, username, password) as client:
    # Register a new target
    target_spec = {
        "displayName": "Edge Gateway",
        "properties": {
            "location": "datacenter-1",
            "os": "linux"
        }
    }
    client.register_target("gateway-001", target_spec)

    # Send heartbeat
    client.ping_target("gateway-001")

    # Unregister when done
    client.unregister_target("gateway-001")
```

### Deploying Solutions

```python
import yaml
from symphony_sdk import SymphonyAPI

with SymphonyAPI(base_url, username, password) as client:
    # Create a solution
    solution = {
        "displayName": "Web App",
        "components": [{
            "name": "nginx",
            "type": "container",
            "properties": {"image": "nginx:latest"}
        }]
    }
    client.create_solution("web-app", yaml.dump(solution))

    # Deploy an instance
    instance_spec = {
        "solution": "web-app",
        "target": {"name": "gateway-001"}
    }
    client.create_instance("web-app-prod", instance_spec)

    # Check status
    status = client.get_instance_status("web-app-prod")
    print(f"Status: {status.get('status')}")
```

### Working with COA

```python
from symphony_sdk import COARequest, COAResponse, State

# Create a COA request
request = COARequest(
    method="GET",
    route="/components/list",
    content_type="application/json"
)
request.set_body({"filter": "active"})

# Create responses
success = COAResponse.success({"components": ["nginx", "redis"]})
error = COAResponse.error("Not found", State.NOT_FOUND)

# Get decoded body
data = success.get_body()
print(data)  # {'components': ['nginx', 'redis']}
```

### Using Type-Safe Dataclasses

```python
from symphony_sdk import (
    TargetSpec, ComponentSpec, InstanceSpec,
    TargetSelector, ObjectMeta
)

# Create strongly-typed specifications
metadata = ObjectMeta(
    name="my-target",
    namespace="default",
    labels={"env": "prod"}
)

component = ComponentSpec(
    name="nginx",
    type="container",
    properties={"image": "nginx:latest"}
)

target = TargetSpec(
    displayName="Production Gateway",
    components=[component],
    metadata={"owner": "platform-team"}
)

instance = InstanceSpec(
    name="my-app",
    solution="web-app",
    target=TargetSelector(name="my-target"),
    parameters={"replicas": "3"}
)
```

## Documentation

📚 **Complete Documentation:**
- [Quick Start Guide](docs/QUICKSTART.md) - Get up and running quickly
- [Glossary](docs/GLOSSARY.md) - Commonly used terms and their explanations
- [API Reference](docs/API.md) - API documentation
- [Examples](examples/) - Practical usage examples

📖 **Examples:**
- [Basic Client Usage](examples/01_basic_client.py) - Authentication and health checks
- [Target Management](examples/02_target_management.py) - Register and manage targets
- [Solution Deployment](examples/03_solution_deployment.py) - Deploy solutions and instances
- [COA Provider](examples/04_coa_provider.py) - Work with COA requests/responses
- [Summary Tracking](examples/05_summary_tracking.py) - Track deployment status

## Key Components

### SymphonyAPI Client

The main client for interacting with Symphony:

```python
client = SymphonyAPI(
    base_url="https://symphony.example.com",
    username="user",
    password="pass",
    timeout=30.0,  # Optional: custom timeout
    logger=my_logger  # Optional: custom logger
)
```

**Methods:**
- Target Management: `register_target()`, `unregister_target()`, `list_targets()`, `get_target()`, `ping_target()`
- Solution Management: `create_solution()`, `get_solution()`, `delete_solution()`, `list_solutions()`
- Instance Management: `create_instance()`, `get_instance()`, `delete_instance()`, `list_instances()`
- Deployment: `apply_deployment()`, `reconcile_solution()`, `get_instance_status()`
- Utilities: `authenticate()`, `health_check()`,

### Data Models

Comprehensive data models for Symphony resources:

- **Target Models**: `TargetSpec`, `TargetState`, `TargetStatus`, `TargetSelector`
- **Solution Models**: `SolutionSpec`, `SolutionState`, `ComponentSpec`, `RouteSpec`
- **Instance Models**: `InstanceSpec`, `DeploymentSpec`, `TopologySpec`, `PipelineSpec`
- **COA Models**: `COARequest`, `COAResponse`, `State`
- **Summary Models**: `SummaryResult`, `SummarySpec`, `ComponentResultSpec`, `TargetResultSpec`
- **Metadata**: `ObjectMeta`, `BindingSpec`, `DeviceSpec`

All models are Python dataclasses with full type hints.

### Error Handling

```python
from symphony_sdk import SymphonyAPIError

try:
    client.register_target("device-001", spec)
except SymphonyAPIError as e:
    print(f"Error: {e}")
    print(f"Status code: {e.status_code}")
    print(f"Response: {e.response_text}")
```

## Development

### Setup

```bash
# Clone the repository
git clone https://github.com/eclipse-symphony/symphony.git
cd symphony/sdks/symphony-python
```

#### Using pip

```bash
# Install in development mode with all development dependencies
pip install -e ".[dev]"
```

#### Using uv

```bash
# Install the package and dev dependencies
uv pip install -e ".[dev]"
```

This installs the package along with:
- `pytest` and `pytest-cov` - Testing and coverage
- `ruff` - Fast linter and formatter
- `mypy` - Static type checker
- `types-PyYAML` and `types-requests` - Type stubs for dependencies

### Running Tests

#### Using pip

```bash
# Run all tests
pytest

# Run with coverage (configured in pyproject.toml)
pytest --cov=src/symphony_sdk --cov-report=html --cov-report=term-missing

# Run specific test file
pytest tests/test_api_client.py

# Run tests in verbose mode
pytest -v
```

#### Using uv

```bash
# Run all tests
uv run pytest

# Run with coverage (configured in pyproject.toml)
uv run pytest --cov=src/symphony_sdk --cov-report=html --cov-report=term-missing

# Run specific test file
uv run pytest tests/test_api_client.py

# Run tests in verbose mode
uv run pytest -v
```

### Code Quality Tools

The project is configured with modern Python tooling for maintaining code quality:

#### Linting and Formatting with Ruff

```bash
# Check code for issues
ruff check .

# Check code and auto-fix issues
ruff check --fix .

# Format code
ruff format .

# Check formatting without making changes
ruff format --check .
```

Ruff is configured in `pyproject.toml` to enforce:
- PEP 8 style guidelines
- Import sorting (isort-compatible)
- Google-style docstrings
- Modern Python idioms
- Common bug patterns

#### Type Checking with Mypy

```bash
# Type check the entire codebase
mypy src/

# Type check with verbose output
mypy --show-error-codes src/

# Type check a specific file
mypy src/symphony_sdk/api_client.py
```

The codebase has comprehensive type hints (77%+ coverage) and is configured for strict type checking in `pyproject.toml`.

### Development Workflow

Recommended workflow before committing:

```bash
# 1. Format code
ruff format .

# 2. Fix linting issues
ruff check --fix .

# 3. Run type checker
mypy src/

# 4. Run tests with coverage
pytest

# 5. Review coverage report
open htmlcov/index.html  # or xdg-open on Linux
```

### Pre-commit Integration (Optional)

To run checks automatically before each commit, install pre-commit hooks:

```bash
# Install pre-commit
pip install pre-commit

# Install the git hooks
pre-commit install
```

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
        args: [--fix]
      - id: ruff-format
  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.0.0
    hooks:
      - id: mypy
        additional_dependencies: [types-PyYAML, types-requests]
```

### Project Structure

```
symphony-python/
├── src/symphony_sdk/
│   ├── __init__.py          # Public API exports
│   ├── api_client.py        # REST API client
│   ├── models.py            # Data models and COA structures
│   ├── types.py             # State enum and constants
│   └── summary.py           # Summary tracking models
├── tests/                   # Unit tests
├── examples/                # Usage examples
├── docs/                    # Documentation
├── README.md
├── CHANGELOG.md
└── pyproject.toml          # Package configuration
```

## Requirements

- Python 3.9 or higher
- `requests` library (automatically installed)

Optional dependencies:
- `paho-mqtt>=2.0` (for MQTT support)

## Contributing

Contributions are welcome! Please follow the Symphony project contribution guidelines:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## Links

- [Eclipse Symphony](https://github.com/eclipse-symphony/symphony) - Main Symphony repository
- [Symphony Documentation](https://github.com/eclipse-symphony/symphony/tree/main/docs) - Architecture and concepts
- [Issue Tracker](https://github.com/eclipse-symphony/symphony/issues) - Report bugs or request features

## Support

- 📖 Read the [documentation](docs/)
- 💬 Check existing [issues](https://github.com/eclipse-symphony/symphony/issues)
- 🐛 Report bugs in the [issue tracker](https://github.com/eclipse-symphony/symphony/issues/new)
- 💡 See [examples](examples/) for common patterns
