#!/usr/bin/env python3
"""Target Management Example

This example demonstrates how to:
1. Register a new target
2. List all targets
3. Get target details
4. Update target status
5. Ping a target
6. Unregister a target
"""

from symphony_sdk import SymphonyAPI, SymphonyAPIError


def create_target_spec():
    """Create a sample target specification."""
    return {
        "displayName": "Example IoT Device",
        "scope": "default",
        "properties": {
            "location": "datacenter-1",
            "os": "linux",
            "arch": "amd64",
            "device-type": "edge-gateway",
        },
        "components": [],
        "metadata": {"owner": "infrastructure-team", "environment": "production"},
    }


def main():
    # Initialize client
    base_url = "http://localhost:8082/v1alpha2"
    username = "admin"
    password = ""

    with SymphonyAPI(base_url, username, password) as client:
        target_name = "example-device-001"

        try:
            # 1. Register a new target
            print(f"1. Registering target '{target_name}'...")
            target_spec = create_target_spec()
            result = client.register_target(target_name, target_spec)
            print("   ✓ Target registered successfully")

            # 2. List all targets
            print("\n2. Listing all targets...")
            targets = client.list_targets()
            print(f"   ✓ Found {len(targets)} targets")
            for target in targets[:5]:  # Show first 5
                print(f"     - {target.get('metadata', {}).get('name', 'unknown')}")

            # 3. Get specific target details
            print(f"\n3. Getting details for '{target_name}'...")
            target_details = client.get_target(target_name)
            print("   ✓ Target details retrieved", target_details)

            # 4. Ping the target (heartbeat)
            print(f"\n4. Sending heartbeat to '{target_name}'...")
            ping_result = client.ping_target(target_name)
            print(f"   ✓ Heartbeat sent successfully - Response: {ping_result}")

            # 5. Update target status
            print("\n5. Updating target status...")
            status_data = {
                "properties": {"health": "healthy", "last_check": "2024-01-01T00:00:00Z"}
            }
            client.update_target_status(target_name, status_data)
            print("   ✓ Target status updated")

            # 6. Unregister the target
            print(f"\n6. Unregistering target '{target_name}'...")
            client.unregister_target(target_name)
            print("   ✓ Target unregistered successfully")

        except SymphonyAPIError as e:
            print(f"\n✗ Error: {e}")
            if e.status_code:
                print(f"  Status code: {e.status_code}")
            if e.response_text:
                print(f"  Response: {e.response_text[:200]}")


def example_using_dataclasses():
    """Example using Symphony SDK dataclasses."""
    from symphony_sdk import ComponentSpec, ObjectMeta, TargetSpec, TargetState

    print("\n" + "=" * 60)
    print("Using Symphony SDK Dataclasses")
    print("=" * 60 + "\n")

    # Create target spec using dataclasses
    metadata = ObjectMeta(
        name="example-device-002",
        namespace="default",
        labels={"env": "prod", "region": "us-west"},
        annotations={"description": "Production edge gateway"},
    )

    # Create component specs
    component = ComponentSpec(
        name="web-app", type="container", properties={"image": "nginx:latest", "port": "80"}
    )

    # Create target spec
    target_spec = TargetSpec(
        displayName="Example Device 002",
        scope="default",
        properties={"location": "datacenter-2", "os": "linux"},
        components=[component],
        metadata={"owner": "dev-team"},
    )

    # Create target state
    target_state = TargetState(metadata=metadata, spec=target_spec)

    print("✓ Created target state with dataclasses")
    print(f"  Name: {target_state.metadata.name}")
    print(f"  Components: {len(target_state.spec.components)}")
    print(f"  First component: {target_state.spec.components[0].name}")


if __name__ == "__main__":
    print("=" * 60)
    print("Symphony SDK - Target Management")
    print("=" * 60 + "\n")

    print("NOTE: Update the credentials in this script before running!\n")

    # Uncomment the lines below after updating credentials
    # main()
    # example_using_dataclasses()

    print("Update base_url, username, and password in the script,")
    print("then uncomment the main() and example_using_dataclasses() call to run this example.")
