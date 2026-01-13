#!/usr/bin/env python3
"""
Comprehensive unit tests for Symphony SDK core functionality.
"""

import base64
import json
import unittest

from symphony_sdk.models import (
    COARequest,
    COAResponse,
    ComponentSpec,
    DeploymentSpec,
    ObjectMeta,
    SolutionSpec,
    SolutionState,
    deserialize_coa_request,
    deserialize_coa_response,
    deserialize_components,
    deserialize_deployment,
    serialize_coa_request,
    serialize_coa_response,
    serialize_components,
    to_dict,
)
from symphony_sdk.types import State


class TestObjectMeta(unittest.TestCase):
    """Test cases for ObjectMeta dataclass."""

    def test_object_meta_creation(self):
        """Test ObjectMeta creation with default values."""
        meta = ObjectMeta()
        self.assertEqual(meta.name, "")
        self.assertEqual(meta.namespace, "")
        self.assertIsNone(meta.labels)
        self.assertIsNone(meta.annotations)

    def test_object_meta_with_values(self):
        """Test ObjectMeta creation with specific values."""
        labels = {"app": "test", "version": "1.0"}
        annotations = {"description": "test object"}
        meta = ObjectMeta(
            name="test-object", namespace="test-namespace", labels=labels, annotations=annotations
        )
        self.assertEqual(meta.name, "test-object")
        self.assertEqual(meta.namespace, "test-namespace")
        self.assertEqual(meta.labels, labels)
        self.assertEqual(meta.annotations, annotations)


class TestComponentSpec(unittest.TestCase):
    """Test cases for ComponentSpec dataclass."""

    def test_component_spec_creation(self):
        """Test ComponentSpec creation with defaults."""
        comp = ComponentSpec()
        self.assertEqual(comp.name, "")
        self.assertEqual(comp.type, "")
        self.assertIsNone(comp.routes)
        self.assertEqual(comp.constraints, "")
        self.assertIsNone(comp.properties)

    def test_component_spec_with_values(self):
        """Test ComponentSpec creation with specific values."""
        properties = {"key1": "value1", "key2": "value2"}
        comp = ComponentSpec(
            name="test-component", type="service", constraints="cpu=2", properties=properties
        )

        self.assertEqual(comp.name, "test-component")
        self.assertEqual(comp.type, "service")
        self.assertEqual(comp.constraints, "cpu=2")
        self.assertEqual(comp.properties, properties)


class TestDeploymentSpec(unittest.TestCase):
    """Test cases for DeploymentSpec dataclass."""

    def test_deployment_spec_creation(self):
        """Test DeploymentSpec creation with defaults."""
        deployment = DeploymentSpec()
        self.assertEqual(deployment.solutionName, "")
        self.assertIsNone(deployment.solution)
        self.assertIsNone(deployment.instance)
        self.assertEqual(deployment.activeTarget, "")

    def test_deployment_spec_get_components_slice_no_solution(self):
        """Test get_components_slice with no solution."""
        deployment = DeploymentSpec()
        components = deployment.get_components_slice()
        self.assertEqual(components, [])

    def test_deployment_spec_get_components_slice_with_indices(self):
        """Test get_components_slice with start and end indices."""
        # Create components
        comp1 = ComponentSpec(name="comp1")
        comp2 = ComponentSpec(name="comp2")
        comp3 = ComponentSpec(name="comp3")

        # Create solution with components
        solution_spec = SolutionSpec(components=[comp1, comp2, comp3])
        solution_state = SolutionState(spec=solution_spec)

        deployment = DeploymentSpec(
            solution=solution_state, componentStartIndex=1, componentEndIndex=3
        )

        components = deployment.get_components_slice()
        self.assertEqual(len(components), 2)
        self.assertEqual(components[0].name, "comp2")
        self.assertEqual(components[1].name, "comp3")

    def test_deployment_spec_get_components_slice_all_components(self):
        """Test get_components_slice returning all components."""
        comp1 = ComponentSpec(name="comp1")
        comp2 = ComponentSpec(name="comp2")

        solution_spec = SolutionSpec(components=[comp1, comp2])
        solution_state = SolutionState(spec=solution_spec)

        deployment = DeploymentSpec(solution=solution_state)

        components = deployment.get_components_slice()
        self.assertEqual(len(components), 2)
        self.assertEqual(components[0].name, "comp1")
        self.assertEqual(components[1].name, "comp2")


class TestCOABodyMixin(unittest.TestCase):
    """Test cases for COABodyMixin functionality."""

    def test_coa_request_set_json_body(self):
        """Test COARequest set_body with JSON data."""
        request = COARequest()
        data = {"key": "value", "number": 123}

        request.set_body(data, "application/json")

        self.assertEqual(request.content_type, "application/json")

        # Decode and verify
        decoded_body = json.loads(base64.b64decode(request.body).decode("utf-8"))
        self.assertEqual(decoded_body, data)

    def test_coa_request_set_text_body(self):
        """Test COARequest set_body with text data."""
        request = COARequest()
        text_data = "Hello, World!"

        request.set_body(text_data, "text/plain")

        self.assertEqual(request.content_type, "text/plain")
        self.assertEqual(request.body, text_data)

    def test_coa_request_set_binary_body(self):
        """Test COARequest set_body with binary data."""
        request = COARequest()
        binary_data = b"Binary data content"

        request.set_body(binary_data, "application/octet-stream")

        self.assertEqual(request.content_type, "application/octet-stream")

        # Decode and verify
        decoded_body = base64.b64decode(request.body)
        self.assertEqual(decoded_body, binary_data)

    def test_coa_request_get_json_body(self):
        """Test COARequest get_body with JSON data."""
        request = COARequest()
        data = {"test": "data", "array": [1, 2, 3]}

        request.set_body(data, "application/json")
        retrieved_data = request.get_body()

        self.assertEqual(retrieved_data, data)

    def test_coa_request_get_text_body(self):
        """Test COARequest get_body with text data."""
        request = COARequest()
        text_data = "Simple text content"

        request.set_body(text_data, "text/plain")
        retrieved_data = request.get_body()

        self.assertEqual(retrieved_data, text_data)

    def test_coa_request_get_binary_body(self):
        """Test COARequest get_body with binary data."""
        request = COARequest()
        binary_data = b"Binary content for testing"

        request.set_body(binary_data, "application/octet-stream")
        retrieved_data = request.get_body()

        self.assertEqual(retrieved_data, binary_data)


class TestCOARequest(unittest.TestCase):
    """Test cases for COARequest dataclass."""

    def test_coa_request_creation(self):
        """Test COARequest creation with defaults."""
        request = COARequest()
        self.assertEqual(request.method, "GET")
        self.assertEqual(request.route, "")
        self.assertEqual(request.content_type, "application/json")
        self.assertEqual(request.body, "")

    def test_coa_request_with_values(self):
        """Test COARequest creation with specific values."""
        metadata = {"version": "1.0"}
        parameters = {"param1": "value1"}

        request = COARequest(
            method="POST", route="/api/v1/deploy", metadata=metadata, parameters=parameters
        )

        self.assertEqual(request.method, "POST")
        self.assertEqual(request.route, "/api/v1/deploy")
        self.assertEqual(request.metadata, metadata)
        self.assertEqual(request.parameters, parameters)

    def test_coa_request_to_json_dict(self):
        """Test COARequest to_json_dict conversion."""
        request = COARequest(
            method="PUT", route="/test", metadata={"key": "value"}, parameters={"param": "test"}
        )
        request.set_body({"data": "test"})

        json_dict = request.to_json_dict()

        expected_keys = ["method", "route", "content-type", "body", "metadata", "parameters"]
        for key in expected_keys:
            self.assertIn(key, json_dict)

        self.assertEqual(json_dict["method"], "PUT")
        self.assertEqual(json_dict["route"], "/test")
        self.assertEqual(json_dict["content-type"], "application/json")


class TestCOAResponse(unittest.TestCase):
    """Test cases for COAResponse dataclass."""

    def test_coa_response_creation(self):
        """Test COAResponse creation with defaults."""
        response = COAResponse()
        self.assertEqual(response.state, State.OK)
        self.assertEqual(response.content_type, "application/json")
        self.assertEqual(response.body, "")

    def test_coa_response_success_factory(self):
        """Test COAResponse.success factory method."""
        data = {"result": "success", "count": 5}
        response = COAResponse.success(data)

        self.assertEqual(response.state, State.OK)
        self.assertEqual(response.content_type, "application/json")

        retrieved_data = response.get_body()
        self.assertEqual(retrieved_data, data)

    def test_coa_response_error_factory(self):
        """Test COAResponse.error factory method."""
        error_msg = "Something went wrong"
        response = COAResponse.error(error_msg, State.INTERNAL_ERROR)

        self.assertEqual(response.state, State.INTERNAL_ERROR)
        retrieved_data = response.get_body()
        self.assertEqual(retrieved_data["error"], error_msg)

    def test_coa_response_not_found_factory(self):
        """Test COAResponse.not_found factory method."""
        response = COAResponse.not_found("Resource not found")

        self.assertEqual(response.state, State.NOT_FOUND)
        retrieved_data = response.get_body()
        self.assertEqual(retrieved_data["error"], "Resource not found")

    def test_coa_response_bad_request_factory(self):
        """Test COAResponse.bad_request factory method."""
        response = COAResponse.bad_request("Invalid input")

        self.assertEqual(response.state, State.BAD_REQUEST)
        retrieved_data = response.get_body()
        self.assertEqual(retrieved_data["error"], "Invalid input")

    def test_coa_response_to_json_dict(self):
        """Test COAResponse to_json_dict conversion."""
        response = COAResponse(
            state=State.OK, metadata={"version": "1.0"}, redirect_uri="https://example.com/redirect"
        )
        response.set_body({"status": "ok"})

        json_dict = response.to_json_dict()

        expected_keys = ["content-type", "body", "state", "metadata", "redirectUri"]
        for key in expected_keys:
            self.assertIn(key, json_dict)

        self.assertEqual(json_dict["state"], State.OK.value)
        self.assertEqual(json_dict["redirectUri"], "https://example.com/redirect")


class TestUtilityFunctions(unittest.TestCase):
    """Test cases for utility functions."""

    def test_to_dict_with_simple_object(self):
        """Test to_dict with simple dataclass object."""
        comp = ComponentSpec(name="test", type="service")
        result = to_dict(comp)

        self.assertIsInstance(result, dict)
        self.assertEqual(result["name"], "test")
        self.assertEqual(result["type"], "service")

    def test_to_dict_with_none(self):
        """Test to_dict with None input."""
        result = to_dict(None)
        self.assertEqual(result, {})

    def test_to_dict_with_nested_objects(self):
        """Test to_dict with nested dataclass objects."""
        meta = ObjectMeta(name="test-meta")
        comp = ComponentSpec(name="test-comp")
        solution_spec = SolutionSpec(components=[comp])
        solution_state = SolutionState(metadata=meta, spec=solution_spec)

        result = to_dict(solution_state)

        self.assertIsInstance(result, dict)
        self.assertIn("metadata", result)
        self.assertIn("spec", result)
        self.assertEqual(result["metadata"]["name"], "test-meta")

    def test_serialize_components(self):
        """Test serialize_components function."""
        comp1 = ComponentSpec(name="comp1", type="service")
        comp2 = ComponentSpec(name="comp2", type="deployment")
        components = [comp1, comp2]

        json_str = serialize_components(components)

        self.assertIsInstance(json_str, str)

        # Verify JSON is valid
        data = json.loads(json_str)
        self.assertEqual(len(data), 2)
        self.assertEqual(data[0]["name"], "comp1")
        self.assertEqual(data[1]["name"], "comp2")

    def test_deserialize_components(self):
        """Test deserialize_components function."""
        json_data = [
            {"name": "test-comp1", "type": "service"},
            {"name": "test-comp2", "type": "deployment"},
        ]
        json_str = json.dumps(json_data)

        components = deserialize_components(json_str)

        self.assertEqual(len(components), 2)
        self.assertEqual(components[0].name, "test-comp1")
        self.assertEqual(components[0].type, "service")
        self.assertEqual(components[1].name, "test-comp2")
        self.assertEqual(components[1].type, "deployment")

    def test_deserialize_components_invalid_json(self):
        """Test deserialize_components with invalid JSON."""
        invalid_json = "{ invalid json }"
        components = deserialize_components(invalid_json)
        self.assertEqual(components, [])

    def test_deserialize_deployment(self):
        """Test deserialize_deployment function."""
        deployment_data = {"solutionName": "test-solution", "activeTarget": "test-target"}
        json_str = json.dumps(deployment_data)

        deployments = deserialize_deployment(json_str)

        self.assertEqual(len(deployments), 1)
        self.assertEqual(deployments[0].solutionName, "test-solution")
        self.assertEqual(deployments[0].activeTarget, "test-target")

    def test_serialize_coa_request(self):
        """Test serialize_coa_request function."""
        request = COARequest(method="POST", route="/test", metadata={"key": "value"})
        request.set_body({"data": "test"})

        json_str = serialize_coa_request(request)

        self.assertIsInstance(json_str, str)

        # Verify JSON is valid
        data = json.loads(json_str)
        self.assertEqual(data["method"], "POST")
        self.assertEqual(data["route"], "/test")

    def test_deserialize_coa_request(self):
        """Test deserialize_coa_request function."""
        request_data = {
            "method": "PUT",
            "route": "/api/test",
            "content-type": "application/json",
            "body": base64.b64encode(json.dumps({"test": "data"}).encode()).decode(),
            "metadata": {"version": "1.0"},
        }
        json_str = json.dumps(request_data)

        request = deserialize_coa_request(json_str)

        self.assertEqual(request.method, "PUT")
        self.assertEqual(request.route, "/api/test")
        self.assertEqual(request.content_type, "application/json")
        self.assertEqual(request.metadata, {"version": "1.0"})

    def test_serialize_coa_response(self):
        """Test serialize_coa_response function."""
        response = COAResponse.success({"result": "ok"})

        json_str = serialize_coa_response(response)

        self.assertIsInstance(json_str, str)

        # Verify JSON is valid
        data = json.loads(json_str)
        self.assertEqual(data["state"], State.OK.value)

    def test_deserialize_coa_response(self):
        """Test deserialize_coa_response function."""
        response_data = {
            "content-type": "application/json",
            "body": base64.b64encode(json.dumps({"status": "success"}).encode()).decode(),
            "state": State.ACCEPTED.value,
            "metadata": {"processed": "true"},
        }
        json_str = json.dumps(response_data)

        response = deserialize_coa_response(json_str)

        self.assertEqual(response.state, State.ACCEPTED)
        self.assertEqual(response.content_type, "application/json")
        self.assertEqual(response.metadata, {"processed": "true"})


if __name__ == "__main__":
    unittest.main()
