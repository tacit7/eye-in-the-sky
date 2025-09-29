#!/usr/bin/env python3

import json
import subprocess
import sys
import time

def send_mcp_request(request):
    """Send an MCP request to the server and get response"""
    process = subprocess.Popen(
        ['./bin/eye-in-the-sky'],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )

    # Send the request
    print(f"Sending: {json.dumps(request)}")
    stdout, stderr = process.communicate(input=json.dumps(request) + '\n')

    print(f"Return code: {process.returncode}")
    print(f"Stdout: {stdout}")
    print(f"Stderr: {stderr}")

    return stdout, stderr, process.returncode

def test_mcp_protocol():
    print("Testing MCP Protocol Communication...")

    # Test 1: Initialize
    print("\n1. Testing MCP Initialize...")
    init_request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {
                "roots": {"listChanged": True}
            },
            "clientInfo": {
                "name": "claude-desktop",
                "version": "0.7.1"
            }
        }
    }

    stdout, stderr, code = send_mcp_request(init_request)

    # Test 2: List tools
    print("\n2. Testing List Tools...")
    list_request = {
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list",
        "params": {}
    }

    stdout, stderr, code = send_mcp_request(list_request)

if __name__ == "__main__":
    test_mcp_protocol()