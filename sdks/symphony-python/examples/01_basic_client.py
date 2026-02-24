#!/usr/bin/env python3
"""Basic Symphony API Client Usage Example

This example demonstrates how to:
1. Initialize the Symphony API client
2. Authenticate with the Symphony API
3. Perform basic health checks
4. Use the client as a context manager
"""

from symphony_sdk import SymphonyAPI


def main():
    # Initialize the Symphony API client
    # Replace with your actual Symphony instance URL and credentials
    base_url = "http://localhost:8082/v1alpha2"
    username = "admin"
    password = ""

    # Method 1: Using context manager (recommended)
    # The context manager ensures the session is properly closed
    print("Example 1: Using context manager")
    with SymphonyAPI(base_url, username, password) as client:
        # Authentication happens automatically when needed
        print(f"Connected to: {client.base_url}")

        # Perform a health check
        if client.health_check():
            print("✓ Symphony API is healthy")
        else:
            print("✗ Symphony API health check failed")

    print("\n" + "=" * 60 + "\n")

    # Method 2: Manual session management
    print("Example 2: Manual session management")
    client = SymphonyAPI(base_url=base_url, username=username, password=password, timeout=10.0)

    try:
        # Explicitly authenticate (Optional. Auth happens automatically)
        token = client.authenticate()
        print("✓ Authenticated successfully")
        print(f"  Token (first 20 chars): {token[:20]}...")
    except Exception as e:
        print(f"✗ Authentication failed: {e}")
    finally:
        # Close the client when done
        client.close()
        print("✓ Client closed")
        return True


if __name__ == "__main__":
    print("=" * 60)
    print("Symphony SDK - Basic Client Usage")
    print("=" * 60 + "\n")

    print("NOTE: Update the credentials in this script before running!\n")

    # Uncomment the line below after updating credentials
    # main()

    print("Update base_url, username, and password in the script,")
    print("then uncomment the main() call to run this example.")
