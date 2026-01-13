#!/usr/bin/env python3
"""
Comprehensive unit tests for Symphony SDK types and enumerations.
"""

import asyncio
import unittest

from symphony_sdk.types import COAConstants, State, get_http_status


class MockTerminable:
    """Mock implementation of Terminable for testing."""

    def __init__(self):
        self.shutdown_called = False

    async def shutdown(self) -> None:
        """Mock shutdown method."""
        self.shutdown_called = True


class TestState(unittest.TestCase):
    """Test cases for State enumeration."""

    def test_state_values(self):
        """Test that State enum has correct values."""
        # Basic states
        self.assertEqual(State.NONE, 0)

        # HTTP Success states
        self.assertEqual(State.OK, 200)
        self.assertEqual(State.ACCEPTED, 202)

        # HTTP Client Error states
        self.assertEqual(State.BAD_REQUEST, 400)
        self.assertEqual(State.UNAUTHORIZED, 401)
        self.assertEqual(State.FORBIDDEN, 403)
        self.assertEqual(State.NOT_FOUND, 404)
        self.assertEqual(State.METHOD_NOT_ALLOWED, 405)
        self.assertEqual(State.CONFLICT, 409)
        self.assertEqual(State.STATUS_UNPROCESSABLE_ENTITY, 422)

        # HTTP Server Error states
        self.assertEqual(State.INTERNAL_ERROR, 500)

        # Config errors
        self.assertEqual(State.BAD_CONFIG, 1000)
        self.assertEqual(State.MISSING_CONFIG, 1001)

        # Operation results
        self.assertEqual(State.UPDATE_FAILED, 8001)
        self.assertEqual(State.DELETE_FAILED, 8002)
        self.assertEqual(State.VALIDATE_FAILED, 8003)
        self.assertEqual(State.UPDATED, 8004)
        self.assertEqual(State.DELETED, 8005)

        # Workflow status
        self.assertEqual(State.RUNNING, 9994)
        self.assertEqual(State.PAUSED, 9995)
        self.assertEqual(State.DONE, 9996)
        self.assertEqual(State.DELAYED, 9997)
        self.assertEqual(State.UNTOUCHED, 9998)
        self.assertEqual(State.NOT_IMPLEMENTED, 9999)

    def test_state_http_success_codes(self):
        """Test HTTP success state codes."""
        success_states = [State.OK, State.ACCEPTED]
        for state in success_states:
            self.assertGreaterEqual(state, 200)
            self.assertLess(state, 300)

    def test_state_http_client_error_codes(self):
        """Test HTTP client error state codes."""
        client_error_states = [
            State.BAD_REQUEST,
            State.UNAUTHORIZED,
            State.FORBIDDEN,
            State.NOT_FOUND,
            State.METHOD_NOT_ALLOWED,
            State.CONFLICT,
            State.STATUS_UNPROCESSABLE_ENTITY,
        ]
        for state in client_error_states:
            self.assertGreaterEqual(state, 400)
            self.assertLess(state, 500)

    def test_state_http_server_error_codes(self):
        """Test HTTP server error state codes."""
        server_error_states = [State.INTERNAL_ERROR]
        for state in server_error_states:
            self.assertGreaterEqual(state, 500)
            self.assertLess(state, 600)

    def test_state_custom_error_codes(self):
        """Test custom Symphony error codes."""
        custom_states = [
            State.BAD_CONFIG,
            State.MISSING_CONFIG,
            State.INVALID_ARGUMENT,
            State.API_REDIRECT,
            State.FILE_ACCESS_ERROR,
            State.SERIALIZATION_ERROR,
            State.DESERIALIZE_ERROR,
            State.DELETE_REQUESTED,
        ]
        for state in custom_states:
            self.assertGreater(state, 999)  # All custom states > 999

    def test_state_from_int(self):
        """Test creating State from integer values."""
        self.assertEqual(State(200), State.OK)
        self.assertEqual(State(404), State.NOT_FOUND)
        self.assertEqual(State(500), State.INTERNAL_ERROR)

    def test_state_int_conversion(self):
        """Test converting State to integer."""
        self.assertEqual(int(State.OK), 200)
        self.assertEqual(int(State.NOT_FOUND), 404)
        self.assertEqual(int(State.INTERNAL_ERROR), 500)


class TestTerminable(unittest.TestCase):
    """Test cases for Terminable protocol."""

    def test_terminable_protocol(self):
        """Test that objects implement Terminable protocol correctly."""
        mock_terminable = MockTerminable()

        # Should have shutdown method
        self.assertTrue(hasattr(mock_terminable, "shutdown"))
        self.assertTrue(callable(mock_terminable.shutdown))

    def test_terminable_shutdown(self):
        """Test Terminable shutdown behavior."""

        async def run_test():
            mock_terminable = MockTerminable()
            self.assertFalse(mock_terminable.shutdown_called)

            await mock_terminable.shutdown()

            self.assertTrue(mock_terminable.shutdown_called)

        # Run the async test
        asyncio.run(run_test())


class TestHttpStatusFunction(unittest.TestCase):
    """Test cases for get_http_status function."""

    def test_get_http_status_success_codes(self):
        """Test get_http_status with success codes."""
        self.assertEqual(get_http_status(200), State.OK)
        self.assertEqual(get_http_status(202), State.ACCEPTED)

    def test_get_http_status_client_error_codes(self):
        """Test get_http_status with client error codes."""
        self.assertEqual(get_http_status(400), State.BAD_REQUEST)
        self.assertEqual(get_http_status(401), State.UNAUTHORIZED)
        self.assertEqual(get_http_status(404), State.NOT_FOUND)

    def test_get_http_status_server_error_codes(self):
        """Test get_http_status with server error codes."""
        self.assertEqual(get_http_status(500), State.INTERNAL_ERROR)

    def test_get_http_status_custom_codes(self):
        """Test get_http_status with custom HTTP codes."""
        # Test other success codes
        self.assertEqual(get_http_status(201), State.OK)
        self.assertEqual(get_http_status(204), State.OK)

        # Test other client error codes
        self.assertEqual(get_http_status(422), State.BAD_REQUEST)

        # Test other server error codes
        self.assertEqual(get_http_status(503), State.INTERNAL_ERROR)


class TestCOAConstants(unittest.TestCase):
    """Test cases for COA constants."""

    def test_coa_constants_exist(self):
        """Test that COA constants are defined."""
        # Test some expected constants exist
        self.assertTrue(hasattr(COAConstants, "COA_META_HEADER"))
        self.assertTrue(hasattr(COAConstants, "TRACING_EXPORTER_CONSOLE"))
        self.assertTrue(hasattr(COAConstants, "PROVIDERS_CONFIG"))

        # Test values are strings
        self.assertEqual(COAConstants.COA_META_HEADER, "COA_META_HEADER")
        self.assertEqual(COAConstants.PROVIDERS_CONFIG, "providers.config")

    def test_coa_constants_types(self):
        """Test that COA constants have correct types."""
        self.assertIsInstance(COAConstants.COA_META_HEADER, str)
        self.assertIsInstance(COAConstants.TRACING_EXPORTER_CONSOLE, str)


if __name__ == "__main__":
    unittest.main()
