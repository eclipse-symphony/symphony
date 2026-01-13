#!/usr/bin/env python3
"""
Simplified unit tests for Symphony API client focusing on actual methods.
"""

import unittest
from unittest.mock import Mock, patch

import requests

from symphony_sdk.api_client import SymphonyAPI, SymphonyAPIError


class TestSymphonyAPIError(unittest.TestCase):
    """Test cases for SymphonyAPIError exception."""

    def test_symphony_api_error_creation(self):
        """Test SymphonyAPIError creation."""
        error = SymphonyAPIError("Test error message")
        self.assertEqual(str(error), "Test error message")
        self.assertIsNone(error.status_code)
        self.assertIsNone(error.response_text)

    def test_symphony_api_error_with_details(self):
        """Test SymphonyAPIError creation with details."""
        error = SymphonyAPIError("API request failed", status_code=404, response_text="Not Found")
        self.assertEqual(str(error), "API request failed")
        self.assertEqual(error.status_code, 404)


class TestSymphonyAPI(unittest.TestCase):
    """Test cases for the SymphonyAPI REST client."""

    def setUp(self):
        """Set up test fixtures."""
        self.base_url = "https://symphony.example.com"
        self.username = "testuser"
        self.password = "testpass"
        self.client = SymphonyAPI(
            base_url=self.base_url, username=self.username, password=self.password, timeout=10.0
        )

    def tearDown(self):
        """Clean up after tests."""
        if self.client:
            self.client.close()

    def test_client_initialization(self):
        """Test SymphonyAPI initialization."""
        self.assertEqual(self.client.base_url, self.base_url)
        self.assertEqual(self.client.username, self.username)
        self.assertEqual(self.client.password, self.password)
        self.assertEqual(self.client.timeout, 10.0)

    def test_client_context_manager(self):
        """Test SymphonyAPI as context manager."""
        with SymphonyAPI(self.base_url, self.username, self.password) as client:
            self.assertIsInstance(client, SymphonyAPI)
        # Client should be closed after context exit

    @patch("requests.Session.request")
    def test_make_request_basic_functionality(self, mock_request):
        """Test basic request functionality."""
        # Mock successful response
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.text = "OK"
        mock_request.return_value = mock_response

        response = self.client._make_request("GET", "/api/test")

        self.assertEqual(response.status_code, 200)
        self.assertTrue(mock_request.called)

    @patch("requests.Session.request")
    def test_make_request_timeout_handling(self, mock_request):
        """Test request timeout handling."""
        mock_request.side_effect = requests.exceptions.Timeout("Request timed out")

        with self.assertRaises(SymphonyAPIError) as context:
            self.client._make_request("GET", "/api/test")

        self.assertIn("request failed", str(context.exception).lower())

    @patch.object(SymphonyAPI, "_handle_response")
    @patch.object(SymphonyAPI, "_make_request")
    def test_authenticate_basic(self, mock_make_request, mock_handle_response):
        """Test basic authentication functionality."""
        # Mock successful auth response with correct key name
        mock_response = Mock()
        mock_make_request.return_value = mock_response
        mock_handle_response.return_value = {"accessToken": "test_token"}

        token = self.client.authenticate()

        self.assertEqual(token, "test_token")
        self.assertEqual(self.client._access_token, "test_token")

    @patch.object(SymphonyAPI, "_handle_response")
    @patch.object(SymphonyAPI, "_make_request")
    def test_register_target_basic(self, mock_make_request, mock_handle_response):
        """Test basic target registration."""
        # Set up authentication
        self.client._access_token = "test_token"

        mock_response = Mock()
        mock_make_request.return_value = mock_response
        mock_handle_response.return_value = {"status": "registered"}

        target_name = "test-target"
        target_spec = {"type": "device", "location": "datacenter1"}

        result = self.client.register_target(target_name, target_spec)

        self.assertEqual(result, {"status": "registered"})
        self.assertTrue(mock_make_request.called)

    @patch.object(SymphonyAPI, "_handle_response")
    @patch.object(SymphonyAPI, "_make_request")
    def test_get_target_basic(self, mock_make_request, mock_handle_response):
        """Test basic target retrieval."""
        # Set up authentication
        self.client._access_token = "test_token"

        mock_response = Mock()
        mock_make_request.return_value = mock_response
        mock_handle_response.return_value = {
            "name": "test-target",
            "status": "active",
            "spec": {"type": "device"},
        }

        target_name = "test-target"
        result = self.client.get_target(target_name)

        self.assertEqual(result["name"], target_name)
        self.assertEqual(result["status"], "active")

    @patch.object(SymphonyAPI, "_handle_response")
    @patch.object(SymphonyAPI, "_make_request")
    def test_list_targets_basic(self, mock_make_request, mock_handle_response):
        """Test basic target listing."""
        # Set up authentication
        self.client._access_token = "test_token"

        mock_response = Mock()
        mock_make_request.return_value = mock_response
        mock_handle_response.return_value = {
            "targets": [
                {"name": "target1", "status": "active"},
                {"name": "target2", "status": "inactive"},
            ]
        }

        result = self.client.list_targets()

        self.assertIn("targets", result)
        self.assertEqual(len(result["targets"]), 2)

    @patch.object(SymphonyAPI, "_handle_response")
    @patch.object(SymphonyAPI, "_make_request")
    def test_unregister_target_basic(self, mock_make_request, mock_handle_response):
        """Test basic target unregistration."""
        # Set up authentication
        self.client._access_token = "test_token"

        mock_response = Mock()
        mock_make_request.return_value = mock_response
        mock_handle_response.return_value = {"status": "unregistered"}

        target_name = "test-target"

        result = self.client.unregister_target(target_name)

        self.assertEqual(result, {"status": "unregistered"})


if __name__ == "__main__":
    unittest.main()
