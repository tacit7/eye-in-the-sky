#!/usr/bin/env python3

import json
import subprocess
import sys
import time
import threading

def test_register_agent():
    """Test the register_claude_desktop_agent specifically"""
    print("Testing register_claude_desktop_agent...")

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

    try:
        # Send initialization
        print("\n1. Initializing...")
        init_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {"roots": {"listChanged": True}},
                "clientInfo": {"name": "claude-desktop", "version": "0.7.1"}
            }
        }
        process.stdin.write(json.dumps(init_request) + '\n')
        process.stdin.flush()
        time.sleep(1)

        # Send notifications/initialized
        print("\n2. Sending initialized notification...")
        initialized_notification = {
            "jsonrpc": "2.0",
            "method": "notifications/initialized",
            "params": {}
        }
        process.stdin.write(json.dumps(initialized_notification) + '\n')
        process.stdin.flush()
        time.sleep(1)

        # Test register_claude_desktop_agent
        print("\n3. Testing register_claude_desktop_agent...")
        register_request = {
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {
                "name": "register_claude_desktop_agent",
                "arguments": {
                    "description": "Testing MCP integration",
                    "project_name": "Eye in the Sky Test"
                }
            }
        }

        process.stdin.write(json.dumps(register_request) + '\n')
        process.stdin.flush()
        print(f"Sent: {json.dumps(register_request)}")

        # Wait for response
        time.sleep(3)

        # Test with agent_id specified
        print("\n4. Testing with specific agent_id...")
        register_request2 = {
            "jsonrpc": "2.0",
            "id": 4,
            "method": "tools/call",
            "params": {
                "name": "register_claude_desktop_agent",
                "arguments": {
                    "agent_id": "test1234",
                    "description": "Testing with specific ID",
                    "project_name": "Eye in the Sky Test",
                    "window_id": "win_test123"
                }
            }
        }

        process.stdin.write(json.dumps(register_request2) + '\n')
        process.stdin.flush()
        print(f"Sent: {json.dumps(register_request2)}")

        # Wait for response
        time.sleep(3)

    except Exception as e:
        print(f"Error: {e}")

    finally:
        # Clean shutdown
        process.stdin.close()
        process.wait()
        print(f"\nProcess finished with return code: {process.returncode}")

if __name__ == "__main__":
    test_register_agent()