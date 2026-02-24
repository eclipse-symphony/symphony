#!/usr/bin/env python3
"""
Comprehensive unit tests for Symphony SDK summary models.
"""

import unittest

from symphony_sdk.summary import (
    ComponentResultSpec,
    SummaryResult,
    SummarySpec,
    SummaryState,
    TargetResultSpec,
    create_failed_component_result,
    create_success_component_result,
    create_target_result,
)
from symphony_sdk.types import State


class TestSummaryState(unittest.TestCase):
    """Test cases for SummaryState enumeration."""

    def test_summary_state_values(self):
        """Test SummaryState enum values."""
        self.assertEqual(SummaryState.PENDING, 0)
        self.assertEqual(SummaryState.RUNNING, 1)
        self.assertEqual(SummaryState.DONE, 2)

    def test_summary_state_from_int(self):
        """Test creating SummaryState from integer."""
        self.assertEqual(SummaryState(0), SummaryState.PENDING)
        self.assertEqual(SummaryState(1), SummaryState.RUNNING)
        self.assertEqual(SummaryState(2), SummaryState.DONE)


class TestComponentResultSpec(unittest.TestCase):
    """Test cases for ComponentResultSpec."""

    def test_component_result_spec_creation(self):
        """Test ComponentResultSpec creation with default values."""
        result = ComponentResultSpec()
        self.assertEqual(result.status, State.OK)
        self.assertEqual(result.message, "")

    def test_component_result_spec_with_values(self):
        """Test ComponentResultSpec creation with specific values."""
        result = ComponentResultSpec(status=State.UPDATE_FAILED, message="Update operation failed")
        self.assertEqual(result.status, State.UPDATE_FAILED)
        self.assertEqual(result.message, "Update operation failed")

    def test_component_result_spec_to_dict(self):
        """Test ComponentResultSpec to_dict conversion."""
        result = ComponentResultSpec(status=State.UPDATED, message="Component updated successfully")

        expected_dict = {"status": State.UPDATED.value, "message": "Component updated successfully"}

        self.assertEqual(result.to_dict(), expected_dict)

    def test_component_result_spec_from_dict(self):
        """Test ComponentResultSpec from_dict conversion."""
        data = {"status": State.DELETE_FAILED.value, "message": "Failed to delete component"}

        result = ComponentResultSpec.from_dict(data)
        self.assertEqual(result.status, State.DELETE_FAILED)
        self.assertEqual(result.message, "Failed to delete component")

    def test_component_result_spec_from_dict_defaults(self):
        """Test ComponentResultSpec from_dict with missing values."""
        data = {}
        result = ComponentResultSpec.from_dict(data)
        self.assertEqual(result.status, State.OK)
        self.assertEqual(result.message, "")

    def test_component_result_spec_from_dict_partial(self):
        """Test ComponentResultSpec from_dict with partial data."""
        data = {"status": State.UPDATED.value}
        result = ComponentResultSpec.from_dict(data)
        self.assertEqual(result.status, State.UPDATED)
        self.assertEqual(result.message, "")


class TestTargetResultSpec(unittest.TestCase):
    """Test cases for TargetResultSpec."""

    def test_target_result_spec_creation(self):
        """Test TargetResultSpec creation with defaults."""
        result = TargetResultSpec()
        self.assertEqual(result.status, "OK")
        self.assertEqual(result.message, "")
        self.assertEqual(result.component_results, {})

    def test_target_result_spec_with_values(self):
        """Test TargetResultSpec creation with specific values."""
        comp1 = ComponentResultSpec(State.UPDATED, "Component 1 updated")
        comp2 = ComponentResultSpec(State.UPDATE_FAILED, "Component 2 failed")

        result = TargetResultSpec(
            status="PARTIAL_SUCCESS",
            message="Some components failed",
            component_results={"comp1": comp1, "comp2": comp2},
        )

        self.assertEqual(result.status, "PARTIAL_SUCCESS")
        self.assertEqual(result.message, "Some components failed")
        self.assertEqual(len(result.component_results), 2)
        self.assertEqual(result.component_results["comp1"], comp1)
        self.assertEqual(result.component_results["comp2"], comp2)

    def test_target_result_spec_to_dict(self):
        """Test TargetResultSpec to_dict conversion."""
        comp_result = ComponentResultSpec(State.UPDATED, "Success")
        result = TargetResultSpec(
            status="SUCCESS",
            message="All components updated",
            component_results={"comp1": comp_result},
        )

        result_dict = result.to_dict()

        expected_dict = {
            "status": "SUCCESS",
            "message": "All components updated",
            "components": {"comp1": {"status": State.UPDATED.value, "message": "Success"}},
        }

        self.assertEqual(result_dict, expected_dict)

    def test_target_result_spec_to_dict_minimal(self):
        """Test TargetResultSpec to_dict with minimal data."""
        result = TargetResultSpec(status="OK")
        result_dict = result.to_dict()

        expected_dict = {"status": "OK"}
        self.assertEqual(result_dict, expected_dict)

    def test_target_result_spec_from_dict(self):
        """Test TargetResultSpec from_dict conversion."""
        data = {
            "status": "FAILED",
            "message": "Operation failed",
            "components": {
                "comp1": {"status": State.UPDATE_FAILED.value, "message": "Component update failed"}
            },
        }

        result = TargetResultSpec.from_dict(data)

        self.assertEqual(result.status, "FAILED")
        self.assertEqual(result.message, "Operation failed")
        self.assertEqual(len(result.component_results), 1)

        comp_result = result.component_results["comp1"]
        self.assertEqual(comp_result.status, State.UPDATE_FAILED)
        self.assertEqual(comp_result.message, "Component update failed")

    def test_target_result_spec_from_dict_no_components(self):
        """Test TargetResultSpec from_dict without components."""
        data = {"status": "SUCCESS", "message": "No components to process"}

        result = TargetResultSpec.from_dict(data)
        self.assertEqual(result.status, "SUCCESS")
        self.assertEqual(result.message, "No components to process")
        self.assertEqual(result.component_results, {})


class TestSummaryResult(unittest.TestCase):
    """Test cases for SummaryResult."""

    def test_summary_result_creation(self):
        """Test SummaryResult creation."""
        summary_spec = SummarySpec()
        summary = SummaryResult(summary=summary_spec, state=SummaryState.DONE, generation="1")

        self.assertEqual(summary.state, SummaryState.DONE)
        self.assertEqual(summary.generation, "1")
        self.assertEqual(summary.summary, summary_spec)

    def test_summary_result_is_deployment_finished(self):
        """Test SummaryResult is_deployment_finished method."""
        # Test with DONE state
        summary_done = SummaryResult(state=SummaryState.DONE)
        self.assertTrue(summary_done.is_deployment_finished())

        # Test with RUNNING state
        summary_running = SummaryResult(state=SummaryState.RUNNING)
        self.assertFalse(summary_running.is_deployment_finished())


class TestSummarySpec(unittest.TestCase):
    """Test cases for SummarySpec."""

    def test_summary_spec_creation(self):
        """Test SummarySpec creation."""
        spec = SummarySpec(target_count=2, success_count=1)

        self.assertEqual(spec.target_count, 2)
        self.assertEqual(spec.success_count, 1)

    def test_summary_spec_defaults(self):
        """Test SummarySpec with default values."""
        spec = SummarySpec()
        self.assertEqual(spec.target_count, 0)
        self.assertEqual(spec.success_count, 0)


class TestHelperFunctions(unittest.TestCase):
    """Test cases for helper functions."""

    def test_create_success_component_result(self):
        """Test create_success_component_result function."""
        result = create_success_component_result("Operation completed")

        self.assertEqual(result.status, State.OK)
        self.assertEqual(result.message, "Operation completed")

    def test_create_success_component_result_default_message(self):
        """Test create_success_component_result with default message."""
        result = create_success_component_result()

        self.assertEqual(result.status, State.OK)
        self.assertEqual(result.message, "")

    def test_create_failed_component_result(self):
        """Test create_failed_component_result function."""
        result = create_failed_component_result("Operation failed", State.UPDATE_FAILED)

        self.assertEqual(result.status, State.UPDATE_FAILED)
        self.assertEqual(result.message, "Operation failed")

    def test_create_failed_component_result_default_state(self):
        """Test create_failed_component_result with default state."""
        result = create_failed_component_result("Something went wrong")

        self.assertEqual(result.status, State.INTERNAL_ERROR)
        self.assertEqual(result.message, "Something went wrong")

    def test_create_target_result(self):
        """Test create_target_result function."""
        components = {
            "comp1": create_success_component_result("Updated"),
            "comp2": create_failed_component_result("Failed"),
        }

        result = create_target_result("PARTIAL", "Some components failed", components)

        self.assertEqual(result.status, "PARTIAL")
        self.assertEqual(result.message, "Some components failed")
        self.assertEqual(len(result.component_results), 2)
        self.assertEqual(result.component_results["comp1"].status, State.OK)
        self.assertEqual(result.component_results["comp2"].status, State.INTERNAL_ERROR)

    def test_create_target_result_minimal(self):
        """Test create_target_result with minimal parameters."""
        result = create_target_result()

        self.assertEqual(result.status, "OK")
        self.assertEqual(result.message, "")
        self.assertEqual(result.component_results, {})


if __name__ == "__main__":
    unittest.main()
