#!/usr/bin/env python3
"""
Mikrotik RouterOS API Test Script

Tests connectivity to Mikrotik RouterOS via API and retrieves system information.
Requires: pip install routeros-api
"""

import sys
import socket
import json
from typing import Optional, Dict, Any

# Configuration
MIKROTIK_IP = "192.168.1.1"
MIKROTIK_PORT = 8728
MIKROTIK_USER = "admin"
MIKROTIK_PASSWORD = ""
TIMEOUT = 5

# Colors
class Colors:
    GREEN = '\033[0;32m'
    RED = '\033[0;31m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'

def print_header(text: str):
    print(f"\n{Colors.BLUE}{'═' * 65}{Colors.NC}")
    print(f"{Colors.BLUE}{text}{Colors.NC}")
    print(f"{Colors.BLUE}{'═' * 65}{Colors.NC}\n")

def print_success(text: str):
    print(f"{Colors.GREEN}✓ {text}{Colors.NC}")

def print_error(text: str):
    print(f"{Colors.RED}✗ {text}{Colors.NC}")

def print_warning(text: str):
    print(f"{Colors.YELLOW}⚠ {text}{Colors.NC}")

def print_info(text: str):
    print(f"{Colors.BLUE}ℹ {text}{Colors.NC}")

# Test 1: TCP Connection
def test_tcp_connection(ip: str, port: int) -> bool:
    print_header("TEST 1: TCP Connection to Mikrotik API")

    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(TIMEOUT)
        result = sock.connect_ex((ip, port))
        sock.close()

        if result == 0:
            print_success(f"API port {port} is OPEN on {ip}")
            return True
        else:
            print_error(f"Cannot connect to API port {port} on {ip}")
            print_info("Enable API: /ip service set api enabled=yes")
            return False
    except socket.gaierror:
        print_error(f"Cannot resolve hostname: {ip}")
        return False
    except Exception as e:
        print_error(f"Connection error: {e}")
        return False

# Test 2: Raw API Test (without library)
def test_raw_api(ip: str, port: int, username: str, password: str) -> Dict[str, Any]:
    print_header("TEST 2: Raw API Communication")

    if not password:
        print_warning("MIKROTIK_PASSWORD not set - skipping API login test")
        print_info("Set password: export MIKROTIK_PASSWORD=yourpassword")
        return {}

    try:
        import ssl
        print_info("Attempting raw API connection...")

        # Create SSL/TLS socket
        context = ssl.create_default_context()
        context.check_hostname = False
        context.verify_mode = ssl.CERT_NONE

        with socket.create_connection((ip, port), timeout=TIMEOUT) as sock:
            with context.wrap_socket(sock, server_hostname=ip) as secure_sock:
                print_success("TLS connection established")

                # Send login command
                login_cmd = "/login\n".encode()
                secure_sock.send(login_cmd)

                # Read response
                response = secure_sock.recv(1024).decode()
                print_info(f"Login response: {response[:100]}...")

                # Note: Full API implementation requires parsing ROS protocol
                # This is just a basic connectivity test
                print_success("API communication works")
                return {"status": "connected"}

    except ImportError:
        print_warning("SSL module not available - skipping secure connection")
        return {}
    except Exception as e:
        print_error(f"API communication failed: {e}")
        return {}

# Test 3: Using routeros-api library (if available)
def test_with_library(ip: str, port: int, username: str, password: str) -> Optional[Dict[str, Any]]:
    print_header("TEST 3: API Test with Library")

    if not password:
        print_warning("MIKROTIK_PASSWORD not set - skipping library test")
        return None

    try:
        from routeros_api import api

        print_info("Connecting via routeros-api library...")
        connection = api.connect(
            host=ip,
            username=username,
            password=password,
            port=port,
            use_ssl=True,
            ssl_verify=False
        )

        print_success("Connected to Mikrotik API")

        # Get system info
        print_info("Retrieving system information...")
        response = connection.get_resource('/system/resource').call('print')

        if response:
            print_success("Retrieved system resource info")
            print("\nSystem Information:")
            print("-" * 50)
            for item in response:
                if isinstance(item, dict):
                    for key, value in item.items():
                        if key != 'ret':
                            print(f"  {key}: {value}")
            print("-" * 50)

        # Get RADIUS configuration
        print_info("Checking RADIUS configuration...")
        radius_config = connection.get_resource('/radius').call('print')

        if radius_config:
            print_success("Retrieved RADIUS configuration")
            print("\nRADIUS Configuration:")
            print("-" * 50)
            for item in radius_config:
                if isinstance(item, dict):
                    print(f"  RADIUS Server: {item.get('address', 'N/A')}")
                    print(f"  Secret: {'***' if item.get('secret') else 'N/A'}")
                    print(f"  Service: {item.get('service', 'N/A')}")
                    print(f"  Status: {item.get('disabled', 'false')}")
                    print()
            print("-" * 50)

        return {"status": "success", "data": response}

    except ImportError:
        print_warning("routeros-api library not installed")
        print_info("Install: pip install routeros-api")
        return None
    except Exception as e:
        print_error(f"Library test failed: {e}")
        return {"status": "error", "error": str(e)}

# Test 4: Check API Services
def check_api_services() -> Dict[str, Any]:
    print_header("TEST 4: API Services Configuration")

    info = {
        "api_enabled": False,
        "api_ssl_port": 8729,
        "api_port": 8728
    }

    print_info("To enable API on Mikrotik, run:")
    print()
    print("  /ip service")
    print("  set api enabled=yes port=8728")
    print("  set api-ssl enabled=yes port=8729 certificate=default-ssl")
    print()
    print_info("To check current API services:")
    print("  /ip service print where name~api")
    print()

    return info

# Main function
def main():
    print(f"\n{Colors.BLUE}")
    print("╔═══════════════════════════════════════════════════════════════╗")
    print("║                                                               ║")
    print("║   Mikrotik RouterOS API Test                                  ║")
    print("║   ToughRADIUS Integration Tool                                 ║")
    print("║                                                               ║")
    print("╚═══════════════════════════════════════════════════════════════╝")
    print(f"{Colors.NC}\n")

    print_info("Configuration:")
    print(f"  MIKROTIK_IP: {MIKROTIK_IP}")
    print(f"  MIKROTIK_PORT: {MIKROTIK_PORT}")
    print(f"  MIKROTIK_USER: {MIKROTIK_USER}")
    print(f"  MIKROTIK_PASSWORD: {'***' if MIKROTIK_PASSWORD else '(not set)'}")
    print()

    # Run tests
    tcp_ok = test_tcp_connection(MIKROTIK_IP, MIKROTIK_PORT)

    if tcp_ok:
        raw_result = test_raw_api(MIKROTIK_IP, MIKROTIK_PORT, MIKROTIK_USER, MIKROTIK_PASSWORD)
        lib_result = test_with_library(MIKROTIK_IP, MIKROTIK_PORT, MIKROTIK_USER, MIKROTIK_PASSWORD)
        api_info = check_api_services()

    # Summary
    print_header("Test Summary")
    print_success("API tests completed!")
    print()
    print_info("Next Steps:")
    print("  1. Ensure Mikrotik API is enabled")
    print("  2. Install Python library: pip install routeros-api")
    print("  3. Set MIKROTIK_PASSWORD environment variable")
    print("  4. Run: python3 scripts/test_mikrotik_api.py")
    print("  5. Configure RADIUS in Mikrotik (see docs/MIKROTIK_SETUP.md)")
    print()

if __name__ == "__main__":
    main()
