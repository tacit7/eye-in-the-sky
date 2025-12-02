#!/usr/bin/env python3
"""
Minimal MCP stdio client to initialize, list tools, and call a tool
for any MCP server command (e.g., Eye-in-the-Sky or Coqui TTS).

Usage examples:

  # List tools from Eye-in-the-Sky using a workspace DB
  python3 scripts/mcp_client.py \
    --cmd "./bin/eye-in-the-sky --db ./data/eits.db" \
    --list-tools

  # Call i-instructions
  python3 scripts/mcp_client.py \
    --cmd "./bin/eye-in-the-sky --db ./data/eits.db" \
    --call i-instructions --params '{}'

  # Call i-start-session
  python3 scripts/mcp_client.py \
    --cmd "./bin/eye-in-the-sky --db ./data/eits.db" \
    --call i-start-session \
    --params '{"session_id":"abc123","description":"Test session"}'

  # List tools from Coqui TTS
  python3 scripts/mcp_client.py \
    --cmd "/Users/urielmaldonado/.nvm/versions/node/v22.19.0/bin/node /Users/urielmaldonado/projects/coqui-ai-TTS/mcp/server.js" \
    --list-tools

  # Call Coqui TTS speak
  python3 scripts/mcp_client.py \
    --cmd "/Users/urielmaldonado/.nvm/versions/node/v22.19.0/bin/node /Users/urielmaldonado/projects/coqui-ai-TTS/mcp/server.js" \
    --call coqui-speak \
    --params '{"phrase":"Hello from MCP"}'
"""

import argparse
import json
import shlex
import subprocess
import sys
import time


def read_json_line(stream, timeout=10.0):
    """Read a single line from stream and parse JSON within timeout seconds."""
    start = time.time()
    buf = b""
    while time.time() - start < timeout:
        ch = stream.read(1)
        if not ch:
            break
        buf += ch
        if ch == b"\n":
            line = buf.decode("utf-8", errors="replace").strip()
            if not line:
                buf = b""
                continue
            try:
                return json.loads(line)
            except json.JSONDecodeError:
                # Not a valid JSON line; continue
                buf = b""
                continue
    return None


def main():
    ap = argparse.ArgumentParser(description="Minimal MCP stdio client")
    ap.add_argument("--cmd", required=True, help="Command to run (quoted string)")
    ap.add_argument("--list-tools", action="store_true", help="List tools and exit")
    ap.add_argument("--call", help="Tool name to call")
    ap.add_argument("--params", default="{}", help="JSON object of tool arguments")
    ap.add_argument("--timeout", type=float, default=10.0, help="Read timeout seconds")
    args = ap.parse_args()

    cmd = shlex.split(args.cmd)

    # Start server process
    proc = subprocess.Popen(
        cmd,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    try:
        # Initialize
        init_req = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {"roots": {"listChanged": True}},
                "clientInfo": {"name": "codex-cli", "version": "1.0"},
            },
        }
        proc.stdin.write((json.dumps(init_req) + "\n").encode("utf-8"))
        proc.stdin.flush()
        init_res = read_json_line(proc.stdout, timeout=args.timeout)
        if not init_res:
            print("Error: no initialize response", file=sys.stderr)
            return 2

        # Send initialized notification
        initialized = {"jsonrpc": "2.0", "method": "notifications/initialized", "params": {}}
        proc.stdin.write((json.dumps(initialized) + "\n").encode("utf-8"))
        proc.stdin.flush()

        next_id = 2

        if args.list_tools:
            req = {"jsonrpc": "2.0", "id": next_id, "method": "tools/list", "params": {}}
            next_id += 1
            proc.stdin.write((json.dumps(req) + "\n").encode("utf-8"))
            proc.stdin.flush()
            res = read_json_line(proc.stdout, timeout=args.timeout)
            print(json.dumps(res, indent=2))
            return 0

        if args.call:
            try:
                call_args = json.loads(args.params)
            except Exception as e:
                print(f"Invalid --params JSON: {e}", file=sys.stderr)
                return 2

            req = {
                "jsonrpc": "2.0",
                "id": next_id,
                "method": "tools/call",
                "params": {"name": args.call, "arguments": call_args},
            }
            proc.stdin.write((json.dumps(req) + "\n").encode("utf-8"))
            proc.stdin.flush()
            res = read_json_line(proc.stdout, timeout=args.timeout)
            print(json.dumps(res, indent=2))
            return 0

        # If no specific action, just report initialized
        print(json.dumps(init_res, indent=2))
        return 0

    finally:
        try:
            if proc.stdin:
                proc.stdin.close()
        except Exception:
            pass
        try:
            proc.terminate()
        except Exception:
            pass
        try:
            proc.wait(timeout=2)
        except Exception:
            proc.kill()


if __name__ == "__main__":
    sys.exit(main())

