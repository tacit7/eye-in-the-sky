#!/usr/bin/env python3

import json
import subprocess
import sys
import time
import threading

def test_mcp_interactive():
    """Test the MCP server with an interactive session"""
    print("Testing MCP Interactive Protocol Communication...")

    # Start the server process
    process = subprocess.Popen(
        ['./bin/eye-in-the-sky'],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )

    def read_output(stream, name):
        """Read output from stream in a separate thread"""
        while True:
            line = stream.readline()
            if not line:
                break
            print(f"{name}: {line.strip()}")

    # Start threads to read stdout and stderr
    stdout_thread = threading.Thread(target=read_output, args=(process.stdout, "STDOUT"))
    stderr_thread = threading.Thread(target=read_output, args=(process.stderr, "STDERR"))
    stdout_thread.daemon = True
    stderr_thread.daemon = True
    stdout_thread.start()
    stderr_thread.start()

    # Give the server time to start
    time.sleep(1)

    # Send initialization
    print("\n1. Sending initialization...")
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

    try:
        process.stdin.write(json.dumps(init_request) + '\n')
        process.stdin.flush()
        print(f"Sent: {json.dumps(init_request)}")

        # Wait for response
        time.sleep(2)

        # Send tools/list
        print("\n2. Sending tools/list...")
        list_request = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/list",
            "params": {}
        }

        process.stdin.write(json.dumps(list_request) + '\n')
        process.stdin.flush()
        print(f"Sent: {json.dumps(list_request)}")

        # Wait for response
        time.sleep(2)

        # Send notifications/initialized
        print("\n3. Sending notifications/initialized...")
        initialized_notification = {
            "jsonrpc": "2.0",
            "method": "notifications/initialized",
            "params": {}
        }

        process.stdin.write(json.dumps(initialized_notification) + '\n')
        process.stdin.flush()
        print(f"Sent: {json.dumps(initialized_notification)}")

        # Wait
        time.sleep(2)

        # Test a tool call
        print("\n4. Testing tool call...")
        tool_request = {
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {
                "name": "help",
                "arguments": {}
            }
        }

        process.stdin.write(json.dumps(tool_request) + '\n')
        process.stdin.flush()
        print(f"Sent: {json.dumps(tool_request)}")

        # Wait for response
        time.sleep(3)

    except BrokenPipeError:
        print("Server closed the connection")

    finally:
        # Clean shutdown
        process.stdin.close()
        process.wait()
        print(f"\nProcess finished with return code: {process.returncode}")

if __name__ == "__main__":
    test_mcp_interactive()