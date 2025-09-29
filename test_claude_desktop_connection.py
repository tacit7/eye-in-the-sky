#!/usr/bin/env python3

import json
import subprocess
import sys
import time
import threading
import select

def test_claude_desktop_connection():
    """Test exactly what Claude Desktop sees when connecting"""
    print("Testing Claude Desktop connection behavior...")

    # Use exact command from config
    cmd = [
        "/Users/urielmaldonado/projects/eye-in-the-sky/bin/eye-in-the-sky",
        "-db",
        "/Users/urielmaldonado/projects/eye-in-the-sky/data/agents.db"
    ]

    print(f"Command: {' '.join(cmd)}")

    # Start process
    try:
        process = subprocess.Popen(
            cmd,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=0
        )

        print(f"Process started with PID: {process.pid}")

        # Send initialize immediately
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

        print("Sending initialize request...")
        process.stdin.write(json.dumps(init_request) + '\n')
        process.stdin.flush()

        # Try to read response with timeout
        print("Waiting for response...")

        # Use select for timeout
        ready, _, _ = select.select([process.stdout], [], [], 5.0)

        if ready:
            response = process.stdout.readline()
            print(f"Response: {response.strip()}")

            # Try to parse
            try:
                parsed = json.loads(response.strip())
                print(f"Successfully parsed JSON response")
                print(f"Server info: {parsed.get('result', {}).get('serverInfo', 'Unknown')}")
            except:
                print("Failed to parse response as JSON")
        else:
            print("No response received within 5 seconds")

        # Check if process is still alive
        poll = process.poll()
        if poll is None:
            print("Process is still running")

            # Send notifications/initialized
            init_notif = {
                "jsonrpc": "2.0",
                "method": "notifications/initialized",
                "params": {}
            }

            print("Sending initialized notification...")
            process.stdin.write(json.dumps(init_notif) + '\n')
            process.stdin.flush()

            # Keep connection alive for a bit
            print("Keeping connection alive for 10 seconds...")
            time.sleep(10)

            # Check if still alive
            poll = process.poll()
            if poll is None:
                print("✅ Process stayed alive - connection looks good")
            else:
                print(f"❌ Process died with code: {poll}")
        else:
            print(f"❌ Process already died with code: {poll}")

        # Read any stderr
        try:
            stderr_data = process.stderr.read()
            if stderr_data:
                print(f"Stderr: {stderr_data}")
        except:
            pass

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
    test_claude_desktop_connection()