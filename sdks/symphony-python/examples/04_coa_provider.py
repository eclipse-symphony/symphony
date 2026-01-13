#!/usr/bin/env python3
"""COA (Cloud Object API) Provider Example

This example demonstrates how to:
1. Create COA requests and responses
2. Handle different content types (JSON, text, binary)
3. Work with COA body encoding/decoding
4. Build a simple provider that responds to COA requests
"""

import json

from symphony_sdk import (
    COARequest,
    COAResponse,
    State,
    deserialize_coa_request,
    serialize_coa_request,
)


def example_json_request():
    """Example: Creating and handling JSON COA requests."""
    print("1. JSON COA Request Example")
    print("-" * 40)

    # Create a COA request with JSON body
    request = COARequest(
        method="GET",
        route="/components/list",
        content_type="application/json",
        parameters={"namespace": "default"},
        metadata={"request-id": "123"},
    )

    # Set JSON body data
    request.set_body({"filter": "active", "limit": 10})

    print("✓ Created COA request")
    print(f"  Method: {request.method}")
    print(f"  Route: {request.route}")
    print(f"  Content-Type: {request.content_type}")

    # Serialize to JSON string
    request_json = serialize_coa_request(request)
    print("\n✓ Serialized request:")
    print(f"  {request_json[:150]}...")

    # Deserialize back
    restored_request = deserialize_coa_request(request_json)
    restored_body = restored_request.get_body()
    print("\n✓ Deserialized and decoded body:")
    print(f"  {restored_body}")


def example_json_response():
    """Example: Creating JSON COA responses."""
    print("\n2. JSON COA Response Example")
    print("-" * 40)

    # Create a success response
    response = COAResponse.success(
        data={
            "components": [
                {"name": "web-server", "status": "running"},
                {"name": "database", "status": "running"},
            ],
            "total": 2,
        }
    )

    print("✓ Created success response")
    print(f"  State: {response.state} ({response.state.name})")

    # Get body data
    body = response.get_body()
    print(f"  Body: {body}")

    # Create error responses
    error_response = COAResponse.error("Component not found", state=State.NOT_FOUND)
    print("\n✓ Created error response")
    print(f"  State: {error_response.state} ({error_response.state.name})")
    print(f"  Body: {error_response.get_body()}")

    # Create bad request response
    bad_request = COAResponse.bad_request("Invalid component name format")
    print("\n✓ Created bad request response")
    print(f"  State: {bad_request.state}")


def example_text_content():
    """Example: Working with plain text content."""
    print("\n3. Plain Text Content Example")
    print("-" * 40)

    # Create request with text content
    request = COARequest(method="POST", route="/logs/append", content_type="text/plain")

    log_message = "Application started successfully at 2024-01-01 10:00:00"
    request.set_body(log_message)

    print("✓ Created text request")
    print(f"  Original text: {log_message}")
    print(f"  Stored body: {request.body[:50]}...")

    # Retrieve text
    retrieved_text = request.get_body()
    print(f"  Retrieved text: {retrieved_text}")

    # Create text response
    response = COAResponse(content_type="text/plain", state=State.OK)
    response.set_body("Log entry appended successfully")

    print("\n✓ Created text response")
    print(f"  Response: {response.get_body()}")


def example_binary_content():
    """Example: Working with binary content."""
    print("\n4. Binary Content Example")
    print("-" * 40)

    # Create request with binary content
    request = COARequest(
        method="POST", route="/files/upload", content_type="application/octet-stream"
    )

    # Simulate binary data
    binary_data = b"\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"
    request.set_body(binary_data)

    print("✓ Created binary request")
    print(f"  Original bytes: {binary_data[:20]}")
    print(f"  Stored body (base64): {request.body[:50]}...")

    # Retrieve binary data
    retrieved_binary = request.get_body()
    print(f"  Retrieved bytes: {retrieved_binary[:20]}")
    print(f"  Match: {binary_data == retrieved_binary}")


def example_provider_simulation():
    """Example: Simulating a COA provider."""
    print("\n5. Provider Simulation Example")
    print("-" * 40)

    def handle_component_get(request: COARequest) -> COAResponse:
        """Handle component GET requests."""
        body = request.get_body()
        component_name = body.get("name", "")

        if not component_name:
            return COAResponse.bad_request("Component name is required")

        # Simulate component lookup
        if component_name == "web-server":
            return COAResponse.success(
                {
                    "name": "web-server",
                    "status": "running",
                    "properties": {"image": "nginx:latest", "port": "80"},
                }
            )
        else:
            return COAResponse.not_found(f"Component '{component_name}' not found")

    # Simulate incoming request
    incoming = COARequest(method="GET", route="/components/get", content_type="application/json")
    incoming.set_body({"name": "web-server"})

    print("✓ Received request")
    print(f"  Route: {incoming.route}")
    print(f"  Body: {incoming.get_body()}")

    # Handle request
    response = handle_component_get(incoming)

    print("\n✓ Generated response")
    print(f"  State: {response.state} ({response.state.name})")
    print(f"  Body: {json.dumps(response.get_body(), indent=2)}")

    # Test with missing component
    print("\n" + "-" * 40)
    incoming2 = COARequest(method="GET", route="/components/get")
    incoming2.set_body({"name": "unknown-component"})

    response2 = handle_component_get(incoming2)
    print("✓ Response for missing component:")
    print(f"  State: {response2.state} ({response2.state.name})")
    print(f"  Body: {response2.get_body()}")


def example_deployment_spec():
    """Example: Working with DeploymentSpec."""
    print("\n6. DeploymentSpec Example")
    print("-" * 40)

    from symphony_sdk import (
        ComponentSpec,
        DeploymentSpec,
        ObjectMeta,
        SolutionSpec,
        SolutionState,
        to_dict,
    )

    # Create a deployment spec
    component1 = ComponentSpec(
        name="frontend", type="container", properties={"image": "nginx:latest"}
    )

    component2 = ComponentSpec(name="backend", type="container", properties={"image": "node:18"})

    solution = SolutionState(
        metadata=ObjectMeta(name="my-app"),
        spec=SolutionSpec(components=[component1, component2], displayName="My Application"),
    )

    deployment = DeploymentSpec(solutionName="my-app", solution=solution, activeTarget="target-001")

    print("✓ Created deployment spec")
    print(f"  Solution: {deployment.solutionName}")
    print(f"  Components: {len(deployment.solution.spec.components)}")

    # Get component slice
    components = deployment.get_components_slice()
    print("\n✓ Component slice:")
    for comp in components:
        print(f"  - {comp.name} ({comp.type})")

    # Convert to dictionary
    deployment_dict = to_dict(deployment)
    print(f"\n✓ Converted to dict (keys): {list(deployment_dict.keys())}")


if __name__ == "__main__":
    print("=" * 60)
    print("Symphony SDK - COA Provider Examples")
    print("=" * 60 + "\n")

    example_json_request()
    example_json_response()
    example_text_content()
    example_binary_content()
    example_provider_simulation()
    example_deployment_spec()

    print("\n" + "=" * 60)
    print("All examples completed successfully!")
    print("=" * 60)
