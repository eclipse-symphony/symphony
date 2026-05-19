#!/usr/bin/env python3
"""SolutionVersion and Instance Management Example

This example demonstrates how to:
1. Create a solutionversion
2. Create an instance from a solutionversion
3. Apply a deployment
4. Check deployment status
5. List solutionversions and instances
6. Clean up resources
"""

import time

import yaml

from symphony_sdk import SymphonyAPI, SymphonyAPIError


def create_solutionversion_yaml():
    """Create a sample solutionversion specification in YAML format."""
    solutionversion = {
        "displayName": "Web Application Stack",
        "rootResource": "web-app-stack",
        "metadata": {"version": "1.0.0", "description": "A simple web application with nginx"},
        "components": [
            {
                "name": "nginx-server",
                "type": "container",
                "properties": {"container.image": "nginx:1.21", "container.ports": "80:8080"},
                "metadata": {"description": "Nginx web server"},
            },
            {
                "name": "app-config",
                "type": "config",
                "properties": {
                    "config.type": "configmap",
                    "config.data": "server_name=example.com",
                },
            },
        ],
    }
    return yaml.dump(solutionversion)


def main():
    base_url = "http://localhost:8082/v1alpha2"
    username = "admin"
    password = ""

    solutionversion_name = "web-app-stack-v-1.0.0"  # Format: <rootResource>-v-<version>
    instance_name = "web-app-prod"

    with SymphonyAPI(base_url, username, password) as client:
        try:
            # 1. Create a solutionversion
            print("1. Creating solutionversion...")
            solutionversion_yaml = create_solutionversion_yaml()
            client.create_solutionversion(solutionversion_name, solutionversion_yaml)
            print(f"   ✓ SolutionVersion '{solutionversion_name}' created")

            # 2. Verify solutionversion was created
            print("\n2. Verifying solutionversion...")
            solutionversion = client.get_solutionversion(solutionversion_name)
            print("   ✓ SolutionVersion retrieved")
            spec = solutionversion.get("spec", {})
            print(f"     Display Name: {spec.get('displayName', 'N/A')}")
            print(f"     Components: {len(spec.get('components', []))}")

            # 3. Create an instance from the solutionversion
            print(f"\n3. Creating instance '{instance_name}'...")
            instance_spec = {
                "solutionversion": solutionversion_name,
                "target": {
                    "name": "example-device-001"  # Target must exist
                },
                "displayName": "Production Web App",
                "parameters": {"environment": "production", "replicas": "3"},
            }
            client.create_instance(instance_name, instance_spec)
            print(f"   ✓ Instance '{instance_name}' created")

            # 4. Check instance status
            print("\n4. Checking instance status...")
            for attempt in range(5):
                try:
                    status = client.get_instance_status(instance_name)
                    print(f"   Attempt {attempt + 1}/5:")
                    print(f"     Status: {status}")

                    # Check if deployment is complete
                    if status.get("status") == "Succeeded":
                        print("   ✓ Deployment completed successfully!")
                        break
                    elif status.get("status") == "Failed":
                        print("   ✗ Deployment failed!")
                        break

                    time.sleep(2)  # Wait before checking again
                except Exception as e:
                    print(f"     Status check error: {e}")
                    break

            # 5. List all solutionversions
            print("\n5. Listing all solutionversions...")
            solutionversions = client.list_solutionversions()
            solutionversions_list = (
                solutionversions if isinstance(solutionversions, list) else solutionversions.get("items", [])
            )
            print(f"   ✓ Found {len(solutionversions_list)} solutionversions")
            for sol in solutionversions_list[:5]:
                print(f"     - {sol.get('metadata', {}).get('name', 'unknown')}")

            # 6. List all instances
            print("\n6. Listing all instances...")
            instances = client.list_instances()
            instances_list = (
                instances if isinstance(instances, list) else instances.get("items", [])
            )
            print(f"   ✓ Found {len(instances_list)} instances")
            for inst in instances_list[:5]:
                print(f"     - {inst.get('metadata', {}).get('name', 'unknown')}")

            # 7. Clean up - delete instance and solutionversion
            print("\n7. Cleaning up resources...")
            print(f"   Deleting instance '{instance_name}'...")
            client.delete_instance(instance_name)
            print("   ✓ Instance deleted")

            print(f"   Deleting solutionversion '{solutionversion_name}'...")
            client.delete_solutionversion(solutionversion_name)
            print("   ✓ SolutionVersion deleted")

        except SymphonyAPIError as e:
            print(f"\n✗ Error: {e}")
            if e.status_code:
                print(f"  Status code: {e.status_code}")


def example_using_instance_spec_dataclass():
    """Example using InstanceSpec dataclass."""
    from symphony_sdk import InstanceSpec, TargetSelector

    print("\n" + "=" * 60)
    print("Using InstanceSpec Dataclass")
    print("=" * 60 + "\n")

    # Create instance spec using dataclass
    instance = InstanceSpec(
        name="my-instance",
        solutionversion="my-solutionversion",
        target=TargetSelector(name="my-target", selector={"location": "datacenter-1"}),
        scope="default",
        display_name="My Application Instance",
        parameters={"replicas": "3", "memory": "2Gi"},
        metadata={"owner": "platform-team"},
    )

    print("✓ Created InstanceSpec with dataclass")
    print(f"  Name: {instance.name}")
    print(f"  SolutionVersion: {instance.solutionversion}")
    print(f"  Target: {instance.target.name}")
    print(f"  Parameters: {instance.parameters}")


if __name__ == "__main__":
    print("=" * 60)
    print("Symphony SDK - SolutionVersion and Instance Management")
    print("=" * 60 + "\n")

    print("NOTE: Update the credentials in this script before running!\n")

    # Uncomment the lines below after updating credentials
    # main()
    # example_using_instance_spec_dataclass()

    print("Update base_url, username, and password in the script,")
    print("then uncomment the main() call to run this example.")
