#!/usr/bin/env python3

import json
import subprocess
import time

def test_i_prefix():
    """Test that all tools now have i- prefix"""
    print("Testing Eye in the Sky MCP Server for i- prefixed tools...")

    cmd = [
        "/Users/urielmaldonado/projects/eye-in-the-sky/bin/eye-in-the-sky",
        "-db",
        "/Users/urielmaldonado/projects/eye-in-the-sky/data/agents.db"
    ]

    process = subprocess.Popen(
        cmd,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=0
    )

    try:
        # Send initialize
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

        # Read initialization response
        response_line = process.stdout.readline()
        print(f"Init response: {response_line.strip()}")

        # Send notifications/initialized
        init_notif = {
            "jsonrpc": "2.0",
            "method": "notifications/initialized",
            "params": {}
        }

        process.stdin.write(json.dumps(init_notif) + '\n')
        process.stdin.flush()

        # Test tools/list
        tools_request = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/list",
            "params": {}
        }

        process.stdin.write(json.dumps(tools_request) + '\n')
        process.stdin.flush()

        # Read tools response
        response_line = process.stdout.readline()
        print(f"Tools response: {response_line.strip()}")

        if response_line:
            try:
                response = json.loads(response_line.strip())
                tools = response.get("result", {}).get("tools", [])
                print(f"\nFound {len(tools)} tools:")
                for tool in tools:
                    name = tool.get('name', 'Unknown')
                    desc = tool.get('description', 'No description')
                    prefix = "✅ i-" if name.startswith('i-') else "❌ no prefix"
                    print(f"  {prefix} {name}: {desc}")

                # Check if all tools have i- prefix
                i_prefixed = [t for t in tools if t.get('name', '').startswith('i-')]
                print(f"\n📊 Summary: {len(i_prefixed)}/{len(tools)} tools have i- prefix")

                if len(i_prefixed) == len(tools):
                    print("✅ All tools successfully have i- prefix!")
                else:
                    print("⚠️  Some tools missing i- prefix")

            except json.JSONDecodeError as e:
                print(f"Failed to parse tools response: {e}")

    except Exception as e:
        print(f"Error: {e}")
    finally:
        try:
            process.terminate()
            process.wait(timeout=2)
        except:
            try:
                process.kill()
            except:
                pass

if __name__ == "__main__":
    test_i_prefix()