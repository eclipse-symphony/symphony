#!/usr/bin/env python3
"""Summary and Status Tracking Example

This example demonstrates how to:
1. Create summary results for deployments
2. Track component deployment status
3. Generate status messages
4. Handle deployment completion
"""

from symphony_sdk import (
    State,
    SummaryResult,
    SummarySpec,
    SummaryState,
    create_failed_component_result,
    create_success_component_result,
    create_target_result,
)


def example_successful_deployment():
    """Example: Tracking a successful deployment."""
    print("1. Successful Deployment Example")
    print("-" * 40)

    # Create summary spec
    summary = SummarySpec(
        target_count=2,
        success_count=2,
        planned_deployment=4,
        current_deployed=4,
        all_assigned_deployed=True,
    )

    # Add successful target results
    target1_result = create_target_result(
        status="OK",
        message="All components deployed successfully",
        component_results={
            "web-server": create_success_component_result("Deployed"),
            "database": create_success_component_result("Deployed"),
        },
    )

    target2_result = create_target_result(
        status="OK",
        message="All components deployed successfully",
        component_results={
            "cache": create_success_component_result("Deployed"),
            "worker": create_success_component_result("Deployed"),
        },
    )

    summary.update_target_result("target-1", target1_result)
    summary.update_target_result("target-2", target2_result)

    # Create summary result
    result = SummaryResult(
        summary=summary,
        summary_id="deployment-001",
        generation="v1",
        state=SummaryState.DONE,
        deployment_hash="abc123",
    )

    print("✓ Deployment completed")
    print(f"  State: {result.state.name}")
    print(f"  Targets: {result.summary.target_count}")
    print(f"  Success: {result.summary.success_count}")
    print(f"  Deployed: {result.summary.current_deployed}/{result.summary.planned_deployment}")
    print(f"  All deployed: {result.summary.all_assigned_deployed}")

    if result.is_deployment_finished():
        print("\n✓ Deployment is finished!")

    # Show target details
    print("\n✓ Target results:")
    for target_name, target_result in result.summary.target_results.items():
        print(f"\n  {target_name}:")
        print(f"    Status: {target_result.status}")
        print(f"    Message: {target_result.message}")
        print("    Components:")
        for comp_name, comp_result in target_result.component_results.items():
            print(f"      - {comp_name}: {comp_result.status.name} ({comp_result.message})")


def example_failed_deployment():
    """Example: Tracking a failed deployment."""
    print("\n2. Failed Deployment Example")
    print("-" * 40)

    # Create summary spec for failed deployment
    summary = SummarySpec(
        target_count=2,
        success_count=1,  # Only 1 target succeeded
        planned_deployment=4,
        current_deployed=2,  # Only 2 components deployed
        all_assigned_deployed=False,
        summary_message="Some components failed to deploy",
    )

    # Target 1: Success
    target1_result = create_target_result(
        status="OK",
        message="All components deployed",
        component_results={
            "web-server": create_success_component_result("Deployed"),
            "database": create_success_component_result("Deployed"),
        },
    )

    # Target 2: Failure
    target2_result = create_target_result(
        status="Failed",
        message="Component deployment failed",
        component_results={
            "cache": create_failed_component_result(
                "Image pull failed: connection timeout", State.INTERNAL_ERROR
            ),
            "worker": create_failed_component_result(
                "Insufficient resources", State.INTERNAL_ERROR
            ),
        },
    )

    summary.update_target_result("target-1", target1_result)
    summary.update_target_result("target-2", target2_result)

    # Create summary result
    result = SummaryResult(summary=summary, summary_id="deployment-002", state=SummaryState.DONE)

    print("✓ Deployment completed with errors")
    print(f"  State: {result.state.name}")
    print(f"  Targets: {result.summary.target_count}")
    print(f"  Success: {result.summary.success_count}")
    print(f"  Deployed: {result.summary.current_deployed}/{result.summary.planned_deployment}")

    # Generate detailed status message
    status_message = result.summary.generate_status_message()
    print("\n✗ Status Message:")
    print(f"  {status_message}")

    # Show failed components
    print("\n✗ Failed components:")
    for target_name, target_result in result.summary.target_results.items():
        if target_result.status != "OK":
            print(f"\n  {target_name}:")
            for comp_name, comp_result in target_result.component_results.items():
                if comp_result.status != State.OK:
                    print(f"    - {comp_name}: {comp_result.message}")


def example_incremental_updates():
    """Example: Incrementally updating deployment status."""
    print("\n3. Incremental Status Updates Example")
    print("-" * 40)

    # Initialize summary
    summary = SummarySpec(target_count=1, planned_deployment=3)

    result = SummaryResult(
        summary=summary,
        summary_id="deployment-003",
        state=SummaryState.RUNNING,  # Still in progress
    )

    print("✓ Deployment started")
    print(f"  State: {result.state.name}")

    # Update 1: First component deployed
    print("\n  Update 1: web-server deployed")
    target_result = create_target_result(
        status="InProgress",
        component_results={"web-server": create_success_component_result("Deployed")},
    )
    result.summary.update_target_result("target-1", target_result)
    result.summary.current_deployed = 1
    print(f"    Deployed: {result.summary.current_deployed}/{result.summary.planned_deployment}")

    # Update 2: Second component deployed
    print("\n  Update 2: database deployed")
    target_result = create_target_result(
        status="InProgress",
        component_results={"database": create_success_component_result("Deployed")},
    )
    result.summary.update_target_result("target-1", target_result)
    result.summary.current_deployed = 2
    print(f"    Deployed: {result.summary.current_deployed}/{result.summary.planned_deployment}")

    # Update 3: Third component deployed - deployment complete
    print("\n  Update 3: cache deployed")
    target_result = create_target_result(
        status="OK",
        message="All components deployed",
        component_results={"cache": create_success_component_result("Deployed")},
    )
    result.summary.update_target_result("target-1", target_result)
    result.summary.current_deployed = 3
    result.summary.success_count = 1
    result.summary.all_assigned_deployed = True
    result.state = SummaryState.DONE
    print(f"    Deployed: {result.summary.current_deployed}/{result.summary.planned_deployment}")

    print("\n✓ Deployment completed!")
    print(f"  Final state: {result.state.name}")
    print(f"  All deployed: {result.summary.all_assigned_deployed}")


def example_serialization():
    """Example: Serializing and deserializing summary results."""
    print("\n4. Serialization Example")
    print("-" * 40)

    # Create a summary result
    summary = SummarySpec(
        target_count=1,
        success_count=1,
        planned_deployment=2,
        current_deployed=2,
        all_assigned_deployed=True,
    )

    target_result = create_target_result(
        status="OK", component_results={"app": create_success_component_result("Deployed")}
    )
    summary.update_target_result("target-1", target_result)

    result = SummaryResult(summary=summary, summary_id="deployment-004", state=SummaryState.DONE)

    print("✓ Created summary result")

    # Convert to dictionary
    result_dict = result.to_dict()
    print("\n✓ Converted to dictionary:")
    print(f"  Keys: {list(result_dict.keys())}")
    print(f"  Summary ID: {result_dict['summaryid']}")
    print(f"  State: {result_dict['state']}")

    # Restore from dictionary
    restored = SummaryResult.from_dict(result_dict)
    print("\n✓ Restored from dictionary:")
    print(f"  Summary ID: {restored.summary_id}")
    print(f"  State: {restored.state.name}")
    print(f"  Target count: {restored.summary.target_count}")
    print(f"  All deployed: {restored.summary.all_assigned_deployed}")


def example_removal_operation():
    """Example: Tracking a removal/uninstall operation."""
    print("\n5. Removal Operation Example")
    print("-" * 40)

    summary = SummarySpec(
        target_count=1,
        success_count=1,
        is_removal=True,  # This is a removal operation
        removed=True,
    )

    target_result = create_target_result(
        status="OK",
        message="Components removed successfully",
        component_results={
            "web-server": create_success_component_result("Removed"),
            "database": create_success_component_result("Removed"),
        },
    )
    summary.update_target_result("target-1", target_result)

    result = SummaryResult(summary=summary, state=SummaryState.DONE)

    print("✓ Removal operation completed")
    print(f"  Is removal: {result.summary.is_removal}")
    print(f"  Removed: {result.summary.removed}")
    print(f"  Target count: {result.summary.target_count}")
    print(f"  Success count: {result.summary.success_count}")


if __name__ == "__main__":
    print("=" * 60)
    print("Symphony SDK - Summary and Status Tracking")
    print("=" * 60 + "\n")

    example_successful_deployment()
    example_failed_deployment()
    example_incremental_updates()
    example_serialization()
    example_removal_operation()

    print("\n" + "=" * 60)
    print("All examples completed successfully!")
    print("=" * 60)
