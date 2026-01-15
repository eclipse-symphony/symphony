"""Symphony REST API Client.

Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
SPDX-License-Identifier: MIT

This module provides a client for interacting with the Symphony REST API,
including authentication, target registration/unregistration, and other
Symphony operations based on the OpenAPI specification.
"""

import json
import logging
from datetime import datetime, timezone, timedelta
from typing import Any, Dict, List, Optional, Union

JSONType = Union[Dict[str, Any], List[Dict[str, Any]]]

import requests


class SymphonyAPIError(Exception):
    """Custom exception for Symphony API errors."""

    def __init__(
        self, message: str, status_code: Optional[int] = None, response_text: Optional[str] = None
    ):
        super().__init__(message)
        self.status_code = status_code
        self.response_text = response_text


class SymphonyAPI:
    """Symphony REST API Client.

    Provides methods for interacting with the Symphony API including:
    - Authentication
    - Target management (register/unregister)
    - Solution management
    - Instance management
    - Other Symphony operations
    """

    def __init__(
        self,
        base_url: str,
        username: str,
        password: str,
        timeout: float = 30.0,
        logger: Optional[logging.Logger] = None,
    ):
        """Initialize the Symphony API client.

        Args:
            base_url: Base URL of the Symphony API (e.g., 'https://symphony.example.com')
            username: Symphony username for authentication
            password: Symphony password for authentication
            timeout: Request timeout in seconds
            logger: Optional logger instance
        """
        self.base_url = base_url.rstrip("/")
        self.username = username
        self.password = password
        self.timeout = timeout
        self.logger = logger or logging.getLogger(__name__)

        # Authentication state
        self._access_token: Optional[str] = None
        self._token_expiry: Optional[datetime] = None

        # Session for connection reuse
        self._session = requests.Session()
        self._session.headers.update(
            {"Content-Type": "application/json", "User-Agent": "SymphonySDK/0.1.0"}
        )

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit - close session."""
        self.close()

    def close(self):
        """Close the HTTP session."""
        if self._session:
            self._session.close()

    def _make_request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make an HTTP request to the Symphony API.

        Args:
            method: HTTP method (GET, POST, PUT, DELETE, etc.)
            endpoint: API endpoint (without base URL)
            **kwargs: Additional arguments for requests

        Returns:
            requests.Response object

        Raises:
            SymphonyAPIError: If the request fails
        """
        url = f"{self.base_url}/{endpoint.lstrip('/')}"

        # Set default timeout
        kwargs.setdefault("timeout", self.timeout)

        try:
            self.logger.debug(f"Making {method} request to {url}")
            response = self._session.request(method, url, **kwargs)

            self.logger.debug(f"Response: {response.status_code} {response.reason}")

            return response

        except requests.exceptions.RequestException as e:
            self.logger.error(f"Request failed: {e}")
            raise SymphonyAPIError(f"Request failed: {str(e)}")

    def _handle_response(
        self, response: requests.Response, expected_codes: List[int] = None
    ) -> Optional[JSONType]:
        """Handle API response and extract JSON data.

        Args:
            response: requests.Response object
            expected_codes: List of expected HTTP status codes (default: [200])

        Returns:
            Parsed JSON response data

        Raises:
            SymphonyAPIError: If response indicates an error
        """
        if expected_codes is None:
            expected_codes = [200]

        if response.status_code not in expected_codes:
            error_msg = f"API request failed with status {response.status_code}: {response.reason}"
            self.logger.error(error_msg)
            self.logger.error(f"Response body: {response.text}")
            raise SymphonyAPIError(error_msg, response.status_code, response.text)

        # Handle empty responses
        if not response.content.strip():
            return {}

        try:
            return response.json()
        except json.JSONDecodeError as e:
            self.logger.error(f"Failed to parse JSON response: {e}")
            self.logger.error(f"Response body: {response.text}")
            raise SymphonyAPIError(f"Invalid JSON response: {str(e)}")

    def authenticate(self, force_refresh: bool = False) -> str:
        """Authenticate with Symphony API and return access token.

        Args:
            force_refresh: Force token refresh even if current token is valid

        Returns:
            Access token string

        Raises:
            SymphonyAPIError: If authentication fails
        """
        # Check if we have a valid token
        if not force_refresh and self._access_token and self._token_expiry:
            # Use timezone-aware UTC datetimes
            if datetime.now(timezone.utc) < self._token_expiry:
                return self._access_token

        self.logger.info(f"Authenticating with Symphony API as user '{self.username}'")

        auth_payload = {"username": self.username, "password": self.password}

        response = self._make_request("POST", "/users/auth", json=auth_payload)
        data = self._handle_response(response)

        access_token = data.get("accessToken")
        if not access_token:
            raise SymphonyAPIError("No access token in authentication response")

        self._access_token = access_token
        # Assume token is valid for ~1 hour
        # Use timezone-aware UTC datetimes and truncate seconds/micros
        self._token_expiry = (
            datetime.now(timezone.utc)
            .replace(second=0, microsecond=0)
            + timedelta(minutes=50)
        )
        # Update session headers with token
        self._session.headers.update({"Authorization": f"Bearer {access_token}"})

        self.logger.info("Successfully authenticated with Symphony API")
        return access_token

    def _ensure_authenticated(self):
        """Ensure we have a valid authentication token."""
        if not self._access_token:
            self.authenticate()
        # Token refresh is handled automatically in authenticate()

    # Target Management Methods
    def register_target(self, target_name: str, target_spec: Dict[str, Any]) -> Dict[str, Any]:
        """Register a target with Symphony.

        Args:
            target_name: Name of the target to register
            target_spec: Target specification dictionary (should contain fields like
                        displayName, scope, properties, components, metadata, etc.)

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If registration fails
        """
        self._ensure_authenticated()

        self.logger.info(f"Registering target '{target_name}' with Symphony")

        # Wrap the spec in a TargetState structure as expected by the API
        target_state = {"metadata": {"name": target_name}, "spec": target_spec}

        response = self._make_request("POST", f"/targets/registry/{target_name}", json=target_state)

        data = self._handle_response(response, [200, 201])
        self.logger.info(f"Successfully registered target '{target_name}'")

        return data

    def unregister_target(self, target_name: str, direct: bool = False) -> Dict[str, Any]:
        """Unregister a target from Symphony.

        Args:
            target_name: Name of the target to unregister
            direct: Whether to use direct delete

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If unregistration fails
        """
        self._ensure_authenticated()

        self.logger.info(f"Unregistering target '{target_name}' from Symphony")

        params = {"direct": "true"} if direct else {}

        response = self._make_request("DELETE", f"/targets/registry/{target_name}", params=params)

        data = self._handle_response(response, [200, 204])
        self.logger.info(f"Successfully unregistered target '{target_name}'")

        return data

    def get_target(self, target_name: str) -> Dict[str, Any]:
        """Get target specification.

        Args:
            target_name: Name of the target

        Returns:
            Target specification data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request(
            "GET",
            f"/targets/registry/{target_name}",
        )

        return self._handle_response(response)

    def list_targets(self) -> List[Dict[str, Any]]:
        """List all registered targets.

        Returns:
            List of targets

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("GET", "/targets/registry")
        return self._handle_response(response)

    def ping_target(self, target_name: str) -> Dict[str, Any]:
        """Send heartbeat ping to target.

        Args:
            target_name: Name of the target to ping

        Returns:
            Ping response data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("POST", f"/targets/ping/{target_name}")
        return self._handle_response(response)

    def update_target_status(self, target_name: str, status_data: Dict[str, Any]) -> Dict[str, Any]:
        """Update target status.

        Args:
            target_name: Name of the target
            status_data: Status information dictionary

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("PUT", f"/targets/status/{target_name}", json=status_data)

        return self._handle_response(response)

    # Solution Management Methods

    def create_solution(
        self,
        solution_name: str,
        solution_spec: str,
        namespace: str = "default",
        embed_type: Optional[str] = None,
        embed_component: Optional[str] = None,
        embed_property: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Create a solution with embedded specification.

        Args:
            solution_name: Name of the solution
            solution_spec: Solution specification as YAML/JSON text (will be parsed)
            namespace: Namespace/scope for the solution (default: "default")
            embed_type: Optional embed type
            embed_component: Optional embed component
            embed_property: Optional embed property

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        params = {"namespace": namespace}
        if embed_type:
            params["embed-type"] = embed_type
        if embed_component:
            params["embed-component"] = embed_component
        if embed_property:
            params["embed-property"] = embed_property

        # Parse the YAML/JSON spec and wrap it in SolutionState structure
        try:
            import yaml

            spec_dict = yaml.safe_load(solution_spec)
        except Exception as e:
            raise SymphonyAPIError(f"Failed to parse solution specification: {str(e)}")

        # Wrap in SolutionState structure expected by the API
        solution_state = {
            "metadata": {"name": solution_name, "namespace": namespace},
            "spec": spec_dict,
        }

        response = self._make_request(
            "POST", f"/solutions/{solution_name}", json=solution_state, params=params
        )

        return self._handle_response(response, [200, 201])

    def get_solution(self, solution_name: str, namespace: str = "default") -> Dict[str, Any]:
        """Get solution specification.

        Args:
            solution_name: Name of the solution
            namespace: Namespace/scope of the solution (default: "default")

        Returns:
            Solution specification data (returns the full SolutionState with metadata and spec)

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        params = {"namespace": namespace}

        response = self._make_request("GET", f"/solutions/{solution_name}", params=params)

        # Return the full solution state (metadata + spec)
        return self._handle_response(response)

    def delete_solution(self, solution_name: str) -> Dict[str, Any]:
        """Delete a solution.

        Args:
            solution_name: Name of the solution to delete

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("DELETE", f"/solutions/{solution_name}")
        return self._handle_response(response)

    def list_solutions(self) -> List[Dict[str, Any]]:
        """List all solutions.

        Returns:
            List of solutions

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("GET", "/solutions")
        return self._handle_response(response)

    # Instance Management Methods

    def create_instance(
        self, instance_name: str, instance_spec: Dict[str, Any], namespace: str = "default"
    ) -> Dict[str, Any]:
        """Create an instance.

        Args:
            instance_name: Name of the instance
            instance_spec: Instance specification dictionary (should contain solution, target, etc.)
            namespace: Namespace/scope for the instance (default: "default")

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        # Wrap the spec in InstanceState structure as expected by the API
        instance_state = {
            "metadata": {"name": instance_name, "namespace": namespace},
            "spec": instance_spec,
        }

        params = {"namespace": namespace}

        response = self._make_request(
            "POST", f"/instances/{instance_name}", json=instance_state, params=params
        )

        return self._handle_response(response, [200, 201])

    def get_instance(self, instance_name: str, namespace: str = "default") -> Dict[str, Any]:
        """Get instance specification.

        Args:
            instance_name: Name of the instance
            namespace: Namespace/scope of the instance (default: "default")

        Returns:
            Instance specification data (returns the full InstanceState with metadata and spec)

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        params = {"namespace": namespace}

        response = self._make_request("GET", f"/instances/{instance_name}", params=params)

        # Return the full instance state (metadata + spec)
        return self._handle_response(response)

    def delete_instance(self, instance_name: str) -> Dict[str, Any]:
        """Delete an instance.

        Args:
            instance_name: Name of the instance to delete

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("DELETE", f"/instances/{instance_name}")
        return self._handle_response(response)

    def list_instances(self) -> List[Dict[str, Any]]:
        """List all instances.

        Returns:
            List of instances

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("GET", "/instances")
        return self._handle_response(response)

    # Solution Operations (for deployments)

    def apply_deployment(self, deployment_spec: Dict[str, Any]) -> Dict[str, Any]:
        """Apply a deployment (POST to /solution/instances).

        Args:
            deployment_spec: Deployment specification dictionary

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("POST", "/solution/instances", json=deployment_spec)
        return self._handle_response(response)

    def get_deployment_components(self) -> Dict[str, Any]:
        """Get deployment components (GET /solution/instances).

        Returns:
            Components data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("GET", "/solution/instances")
        return self._handle_response(response)

    def delete_deployment_components(self) -> Dict[str, Any]:
        """Delete deployment components (DELETE /solution/instances).

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        response = self._make_request("DELETE", "/solution/instances")
        return self._handle_response(response)

    def reconcile_solution(
        self, deployment_spec: Dict[str, Any], delete: bool = False
    ) -> Dict[str, Any]:
        """Direct reconcile/delete deployment (POST to /solution/reconcile).

        Args:
            deployment_spec: Deployment specification dictionary
            delete: Whether this is a delete operation

        Returns:
            API response data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        params = {"delete": "true"} if delete else {}

        response = self._make_request(
            "POST", "/solution/reconcile", json=deployment_spec, params=params
        )

        return self._handle_response(response)

    def get_instance_status(self, instance_name: str) -> Dict[str, Any]:
        """Get instance status (GET /solution/queue).

        Args:
            instance_name: Name of the instance

        Returns:
            Instance status data

        Raises:
            SymphonyAPIError: If request fails
        """
        self._ensure_authenticated()

        params = {"instance": instance_name}

        response = self._make_request("GET", "/solution/queue", params=params)
        return self._handle_response(response)

    # Utility Methods

    def health_check(self) -> bool:
        """Perform a basic health check of the Symphony API.

        Returns:
            True if API is accessible, False otherwise
        """
        try:
            # Try a simple endpoint that doesn't require auth
            response = self._make_request("GET", "/greetings")
            return response.status_code == 200
        except Exception as e:
            self.logger.error(f"Health check failed: {e}")
            return False


__all__ = [
    "SymphonyAPI",
    "SymphonyAPIError",
]
