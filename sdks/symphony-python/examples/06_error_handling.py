#!/usr/bin/env python3
"""Error Handling and Best Practices Example

This example demonstrates:
1. Proper error handling with SymphonyAPIError
2. Retry logic for transient failures
3. Logging and debugging
4. Timeout handling
5. Validation best practices
"""

import logging
import time
from typing import Any, Dict, Optional

from symphony_sdk import SymphonyAPI, SymphonyAPIError

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


def example_basic_error_handling():
    """Example: Basic error handling."""
    print("1. Basic Error Handling")
    print("-" * 40)

    base_url = "http://localhost:8082/v1alpha2"
    username = "admin"
    password = ""

    try:
        with SymphonyAPI(base_url, username, password) as client:
            # Attempt to get a non-existent target
            target = client.get_target("non-existent-target")

    except SymphonyAPIError as e:
        print("✗ Caught SymphonyAPIError")
        print(f"  Message: {e}")
        print(f"  Status Code: {e.status_code}")

        # Handle specific HTTP status codes
        if e.status_code == 404:
            print("  → Target not found")
        elif e.status_code == 401:
            print("  → Authentication failed - check credentials")
        elif e.status_code == 403:
            print("  → Permission denied")
        elif e.status_code == 500:
            print("  → Server error - try again later")
        else:
            print(f"  → Unexpected error: {e.status_code}")

        # Response text may contain additional details
        if e.response_text:
            print(f"  Response details: {e.response_text[:200]}")


def retry_with_backoff(
    func,
    max_retries: int = 3,
    initial_delay: float = 1.0,
    backoff_factor: float = 2.0,
    retryable_codes: tuple = (500, 502, 503, 504),
):
    """Retry a function with exponential backoff.

    Args:
        func: Function to retry
        max_retries: Maximum number of retry attempts
        initial_delay: Initial delay between retries in seconds
        backoff_factor: Multiplier for delay after each retry
        retryable_codes: HTTP status codes that should trigger retry
    """
    delay = initial_delay
    last_exception = None

    for attempt in range(max_retries + 1):
        try:
            return func()
        except SymphonyAPIError as e:
            last_exception = e

            # Don't retry client errors (4xx) except for specific cases
            if e.status_code and 400 <= e.status_code < 500:
                if e.status_code not in (408, 429):  # Timeout, Too Many Requests
                    raise

            # Don't retry if not a retryable server error
            if e.status_code and e.status_code not in retryable_codes:
                raise

            if attempt < max_retries:
                logger.warning(f"Attempt {attempt + 1} failed: {e}. Retrying in {delay}s...")
                time.sleep(delay)
                delay *= backoff_factor
            else:
                logger.error(f"All {max_retries} retry attempts failed")

    raise last_exception


def example_retry_logic():
    """Example: Implementing retry logic."""
    print("\n2. Retry Logic with Exponential Backoff")
    print("-" * 40)

    base_url = "http://localhost:8082/v1alpha2"
    username = "admin"
    password = ""

    with SymphonyAPI(base_url, username, password) as client:

        def create_target():
            """Function to retry."""
            return client.register_target(
                "my-target", {"displayName": "My Target", "properties": {"location": "dc1"}}
            )

        try:
            # Retry with exponential backoff
            result = retry_with_backoff(create_target, max_retries=3, initial_delay=1.0)
            print("✓ Target created successfully (possibly after retries)")

        except SymphonyAPIError as e:
            print(f"✗ Failed after all retries: {e}")


def example_connection_validation():
    """Example: Validating connection before operations."""
    print("\n3. Connection Validation")
    print("-" * 40)

    base_url = "http://localhost:8082/v1alpha2"
    username = "admin"
    password = ""

    try:
        with SymphonyAPI(base_url, username, password) as client:
            # Validate connection first
            print("  Checking connection...")
            if not client.health_check():
                print("  ✗ Cannot connect to Symphony API")
                print("  Check:")
                print(f"    - URL is correct: {base_url}")
                print("    - Symphony is running")
                print("    - Network connectivity")
                return

            print("  ✓ Connection successful")

            # Validate authentication
            print("  Checking authentication...")
            try:
                client.authenticate()
                print("  ✓ Authentication successful")
            except SymphonyAPIError as e:
                if e.status_code == 401:
                    print("  ✗ Authentication failed")
                    print("  Check:")
                    print("    - Username is correct")
                    print("    - Password is correct")
                    print("    - User has necessary permissions")
                return

            # Now safe to perform operations
            print("  ✓ Ready to perform operations")

    except Exception as e:
        print(f"✗ Unexpected error: {e}")


def example_timeout_handling():
    """Example: Handling timeouts."""
    print("\n4. Timeout Handling")
    print("-" * 40)

    base_url = "http://localhost:8082/v1alpha2"
    username = "admin"
    password = ""

    # Create client with custom timeout
    client = SymphonyAPI(
        base_url,
        username,
        password,
        timeout=5.0,  # 5 second timeout
    )

    try:
        # This might timeout for slow operations
        result = client.list_targets()
        print("✓ Operation completed within timeout")

    except SymphonyAPIError as e:
        if "timeout" in str(e).lower():
            print("✗ Operation timed out")
            print("  Consider:")
            print("    - Increasing timeout value")
            print("    - Checking network connectivity")
            print("    - Checking server performance")
        else:
            print(f"✗ Error: {e}")

    finally:
        client.close()


def example_input_validation():
    """Example: Input validation best practices."""
    print("\n5. Input Validation")
    print("-" * 40)

    def validate_target_name(name: str) -> bool:
        """Validate target name format."""
        if not name:
            logger.error("Target name cannot be empty")
            return False

        if len(name) > 253:
            logger.error("Target name too long (max 253 characters)")
            return False

        # Symphony typically uses lowercase DNS names
        if not name.islower():
            logger.warning("Target name should be lowercase")

        # Check for invalid characters
        import re

        if not re.match(r"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", name):
            logger.error("Target name contains invalid characters")
            return False

        return True

    # Test validation
    test_names = [
        "valid-target-001",
        "Invalid-Name",  # Has uppercase
        "",  # Empty
        "target_with_underscore",  # Has underscore
        "valid-target",
    ]

    for name in test_names:
        is_valid = validate_target_name(name)
        status = "✓" if is_valid else "✗"
        print(f"  {status} '{name}': {'valid' if is_valid else 'invalid'}")


def example_comprehensive_error_handler():
    """Example: Comprehensive error handling wrapper."""
    print("\n6. Comprehensive Error Handler")
    print("-" * 40)

    def safe_api_call(client: SymphonyAPI, operation: str, func, *args, **kwargs):
        """Safely execute an API call with comprehensive error handling.

        Args:
            client: SymphonyAPI client
            operation: Description of operation for logging
            func: Function to call
            *args, **kwargs: Arguments for the function
        """
        try:
            logger.info(f"Starting operation: {operation}")
            result = func(*args, **kwargs)
            logger.info(f"✓ {operation} succeeded")
            return result

        except SymphonyAPIError as e:
            logger.error(f"✗ {operation} failed: {e}")

            # Categorize error
            if e.status_code:
                if e.status_code == 400:
                    logger.error("  Bad request - check input parameters")
                elif e.status_code == 401:
                    logger.error("  Authentication failed - check credentials")
                elif e.status_code == 403:
                    logger.error("  Permission denied - check user permissions")
                elif e.status_code == 404:
                    logger.error("  Resource not found")
                elif e.status_code == 409:
                    logger.error("  Conflict - resource may already exist")
                elif 500 <= e.status_code < 600:
                    logger.error("  Server error - may be transient, consider retry")

            # Log details for debugging
            if e.response_text:
                logger.debug(f"  Response: {e.response_text}")

            raise

        except Exception as e:
            logger.error(f"✗ {operation} failed with unexpected error: {e}")
            raise

    # Example usage
    base_url = "http://localhost:8082/v1alpha2"
    username = "admin"
    password = ""

    with SymphonyAPI(base_url, username, password) as client:
        # Use the wrapper
        try:
            targets = safe_api_call(client, "List targets", client.list_targets)
            print(f"  Found {len(targets.get('items', []))} targets")
        except Exception:
            print("  Operation failed - check logs for details")


def example_graceful_degradation():
    """Example: Graceful degradation when services are unavailable."""
    print("\n7. Graceful Degradation")
    print("-" * 40)

    base_url = "http://localhost:8082/v1alpha2"
    username = "admin"
    password = ""

    class TargetManager:
        """Manager with graceful degradation."""

        def __init__(self, client: SymphonyAPI):
            self.client = client
            self.cache: Dict[str, Any] = {}
            self.degraded_mode = False

        def get_target(self, name: str) -> Optional[Dict[str, Any]]:
            """Get target with fallback to cache."""
            if not self.degraded_mode:
                try:
                    target = self.client.get_target(name)
                    self.cache[name] = target
                    return target
                except SymphonyAPIError as e:
                    logger.warning(f"API call failed: {e}")
                    logger.info("Entering degraded mode, using cache")
                    self.degraded_mode = True

            # Fallback to cache
            if name in self.cache:
                logger.info(f"Returning cached data for {name}")
                return self.cache[name]

            logger.error(f"No cached data available for {name}")
            return None

    try:
        with SymphonyAPI(base_url, username, password) as client:
            manager = TargetManager(client)

            # Try to get target (will cache if successful)
            target = manager.get_target("my-target")
            if target:
                print("  ✓ Got target (from API or cache)")
            else:
                print("  ✗ Target not available")

    except Exception as e:
        print(f"  ✗ Error: {e}")


if __name__ == "__main__":
    print("=" * 60)
    print("Symphony SDK - Error Handling and Best Practices")
    print("=" * 60 + "\n")

    # Note: These examples demonstrate error handling patterns
    # They won't actually run without valid Symphony credentials

    print("These examples demonstrate error handling patterns.")
    print("Update credentials and uncomment function calls to test.\n")

    # Uncomment to run examples:
    example_basic_error_handling()
    example_retry_logic()
    example_connection_validation()
    example_timeout_handling()
    example_input_validation()
    example_comprehensive_error_handler()
    example_graceful_degradation()

    print("\n" + "=" * 60)
    print("Review the code to see error handling patterns")
    print("=" * 60)
