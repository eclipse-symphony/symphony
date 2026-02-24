"""Symphony SDK data structures and utilities.

Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
SPDX-License-Identifier: MIT

This module provides Symphony-compatible data structures and helpers
following the official Eclipse Symphony COA patterns.
"""

# Standard library imports
import base64
import json
import logging
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional, get_args, get_origin

from symphony_sdk.types import State

logger = logging.getLogger(__name__)


@dataclass
class ObjectMeta:
    """Object metadata following Kubernetes-style metadata."""

    namespace: str = ""
    name: str = ""
    labels: Optional[Dict[str, str]] = None
    annotations: Optional[Dict[str, str]] = None


@dataclass
class TargetSelector:
    """Target selector for component binding."""

    name: str = ""
    selector: Optional[Dict[str, str]] = None


@dataclass
class BindingSpec:
    """Binding specification for component deployment."""

    role: str = ""
    provider: str = ""
    config: Optional[Dict[str, str]] = None


@dataclass
class TopologySpec:
    """Topology specification for deployment."""

    device: str = ""
    selector: Optional[Dict[str, str]] = None
    bindings: Optional[List[BindingSpec]] = None


@dataclass
class PipelineSpec:
    """Pipeline specification for data processing."""

    name: str = ""
    skill: str = ""
    parameters: Optional[Dict[str, str]] = None


@dataclass
class VersionSpec:
    """Version specification for solutions."""

    solution: str = ""
    percentage: int = 100


@dataclass
class InstanceSpec:
    """Instance specification for Symphony deployments."""

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


@dataclass
class FilterSpec:
    """Filter specification for routing."""

    direction: str = ""
    parameters: Optional[Dict[str, str]] = None
    type: str = ""


@dataclass
class RouteSpec:
    route: str = ""
    properties: Dict[str, str] = None
    filters: List[FilterSpec] = None
    type: str = ""


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


@dataclass
class SolutionSpec:
    components: List[ComponentSpec] = None
    scope: str = ""
    displayName: str = ""
    metadata: Dict[str, str] = None


@dataclass
class SolutionState:
    metadata: ObjectMeta = None
    spec: SolutionSpec = None


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


@dataclass
class ComponentError:
    code: str = ""
    message: str = ""
    target: str = ""


@dataclass
class TargetError:
    code: str = ""
    message: str = ""
    target: str = ""
    details: Dict[str, ComponentError] = None


@dataclass
class ErrorType:
    code: str = ""
    message: str = ""
    target: str = ""
    details: Dict[str, TargetError] = None


@dataclass
class ProvisioningStatus:
    operationId: str = ""
    status: str = ""
    failureCause: str = ""
    logErrors: bool = False
    error: ErrorType = None
    output: Dict[str, str] = None


@dataclass
class TargetStatus:
    properties: Dict[str, str] = None
    provisioningStatus: ProvisioningStatus = None
    lastModififed: str = ""


@dataclass
class TargetState:
    metadata: ObjectMeta = None
    spec: TargetSpec = None
    status: TargetStatus = None


@dataclass
class DeviceSpec:
    properties: Dict[str, str] = None
    bindings: List[BindingSpec] = None
    displayName: str = ""


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
        if self.solution != None:
            if (
                self.componentStartIndex >= 0
                and self.componentEndIndex >= 0
                and self.componentEndIndex > self.componentStartIndex
            ):
                return self.solution.spec.components[
                    self.componentStartIndex : self.componentEndIndex
                ]
            return self.solution.spec.components
        return []


@dataclass
class ComparisonPack:
    desired: List[ComponentSpec]
    current: List[ComponentSpec]


@dataclass
class COABodyMixin:
    """Common functionality for COA request and response body handling.

    Provides content-type aware body encoding/decoding with support for:
    - "application/json": JSON objects
    - "text/plain": Plain text strings
    - "application/octet-stream": Binary data
    """

    content_type: str = "application/json"
    body: str = ""  # Base64 encoded data (type determined by content_type)

    def set_body(self, data: Any, content_type: Optional[str] = None) -> None:
        """Set body data with content type detection or explicit content type.

        Args:
            data: Data to set (JSON object, string, or bytes)
            content_type: Optional explicit content type override
        """
        # Set content type if provided
        if content_type:
            self.content_type = content_type

        if self.content_type == "application/json":
            # Handle JSON data
            if isinstance(data, (str, bytes)):
                # If it's a string/bytes, try to parse as JSON first
                if isinstance(data, bytes):
                    data = data.decode("utf-8")
                json.loads(data)  # Validate JSON
                self.body = base64.b64encode(data.encode("utf-8")).decode("utf-8")
            else:
                # Serialize object to JSON
                json_str = json.dumps(data, ensure_ascii=False)
                self.body = base64.b64encode(json_str.encode("utf-8")).decode("utf-8")

        elif self.content_type == "text/plain":
            # Handle plain text - no encoding necessary
            if isinstance(data, bytes):
                self.body = data.decode("utf-8")
            else:
                self.body = str(data)

        elif self.content_type == "application/octet-stream":
            # Handle binary data
            if isinstance(data, str):
                # If string, assume it's already base64 encoded
                self.body = data
            elif isinstance(data, bytes):
                # Encode bytes to base64
                self.body = base64.b64encode(data).decode("utf-8")
            else:
                raise ValueError(
                    f"Binary content type requires bytes or base64 string, got {type(data)}"
                )

        else:
            # Default fallback - treat as text
            if isinstance(data, bytes):
                text_str = data.decode("utf-8")
            else:
                text_str = str(data)
            self.body = base64.b64encode(text_str.encode("utf-8")).decode("utf-8")

    def get_body(self) -> Any:
        """Get body data decoded according to content type.

        Returns:
            Decoded body data (JSON object, string, or bytes depending on content_type)
        """
        if not self.body:
            return None

        if self.content_type == "application/json":
            # Return parsed JSON object
            try:
                json_str = base64.b64decode(self.body).decode("utf-8")
                return json.loads(json_str)
            except (ValueError, json.JSONDecodeError) as e:
                raise ValueError(f"Invalid JSON in body: {e}")

        elif self.content_type == "text/plain":
            # Return text string directly - no decoding necessary
            return self.body

        elif self.content_type == "application/octet-stream":
            # Return raw bytes
            return base64.b64decode(self.body)

        else:
            # Default fallback - return as string
            return base64.b64decode(self.body).decode("utf-8")


@dataclass
class COARequest(COABodyMixin):
    """COA Request structure based on Symphony COA API.

    This corresponds to the Go struct COARequest from Symphony codebase.
    The body field contains base64 encoded data, with the content type determined by content_type:
    - "application/json": Base64 encoded UTF-8 string of JSON object
    - "text/plain": Base64 encoded UTF-8 string of plain text
    - "application/octet-stream": Base64 encoded binary data
    """

    method: str = "GET"
    route: str = ""
    metadata: Optional[Dict[str, str]] = field(default_factory=dict)
    parameters: Optional[Dict[str, str]] = field(default_factory=dict)

    def to_json_dict(self) -> Dict[str, Any]:
        """Convert to JSON-serializable dictionary."""
        result = {
            "method": self.method,
            "route": self.route,
            "content-type": self.content_type,
            "body": self.body,
        }
        if self.metadata:
            result["metadata"] = self.metadata
        if self.parameters:
            result["parameters"] = self.parameters
        return result


@dataclass
class COAResponse(COABodyMixin):
    """COA Response structure based on Symphony COA API.

    This corresponds to the Go struct COAResponse from Symphony codebase.
    The body field contains base64 encoded data, with the content type determined by content_type:
    - "application/json": Base64 encoded UTF-8 string of JSON object
    - "text/plain": Base64 encoded UTF-8 string of plain text
    - "application/octet-stream": Base64 encoded binary data
    """

    state: State = State.OK
    metadata: Optional[Dict[str, str]] = field(default_factory=dict)
    redirect_uri: Optional[str] = None

    def to_json_dict(self) -> Dict[str, Any]:
        """Convert to JSON-serializable dictionary."""
        result = {"content-type": self.content_type, "body": self.body, "state": self.state.value}
        if self.metadata:
            result["metadata"] = self.metadata
        if self.redirect_uri:
            result["redirectUri"] = self.redirect_uri
        return result

    @classmethod
    def success(cls, data: Any = None, content_type: str = "application/json") -> "COAResponse":
        """Create a success response."""
        response = cls(content_type=content_type, state=State.OK)
        if data is not None:
            response.set_body(data, content_type)
        return response

    @classmethod
    def error(
        cls,
        message: str,
        state: State = State.INTERNAL_ERROR,
        content_type: str = "application/json",
    ) -> "COAResponse":
        """Create an error response."""
        response = cls(content_type=content_type, state=state)
        if content_type == "application/json":
            response.set_body({"error": message}, content_type)
        elif content_type == "text/plain":
            response.set_body(f"Error: {message}", content_type)
        else:
            response.set_body({"error": message}, "application/json")
        return response

    @classmethod
    def not_found(cls, message: str = "Resource not found") -> "COAResponse":
        """Create a not found response."""
        return cls.error(message, State.NOT_FOUND)

    @classmethod
    def bad_request(cls, message: str = "Bad request") -> "COAResponse":
        """Create a bad request response."""
        return cls.error(message, State.BAD_REQUEST)


# Utility functions for COA data conversion
def to_dict(obj: Any) -> Dict[str, Any]:
    """Convert dataclass object to dictionary."""
    if obj is None:
        return {}

    if hasattr(obj, "__dict__"):
        result = {}
        for key, value in obj.__dict__.items():
            if value is not None:
                if isinstance(value, list):
                    result[key] = [to_dict(item) for item in value]
                elif isinstance(value, dict):
                    result[key] = {k: to_dict(v) for k, v in value.items()}
                elif hasattr(value, "__dict__"):
                    result[key] = to_dict(value)
                elif isinstance(value, Enum):
                    result[key] = value.value
                else:
                    result[key] = value
        return result

    return obj


def from_dict(data: Dict[str, Any], cls: type) -> Any:
    """Convert dictionary to dataclass object."""
    if not data:
        return cls()

    try:
        # Handle enums
        if isinstance(cls, type) and issubclass(cls, Enum):
            return cls(data)

        # Handle basic types
        if cls in (str, int, float, bool):
            return cls(data)

        # Handle dataclass
        if hasattr(cls, "__dataclass_fields__"):
            kwargs = {}
            for field_name, field_info in cls.__dataclass_fields__.items():
                if field_name in data:
                    field_type = field_info.type
                    field_value = data[field_name]

                    # Handle Optional types
                    if get_origin(field_type) is Optional and type(None) in get_args(field_type):
                        if field_value is None:
                            kwargs[field_name] = None
                        else:
                            inner_type = field_type.__args__[0]
                            kwargs[field_name] = from_dict(field_value, inner_type)

                    # Handle List types
                    elif hasattr(field_type, "__origin__") and field_type.__origin__ is list:
                        if field_value and isinstance(field_value, list):
                            inner_type = field_type.__args__[0]
                            kwargs[field_name] = [
                                from_dict(item, inner_type) for item in field_value
                            ]
                        else:
                            kwargs[field_name] = field_value or []

                    # Handle Dict types
                    elif hasattr(field_type, "__origin__") and field_type.__origin__ is dict:
                        kwargs[field_name] = field_value or {}

                    # Handle nested dataclasses
                    elif hasattr(field_type, "__dataclass_fields__"):
                        kwargs[field_name] = from_dict(field_value, field_type)

                    # Handle enums
                    elif isinstance(field_type, type) and issubclass(field_type, Enum):
                        kwargs[field_name] = field_type(field_value)

                    else:
                        kwargs[field_name] = field_value

            return cls(**kwargs)

    except Exception:
        # Fallback: return default instance
        return cls()


def serialize_components(components: List[ComponentSpec]) -> str:
    """Serialize components to JSON string."""
    return json.dumps([to_dict(comp) for comp in components], indent=2)


def deserialize_components(json_str: str) -> List[ComponentSpec]:
    """Deserialize JSON string to components list."""
    try:
        data = json.loads(json_str)
        return [from_dict(item, ComponentSpec) for item in data]
    except Exception:
        return []


def deserialize_solution(json_str: str) -> List[SolutionState]:
    """Deserialize JSON string to components list."""
    try:
        data = json.loads(json_str)
        return [from_dict(item, SolutionState) for item in data]
    except Exception:
        return []


# Desrialize a DeploymentSpec object  from Json String
def deserialize_deployment(json_str: str) -> List[DeploymentSpec]:
    """Deserialize JSON string to DeploymentSpec list."""
    try:
        data = json.loads(json_str)
        return [from_dict(data, DeploymentSpec)]
    except Exception as e:
        logger.error("Error deserializing deployment: %s", e)
        return []


def serialize_coa_request(coa_request: COARequest) -> str:
    """Serialize COARequest to JSON string."""
    return json.dumps(coa_request.to_json_dict(), indent=2)


def deserialize_coa_request(json_str: str) -> COARequest:
    """Deserialize JSON string to COARequest."""
    try:
        data = json.loads(json_str)
        request = COARequest()

        # Map JSON fields to dataclass fields
        if "method" in data:
            request.method = data["method"]
        if "route" in data:
            request.route = data["route"]
        if "content-type" in data:
            request.content_type = data["content-type"]
        if "body" in data:
            body_data = data["body"]
            if isinstance(body_data, str):
                # Assume it's already base64 encoded JSON string
                request.body = body_data
            else:
                # Convert object to JSON and base64 encode
                request.set_body(body_data, "application/json")
        if "metadata" in data:
            request.metadata = data["metadata"]
        if "parameters" in data:
            request.parameters = data["parameters"]

        return request
    except Exception as e:
        logger.error("Error deserializing COA request: %s", e)
        return COARequest()


def serialize_coa_response(coa_response: COAResponse) -> str:
    """Serialize COAResponse to JSON string."""
    return json.dumps(coa_response.to_json_dict(), indent=2)


def deserialize_coa_response(json_str: str) -> COAResponse:
    """Deserialize JSON string to COAResponse."""
    try:
        data = json.loads(json_str)
        response = COAResponse()

        # Map JSON fields to dataclass fields
        if "content-type" in data:
            response.content_type = data["content-type"]
        if "body" in data:
            body_data = data["body"]
            if isinstance(body_data, str):
                # Assume it's already base64 encoded JSON string
                response.body = body_data
            else:
                # Convert object to JSON and base64 encode
                response.set_body(body_data, "application/json")
        if "state" in data:
            try:
                response.state = State(data["state"])
            except ValueError:
                response.state = State.INTERNAL_ERROR
        if "metadata" in data:
            response.metadata = data["metadata"]
        if "redirectUri" in data:
            response.redirect_uri = data["redirectUri"]

        return response
    except Exception as e:
        logger.error("Error deserializing COA response: %s", e)
        return COAResponse()


__all__ = [
    "ObjectMeta",
    "TargetSelector",
    "BindingSpec",
    "TopologySpec",
    "PipelineSpec",
    "VersionSpec",
    "InstanceSpec",
    "FilterSpec",
    "RouteSpec",
    "ComponentSpec",
    "SolutionSpec",
    "SolutionState",
    "TargetSpec",
    "ComponentError",
    "TargetError",
    "ErrorType",
    "ProvisioningStatus",
    "TargetStatus",
    "TargetState",
    "DeviceSpec",
    "DeploymentSpec",
    "ComparisonPack",
    "COABodyMixin",
    "COARequest",
    "COAResponse",
    "to_dict",
    "from_dict",
    "serialize_components",
    "deserialize_components",
    "deserialize_solution",
    "deserialize_deployment",
    "serialize_coa_request",
    "deserialize_coa_request",
    "serialize_coa_response",
    "deserialize_coa_response",
]
