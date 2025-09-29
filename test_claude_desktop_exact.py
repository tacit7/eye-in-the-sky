#!/usr/bin/env python3

import json
import subprocess
import sys
import time

def test_exactly_like_claude_desktop():
    """Test the MCP server exactly like Claude Desktop would"""
    print("Testing Eye in the Sky MCP Server exactly like Claude Desktop...")

    # Use the exact command from Claude Desktop config
    cmd = [
        "/Users/urielmaldonado/projects/eye-in-the-sky/bin/eye-in-the-sky",
        "-db",
        "/Users/urielmaldonado/projects/eye-in-the-sky/data/agents.db"
    ]

    print(f"Running command: {' '.join(cmd)}")

    # Start the server process
    process = subprocess.Popen(
        cmd,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=0  # Unbuffered like Claude Desktop
    )

    # Give the server time to start
    time.sleep(0.5)

    try:
        # Exact initialization sequence Claude Desktop uses
        print("\n1. Sending Claude Desktop initialization...")
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

        request_str = json.dumps(init_request) + '\n'
        print(f"Sending: {request_str.strip()}")

        process.stdin.write(request_str)
        process.stdin.flush()

        # Read the response
        response_line = process.stdout.readline()
        print(f"Response: {response_line.strip()}")

        if response_line:
            try:
                response = json.loads(response_line.strip())
                print(f"Parsed response: {json.dumps(response, indent=2)}")
            except json.JSONDecodeError as e:
                print(f"Failed to parse response as JSON: {e}")

        # Send notifications/initialized
        print("\n2. Sending initialized notification...")
        init_notification = {
            "jsonrpc": "2.0",
            "method": "notifications/initialized",
            "params": {}
        }

        notif_str = json.dumps(init_notification) + '\n'
        print(f"Sending: {notif_str.strip()}")

        process.stdin.write(notif_str)
        process.stdin.flush()

        # Test tools/list
        print("\n3. Testing tools/list...")
        tools_request = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/list",
            "params": {}
        }

        tools_str = json.dumps(tools_request) + '\n'
        print(f"Sending: {tools_str.strip()}")

        process.stdin.write(tools_str)
        process.stdin.flush()

        # Read the response
        response_line = process.stdout.readline()
        print(f"Tools response: {response_line.strip()}")

        if response_line:
            try:
                response = json.loads(response_line.strip())
                tools = response.get("result", {}).get("tools", [])
                print(f"Found {len(tools)} tools:")
                for tool in tools:
                    print(f"  - {tool.get('name')}: {tool.get('description')}")

                # Check if register_claude_desktop_agent is in the list
                agent_tool = next((t for t in tools if t.get('name') == 'register_claude_desktop_agent'), None)
                if agent_tool:
                    print(f"\n✅ Found register_claude_desktop_agent tool!")
                    print(f"   Description: {agent_tool.get('description')}")
                else:
                    print(f"\n❌ register_claude_desktop_agent tool not found!")

            except json.JSONDecodeError as e:
                print(f"Failed to parse tools response as JSON: {e}")

        # Test the problematic tool call
        print("\n4. Testing register_claude_desktop_agent call...")
        register_request = {
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {
                "name": "register_claude_desktop_agent",
                "arguments": {
                    "description": "Testing Claude Desktop integration",
                    "project_name": "Test Project"
                }
            }
        }

        register_str = json.dumps(register_request) + '\n'
        print(f"Sending: {register_str.strip()}")

        process.stdin.write(register_str)
        process.stdin.flush()

        # Read the response
        response_line = process.stdout.readline()
        print(f"Register response: {response_line.strip()}")

        if response_line:
            try:
                response = json.loads(response_line.strip())
                print(f"Parsed register response: {json.dumps(response, indent=2)}")
            except json.JSONDecodeError as e:
                print(f"Failed to parse register response as JSON: {e}")

    except Exception as e:
        print(f"Error during testing: {e}")

    finally:
        # Read any remaining stderr
        stderr_output = process.stderr.read()
        if stderr_output:
            print(f"\nServer stderr:\n{stderr_output}")

        # Clean shutdown
        try:
            process.stdin.close()
        except:
            pass

        process.wait(timeout=5)
        print(f"\nProcess finished with return code: {process.returncode}")

if __name__ == "__main__":
    test_exactly_like_claude_desktop()