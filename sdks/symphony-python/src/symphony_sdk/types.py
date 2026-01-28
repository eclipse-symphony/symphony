"""Symphony COA API Types - Python Translation

Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
SPDX-License-Identifier: MIT

This module provides Python translations of the Symphony COA API types
from the original Go implementation.
"""

import abc
from enum import IntEnum
from typing import Protocol


class Terminable(Protocol):
    """Interface for objects that can be gracefully terminated."""

    @abc.abstractmethod
    async def shutdown(self) -> None:
        """Shutdown the object gracefully."""
        pass


class State(IntEnum):
    """State represents a response state matching Symphony Go types."""

    # Basic states
    NONE = 0

    # HTTP Success states
    OK = 200  # HTTP 200
    ACCEPTED = 202  # HTTP 202

    # HTTP Client Error states
    BAD_REQUEST = 400  # HTTP 400
    UNAUTHORIZED = 401  # HTTP 401
    FORBIDDEN = 403  # HTTP 403
    NOT_FOUND = 404  # HTTP 404
    METHOD_NOT_ALLOWED = 405  # HTTP 405
    CONFLICT = 409  # HTTP 409
    STATUS_UNPROCESSABLE_ENTITY = 422  # HTTP 422

    # HTTP Server Error states
    INTERNAL_ERROR = 500  # HTTP 500

    # Config errors
    BAD_CONFIG = 1000
    MISSING_CONFIG = 1001

    # API invocation errors
    INVALID_ARGUMENT = 2000
    API_REDIRECT = 3030

    # IO errors
    FILE_ACCESS_ERROR = 4000

    # Serialization errors
    SERIALIZATION_ERROR = 5000
    DESERIALIZE_ERROR = 5001

    # Async requests
    DELETE_REQUESTED = 6000

    # Operation results
    UPDATE_FAILED = 8001
    DELETE_FAILED = 8002
    VALIDATE_FAILED = 8003
    UPDATED = 8004
    DELETED = 8005

    # Workflow status
    RUNNING = 9994
    PAUSED = 9995
    DONE = 9996
    DELAYED = 9997
    UNTOUCHED = 9998
    NOT_IMPLEMENTED = 9999

    # Detailed error codes
    INIT_FAILED = 10000
    CREATE_ACTION_CONFIG_FAILED = 10001
    HELM_ACTION_FAILED = 10002
    GET_COMPONENT_SPEC_FAILED = 10003
    CREATE_PROJECTOR_FAILED = 10004
    K8S_REMOVE_SERVICE_FAILED = 10005
    K8S_REMOVE_DEPLOYMENT_FAILED = 10006
    K8S_DEPLOYMENT_FAILED = 10007
    READ_YAML_FAILED = 10008
    APPLY_YAML_FAILED = 10009
    READ_RESOURCE_PROPERTY_FAILED = 10010
    APPLY_RESOURCE_FAILED = 10011
    DELETE_YAML_FAILED = 10012
    DELETE_RESOURCE_FAILED = 10013
    CHECK_RESOURCE_STATUS_FAILED = 10014
    APPLY_SCRIPT_FAILED = 10015
    REMOVE_SCRIPT_FAILED = 10016
    YAML_RESOURCE_PROPERTY_NOT_FOUND = 10017
    GET_HELM_PROPERTY_FAILED = 10018
    HELM_CHART_PULL_FAILED = 10019
    HELM_CHART_LOAD_FAILED = 10020
    HELM_CHART_APPLY_FAILED = 10021
    HELM_CHART_UNINSTALL_FAILED = 10022
    INGRESS_APPLY_FAILED = 10023
    HTTP_NEW_REQUEST_FAILED = 10024
    HTTP_SEND_REQUEST_FAILED = 10025
    HTTP_ERROR_RESPONSE = 10026
    MQTT_PUBLISH_FAILED = 10027
    MQTT_APPLY_FAILED = 10028
    MQTT_APPLY_TIMEOUT = 10029
    CONFIG_MAP_APPLY_FAILED = 10030
    HTTP_BAD_WAIT_STATUS_CODE = 10031
    HTTP_NEW_WAIT_REQUEST_FAILED = 10032
    HTTP_SEND_WAIT_REQUEST_FAILED = 10033
    HTTP_ERROR_WAIT_RESPONSE = 10034
    HTTP_BAD_WAIT_EXPRESSION = 10035
    SCRIPT_EXECUTION_FAILED = 10036
    SCRIPT_RESULT_PARSING_FAILED = 10037
    WAIT_TO_GET_INSTANCES_FAILED = 10038
    WAIT_TO_GET_SITES_FAILED = 10039
    WAIT_TO_GET_CATALOGS_FAILED = 10040
    INVALID_WAIT_OBJECT_TYPE = 10041
    CATALOGS_GET_FAILED = 10042
    INVALID_INSTANCE_CATALOG = 10043
    CREATE_INSTANCE_FROM_CATALOG_FAILED = 10044
    INVALID_SOLUTION_CATALOG = 10045
    CREATE_SOLUTION_FROM_CATALOG_FAILED = 10046
    INVALID_TARGET_CATALOG = 10047
    CREATE_TARGET_FROM_CATALOG_FAILED = 10048
    INVALID_CATALOG_CATALOG = 10049
    CREATE_CATALOG_FROM_CATALOG_FAILED = 10050
    PARENT_OBJECT_MISSING = 10051
    PARENT_OBJECT_CREATE_FAILED = 10052
    MATERIALIZE_BATCH_FAILED = 10053
    DELETE_INSTANCE_FAILED = 10054
    CREATE_INSTANCE_FAILED = 10055
    DEPLOYMENT_NOT_REACHED = 10056
    INVALID_OBJECT_TYPE = 10057
    UNSUPPORTED_ACTION = 10058
    INSTANCE_GET_FAILED = 10059
    TARGET_GET_FAILED = 10060
    DELETE_SOLUTION_FAILED = 10061
    CREATE_SOLUTION_FAILED = 10062
    GET_ARM_DEPLOYMENT_PROPERTY_FAILED = 10071
    ENSURE_ARM_RESOURCE_GROUP_FAILED = 10072
    CREATE_ARM_DEPLOYMENT_FAILED = 10073
    CLEANUP_ARM_DEPLOYMENT_FAILED = 10074

    # Instance controller errors
    SOLUTION_GET_FAILED = 11000
    TARGET_CANDIDATES_NOT_FOUND = 11001
    TARGET_LIST_GET_FAILED = 11002
    OBJECT_INSTANCE_CONVERSION_FAILED = 11003
    TIMED_OUT = 11004

    # Target controller errors
    TARGET_PROPERTY_NOT_FOUND = 12000

    # Non-transient errors
    GET_COMPONENT_PROPS_FAILED = 50000

    def __str__(self) -> str:
        """Return human-readable string representation of the state."""
        state_strings = {
            State.OK: "OK",
            State.ACCEPTED: "Accepted",
            State.BAD_REQUEST: "Bad Request",
            State.UNAUTHORIZED: "Unauthorized",
            State.FORBIDDEN: "Forbidden",
            State.NOT_FOUND: "Not Found",
            State.METHOD_NOT_ALLOWED: "Method Not Allowed",
            State.CONFLICT: "Conflict",
            State.STATUS_UNPROCESSABLE_ENTITY: "Unprocessable Entity",
            State.INTERNAL_ERROR: "Internal Error",
            State.BAD_CONFIG: "Bad Config",
            State.MISSING_CONFIG: "Missing Config",
            State.INVALID_ARGUMENT: "Invalid Argument",
            State.API_REDIRECT: "API Redirect",
            State.FILE_ACCESS_ERROR: "File Access Error",
            State.SERIALIZATION_ERROR: "Serialization Error",
            State.DESERIALIZE_ERROR: "De-serialization Error",
            State.DELETE_REQUESTED: "Delete Requested",
            State.UPDATE_FAILED: "Update Failed",
            State.DELETE_FAILED: "Delete Failed",
            State.VALIDATE_FAILED: "Validate Failed",
            State.UPDATED: "Updated",
            State.DELETED: "Deleted",
            State.RUNNING: "Running",
            State.PAUSED: "Paused",
            State.DONE: "Done",
            State.DELAYED: "Delayed",
            State.UNTOUCHED: "Untouched",
            State.NOT_IMPLEMENTED: "Not Implemented",
            State.INIT_FAILED: "Init Failed",
            State.CREATE_ACTION_CONFIG_FAILED: "Create Action Config Failed",
            State.HELM_ACTION_FAILED: "Helm Action Failed",
            State.GET_COMPONENT_SPEC_FAILED: "Get Component Spec Failed",
            State.CREATE_PROJECTOR_FAILED: "Create Projector Failed",
            State.K8S_REMOVE_SERVICE_FAILED: "Remove K8s Service Failed",
            State.K8S_REMOVE_DEPLOYMENT_FAILED: "Remove K8s Deployment Failed",
            State.K8S_DEPLOYMENT_FAILED: "K8s Deployment Failed",
            State.READ_YAML_FAILED: "Read Yaml Failed",
            State.APPLY_YAML_FAILED: "Apply Yaml Failed",
            State.READ_RESOURCE_PROPERTY_FAILED: "Read Resource Property Failed",
            State.APPLY_RESOURCE_FAILED: "Apply Resource Failed",
            State.DELETE_YAML_FAILED: "Delete Yaml Failed",
            State.DELETE_RESOURCE_FAILED: "Delete Resource Failed",
            State.CHECK_RESOURCE_STATUS_FAILED: "Check Resource Status Failed",
            State.APPLY_SCRIPT_FAILED: "Apply Script Failed",
            State.REMOVE_SCRIPT_FAILED: "Remove Script Failed",
            State.YAML_RESOURCE_PROPERTY_NOT_FOUND: "Yaml or Resource Property Not Found",
            State.GET_HELM_PROPERTY_FAILED: "Get Helm Property Failed",
            State.HELM_CHART_PULL_FAILED: "Helm Chart Pull Failed",
            State.HELM_CHART_LOAD_FAILED: "Helm Chart Load Failed",
            State.HELM_CHART_APPLY_FAILED: "Helm Chart Apply Failed",
            State.HELM_CHART_UNINSTALL_FAILED: "Helm Chart Uninstall Failed",
            State.INGRESS_APPLY_FAILED: "Ingress Apply Failed",
            State.HTTP_NEW_REQUEST_FAILED: "Http New Request Failed",
            State.HTTP_SEND_REQUEST_FAILED: "Http Send Request Failed",
            State.HTTP_ERROR_RESPONSE: "Http Error Response",
            State.MQTT_PUBLISH_FAILED: "Mqtt Publish Failed",
            State.MQTT_APPLY_FAILED: "Mqtt Apply Failed",
            State.MQTT_APPLY_TIMEOUT: "Mqtt Apply Timeout",
            State.CONFIG_MAP_APPLY_FAILED: "ConfigMap Apply Failed",
            State.HTTP_BAD_WAIT_STATUS_CODE: "Http Bad Wait Status Code",
            State.HTTP_NEW_WAIT_REQUEST_FAILED: "Http New Wait Request Failed",
            State.HTTP_SEND_WAIT_REQUEST_FAILED: "Http Send Wait Request Failed",
            State.HTTP_ERROR_WAIT_RESPONSE: "Http Error Wait Response",
            State.HTTP_BAD_WAIT_EXPRESSION: "Http Bad Wait Expression",
            State.SCRIPT_EXECUTION_FAILED: "Script Execution Failed",
            State.SCRIPT_RESULT_PARSING_FAILED: "Script Result Parsing Failed",
            State.WAIT_TO_GET_INSTANCES_FAILED: "Wait To Get Instances Failed",
            State.WAIT_TO_GET_SITES_FAILED: "Wait To Get Sites Failed",
            State.WAIT_TO_GET_CATALOGS_FAILED: "Wait To Get Catalogs Failed",
            State.INVALID_WAIT_OBJECT_TYPE: "Invalid Wait Object Type",
            State.CATALOGS_GET_FAILED: "Get Catalogs Failed",
            State.INVALID_INSTANCE_CATALOG: "Invalid Instance Catalog",
            State.CREATE_INSTANCE_FROM_CATALOG_FAILED: "Create Instance From Catalog Failed",
            State.INVALID_SOLUTION_CATALOG: "Invalid Solution Object in Catalog",
            State.CREATE_SOLUTION_FROM_CATALOG_FAILED: "Create Solution Object From Catalog Failed",
            State.INVALID_TARGET_CATALOG: "Invalid Target Object in Catalog",
            State.CREATE_TARGET_FROM_CATALOG_FAILED: "Create Target Object From Catalog Failed",
            State.INVALID_CATALOG_CATALOG: "Invalid Catalog Object in Catalog",
            State.CREATE_CATALOG_FROM_CATALOG_FAILED: "Create Catalog Object From Catalog Failed",
            State.PARENT_OBJECT_MISSING: "Parent Object Missing",
            State.PARENT_OBJECT_CREATE_FAILED: "Parent Object Create Failed",
            State.MATERIALIZE_BATCH_FAILED: "Failed to Materialize all objects",
            State.DELETE_INSTANCE_FAILED: "Failed to Delete Instance",
            State.CREATE_INSTANCE_FAILED: "Failed to Create Instance",
            State.DEPLOYMENT_NOT_REACHED: "Deployment Not Reached",
            State.INVALID_OBJECT_TYPE: "Invalid Object Type",
            State.UNSUPPORTED_ACTION: "Unsupported Action",
            State.INSTANCE_GET_FAILED: "Get instance failed",
            State.TARGET_GET_FAILED: "Get target failed",
            State.SOLUTION_GET_FAILED: "Solution does not exist",
            State.TARGET_CANDIDATES_NOT_FOUND: "Target does not exist",
            State.TARGET_LIST_GET_FAILED: "Target list does not exist",
            State.OBJECT_INSTANCE_CONVERSION_FAILED: "Object to Instance conversion failed",
            State.TIMED_OUT: "Timed Out",
            State.TARGET_PROPERTY_NOT_FOUND: "Target Property Not Found",
            State.GET_COMPONENT_PROPS_FAILED: "Get component property failed",
        }

        return state_strings.get(self, f"Unknown State: {self.value}")

    def equals_with_string(self, string: str) -> bool:
        """Check if state equals a string representation."""
        return str(self) == string

    @classmethod
    def from_http_status(cls, code: int) -> "State":
        """Get State from HTTP status code."""
        if code == 200:
            return cls.OK
        elif code == 202:
            return cls.ACCEPTED
        elif 200 <= code < 300:
            return cls.OK
        elif code == 401:
            return cls.UNAUTHORIZED
        elif code == 403:
            return cls.FORBIDDEN
        elif code == 404:
            return cls.NOT_FOUND
        elif code == 405:
            return cls.METHOD_NOT_ALLOWED
        elif code == 409:
            return cls.CONFLICT
        elif 400 <= code < 500:
            return cls.BAD_REQUEST
        elif code >= 500:
            return cls.INTERNAL_ERROR
        else:
            return cls.NONE


# COA Constants
class COAConstants:
    """Constants used in Symphony COA operations."""

    # Header constants
    COA_META_HEADER = "COA_META_HEADER"

    # Tracing and monitoring
    TRACING_EXPORTER_CONSOLE = "tracing.exporters.console"
    METRICS_EXPORTER_OTLP_GRPC = "metrics.exporters.otlpgrpc"
    TRACING_EXPORTER_ZIPKIN = "tracing.exporters.zipkin"
    TRACING_EXPORTER_OTLP_GRPC = "tracing.exporters.otlpgrpc"
    LOG_EXPORTER_CONSOLE = "log.exporters.console"
    LOG_EXPORTER_OTLP_GRPC = "log.exporters.otlpgrpc"
    LOG_EXPORTER_OTLP_HTTP = "log.exporters.otlphttp"

    # Provider constants
    PROVIDERS_PERSISTENT_STATE = "providers.persistentstate"
    PROVIDERS_VOLATILE_STATE = "providers.volatilestate"
    PROVIDERS_CONFIG = "providers.config"
    PROVIDERS_SECRET = "providers.secret"
    PROVIDERS_REFERENCE = "providers.reference"
    PROVIDERS_PROBE = "providers.probe"
    PROVIDERS_UPLOADER = "providers.uploader"
    PROVIDERS_REPORTER = "providers.reporter"
    PROVIDER_QUEUE = "providers.queue"
    PROVIDER_LEDGER = "providers.ledger"
    PROVIDERS_KEY_LOCK = "providers.keylock"

    # Output constants
    STATUS_OUTPUT = "status"
    ERROR_OUTPUT = "error"
    STATE_OUTPUT = "__state"


# Helper functions for compatibility
def get_http_status(code: int) -> State:
    """Get State from HTTP status code (compatibility function)."""
    return State.from_http_status(code)


# Export commonly used items
__all__ = ["State", "Terminable", "COAConstants", "get_http_status"]
