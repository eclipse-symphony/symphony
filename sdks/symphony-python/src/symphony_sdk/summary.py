"""Symphony API Summary Models - Python Translation

Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
SPDX-License-Identifier: MIT

This module provides Python translations of the Symphony API summary models
from the original Go implementation.
"""

from dataclasses import dataclass, field
from datetime import datetime
from enum import IntEnum
from typing import Dict

from symphony_sdk.types import State


class SummaryState(IntEnum):
    """State enumeration for Symphony summary operations."""

    PENDING = 0  # Currently unused
    RUNNING = 1  # Indicates that a reconcile operation is in progress
    DONE = 2  # Indicates that a reconcile operation has completed (successfully or unsuccessfully)


@dataclass
class ComponentResultSpec:
    """Result specification for a single component operation.

    Attributes:
        status: State indicating success/failure of the component operation
        message: Optional message with details about the operation
    """

    status: State = State.OK
    message: str = ""

    def to_dict(self) -> Dict[str, any]:
        """Convert to dictionary representation."""
        return {"status": self.status.value, "message": self.message}

    @classmethod
    def from_dict(cls, data: Dict[str, any]) -> "ComponentResultSpec":
        """Create instance from dictionary."""
        return cls(
            status=State(data.get("status", State.OK.value)), message=data.get("message", "")
        )


@dataclass
class TargetResultSpec:
    """Result specification for a target containing multiple components.

    Attributes:
        status: Overall status string for the target
        message: Optional message with target-level details
        component_results: Map of component name to component result
    """

    status: str = "OK"
    message: str = ""
    component_results: Dict[str, ComponentResultSpec] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, any]:
        """Convert to dictionary representation."""
        result = {"status": self.status}
        if self.message:
            result["message"] = self.message
        if self.component_results:
            result["components"] = {
                name: comp_result.to_dict() for name, comp_result in self.component_results.items()
            }
        return result

    @classmethod
    def from_dict(cls, data: Dict[str, any]) -> "TargetResultSpec":
        """Create instance from dictionary."""
        component_results = {}
        if "components" in data:
            component_results = {
                name: ComponentResultSpec.from_dict(comp_data)
                for name, comp_data in data["components"].items()
            }

        return cls(
            status=data.get("status", "OK"),
            message=data.get("message", ""),
            component_results=component_results,
        )


@dataclass
class SummarySpec:
    """Specification for deployment summary containing target and component results.

    Attributes:
        target_count: Total number of targets
        success_count: Number of successful deployments
        planned_deployment: Number of planned deployments
        current_deployed: Number of currently deployed components
        target_results: Map of target name to target result
        summary_message: Overall summary message
        job_id: Optional job identifier
        skipped: Whether the deployment was skipped
        is_removal: Whether this is a removal operation
        all_assigned_deployed: Whether all assigned components are deployed
        removed: Whether components were removed
    """

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
        """Update target result, merging with existing result if present.

        Args:
            target: Target name
            spec: New target result specification
        """
        if target not in self.target_results:
            self.target_results[target] = spec
        else:
            existing = self.target_results[target]

            # Update status - use new status if it's not "OK"
            status = existing.status
            if spec.status != "OK":
                status = spec.status

            # Merge messages
            message = existing.message
            if spec.message:
                if message:
                    message += "; "
                message += spec.message

            # Merge component results
            merged_components = existing.component_results.copy()
            merged_components.update(spec.component_results)

            # Update the existing result
            existing.status = status
            existing.message = message
            existing.component_results = merged_components

            self.target_results[target] = existing

    def generate_status_message(self) -> str:
        """Generate a detailed status message from target and component results.

        Returns:
            Formatted status message with error details
        """
        if self.all_assigned_deployed:
            return ""

        error_message = "Failed to deploy"
        if self.summary_message:
            error_message += f": {self.summary_message}"
        error_message += ". "

        # Get target names and sort them for consistent output
        target_names = sorted(self.target_results.keys())

        # Build target errors in sorted order
        target_errors = []
        for target in target_names:
            result = self.target_results[target]
            target_error = f'{target}: "{result.message}"'

            # Get component names and sort them for consistency
            component_names = sorted(result.component_results.keys())

            # Add component results in sorted order
            for component in component_names:
                component_result = result.component_results[component]
                target_error += f" ({target}.{component}: {component_result.message})"

            target_errors.append(target_error)

        return error_message + f"Detailed status: {', '.join(target_errors)}"

    def to_dict(self) -> Dict[str, any]:
        """Convert to dictionary representation."""
        result = {
            "targetCount": self.target_count,
            "successCount": self.success_count,
            "plannedDeployment": self.planned_deployment,
            "currentDeployed": self.current_deployed,
            "skipped": self.skipped,
            "isRemoval": self.is_removal,
            "allAssignedDeployed": self.all_assigned_deployed,
            "removed": self.removed,
        }

        if self.target_results:
            result["targets"] = {
                name: target_result.to_dict() for name, target_result in self.target_results.items()
            }
        if self.summary_message:
            result["message"] = self.summary_message
        if self.job_id:
            result["jobID"] = self.job_id

        return result

    @classmethod
    def from_dict(cls, data: Dict[str, any]) -> "SummarySpec":
        """Create instance from dictionary."""
        target_results = {}
        if "targets" in data:
            target_results = {
                name: TargetResultSpec.from_dict(target_data)
                for name, target_data in data["targets"].items()
            }

        return cls(
            target_count=data.get("targetCount", 0),
            success_count=data.get("successCount", 0),
            planned_deployment=data.get("plannedDeployment", 0),
            current_deployed=data.get("currentDeployed", 0),
            target_results=target_results,
            summary_message=data.get("message", ""),
            job_id=data.get("jobID", ""),
            skipped=data.get("skipped", False),
            is_removal=data.get("isRemoval", False),
            all_assigned_deployed=data.get("allAssignedDeployed", False),
            removed=data.get("removed", False),
        )


@dataclass
class SummaryResult:
    """Complete summary result for a deployment operation.

    Attributes:
        summary: The summary specification with all results
        summary_id: Optional unique identifier for the summary
        generation: Generation string for versioning
        time: Timestamp when the summary was created
        state: Current state of the summary operation
        deployment_hash: Hash of the deployment configuration
    """

    summary: SummarySpec = field(default_factory=SummarySpec)
    summary_id: str = ""
    generation: str = ""
    time: datetime = field(default_factory=datetime.now)
    state: SummaryState = SummaryState.PENDING
    deployment_hash: str = ""

    def is_deployment_finished(self) -> bool:
        """Check if the deployment operation has finished.

        Returns:
            True if deployment is done (successfully or unsuccessfully)
        """
        return self.state == SummaryState.DONE

    def to_dict(self) -> Dict[str, any]:
        """Convert to dictionary representation."""
        return {
            "summary": self.summary.to_dict(),
            "summaryid": self.summary_id,
            "generation": self.generation,
            "time": self.time.isoformat(),
            "state": self.state.value,
            "deploymentHash": self.deployment_hash,
        }

    @classmethod
    def from_dict(cls, data: Dict[str, any]) -> "SummaryResult":
        """Create instance from dictionary."""
        # Parse time string
        time_obj = datetime.now()
        if "time" in data:
            try:
                time_obj = datetime.fromisoformat(data["time"])
            except (ValueError, TypeError):
                pass

        return cls(
            summary=SummarySpec.from_dict(data.get("summary", {})),
            summary_id=data.get("summaryid", ""),
            generation=data.get("generation", ""),
            time=time_obj,
            state=SummaryState(data.get("state", SummaryState.PENDING.value)),
            deployment_hash=data.get("deploymentHash", ""),
        )


# Helper functions for creating common result types
def create_success_component_result(message: str = "") -> ComponentResultSpec:
    """Create a successful component result."""
    return ComponentResultSpec(status=State.OK, message=message)


def create_failed_component_result(message: str, status: State = None) -> ComponentResultSpec:
    """Create a failed component result."""
    if status is None:
        status = State.INTERNAL_ERROR
    return ComponentResultSpec(status=status, message=message)


def create_target_result(
    status: str = "OK", message: str = "", component_results: Dict[str, ComponentResultSpec] = None
) -> TargetResultSpec:
    """Create a target result specification."""
    if component_results is None:
        component_results = {}
    return TargetResultSpec(status=status, message=message, component_results=component_results)


# Export commonly used items
__all__ = [
    "SummaryState",
    "ComponentResultSpec",
    "TargetResultSpec",
    "SummarySpec",
    "SummaryResult",
    "create_success_component_result",
    "create_failed_component_result",
    "create_target_result",
]
