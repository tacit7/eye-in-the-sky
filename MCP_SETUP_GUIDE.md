# MCP Setup Guide for Claude Code

This guide helps you configure the Eye in the Sky MCP server for Claude Code using the official CLI.

## Overview

Eye in the Sky provides an MCP (Model Context Protocol) server for tracking Claude Code agent sessions. The server runs in stdio mode and integrates seamlessly with Claude Code.

## Quick Fix: “No MCP servers configured”

If Claude Code shows no MCP servers or you see “No MCP server available”, add the server with the Claude CLI (project‑scoped):

```bash
claude mcp add --transport stdio eye-in-the-sky -- /absolute/path/to/eye-in-the-sky/bin/eye-in-the-sky
claude mcp list
claude mcp get eye-in-the-sky
```

Expected output includes “✓ Connected” and “Scope: Local config (private to you in this project)”.

## Prerequisites

1. **Build the MCP server binary:**
   ```bash
   cd /path/to/eye-in-the-sky
   go build -o bin/eye-in-the-sky ./cmd/server
   ```

2. **Verify the binary:**
   ```bash
   ./bin/eye-in-the-sky --help
   ```

   Should output:
   ```
   Eye in the Sky - Claude Code Multi-Agent Management System (MCP Server)
   Usage:
     -db string
       Database path (default: ~/.config/eye-in-the-sky/eits.db)
     -help
       Show help
   ```

## Step 1: Add MCP Server Using Claude CLI (Recommended)

The easiest way to add the MCP server is using the `claude mcp add` command:

```bash
# Add the server (no arguments needed for stdio mode)
claude mcp add --transport stdio eits -- /absolute/path/to/eye-in-the-sky/bin/eye-in-the-sky
```

**Example:**
```bash
claude mcp add --transport stdio eits -- /Users/urielmaldonado/projects/eye-in-the-sky/bin/eye-in-the-sky
```

**Important:** Do NOT include the `-db` flag when running in MCP stdio mode. The binary uses the default database location (`~/.config/eye-in-the-sky/eits.db`) automatically.

Note on naming: You can name the server `eits` or `eye-in-the-sky`. Keep it consistent with your settings and examples.

## Step 2: Verify Installation

Check that the server is registered and healthy:

```bash
claude mcp list
```

You should see:
```
Checking MCP server health...

eits: /path/to/eye-in-the-sky/bin/eye-in-the-sky  - ✓ Connected
```

## Step 3: Get Server Details

View the server configuration:

```bash
claude mcp get eits
```

Output:
```
eits:
  Scope: Local config (private to you in this project)
  Status: ✓ Connected
  Type: stdio
  Command: /path/to/eye-in-the-sky/bin/eye-in-the-sky
  Args:
  Environment:
```

## Alternative: Manual Configuration

If you prefer to manually configure, edit `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "eits": {
      "command": "/absolute/path/to/eye-in-the-sky/bin/eye-in-the-sky",
      "args": [],
      "env": {},
      "instructions": "Eye in the Sky - Agent tracking for Claude Code. Call i-instructions to learn the complete workflow."
    }
  }
}
```

**Note:** The CLI method (`claude mcp add`) is recommended as it handles configuration validation automatically.

## Troubleshooting

### Server Shows "Failed to connect"

If `claude mcp list` shows "✗ Failed to connect":

1. **Check the binary exists:**
   ```bash
   ls -la /path/to/eye-in-the-sky/bin/eye-in-the-sky
   ```

2. **Verify it's executable:**
   ```bash
   chmod +x /path/to/eye-in-the-sky/bin/eye-in-the-sky
   ```

3. **Test MCP stdio mode manually:**
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | /path/to/eye-in-the-sky/bin/eye-in-the-sky
   ```

   Should output JSON-RPC response, not "Eye in the Sky - CLI Mode"

### “No MCP servers configured” or “No MCP server available”

This means Claude Code doesn’t have a registered server for your current project. Fix by adding the server with the CLI (project‑scoped):

```bash
claude mcp add --transport stdio eye-in-the-sky -- /absolute/path/to/eye-in-the-sky/bin/eye-in-the-sky
claude mcp list
```

You should then see the server listed with a ✓ Connected status.

### Common Mistakes

1. **❌ Including `-db` flag:** The `-db` flag is for CLI mode only, not MCP stdio mode
   ```bash
   # WRONG - will fail
   claude mcp add --transport stdio eits -- /path/to/binary -db /path/to/db

   # CORRECT
   claude mcp add --transport stdio eits -- /path/to/binary
   ```

2. **❌ Wrong binary:** Make sure you're using `bin/eye-in-the-sky`, not `bin/eye-in-the-sky-integrated` or other variants

3. **❌ Relative paths:** Always use absolute paths, not relative paths like `./bin/eye-in-the-sky`

4. **⚠️ Mixed configuration sources:** Prefer `claude mcp add` over manually editing config files. If you previously added entries in `~/.claude/mcp.json` or `~/.claude/settings.json`, they may conflict or be ignored depending on scope. The CLI stores project‑scoped configuration in `~/.claude.json` and is the recommended approach.

### Viewing Logs

Check Claude Code debug logs if the server isn't connecting:

```bash
ls -la ~/.claude/debug/
cat ~/.claude/debug/latest | grep -i "mcp\|eits"
```

## Using the MCP Server

Once configured, the MCP tools are available in Claude Code. Call the initialization tool to learn the workflow:

```
Use the mcp__eits__i-instructions tool to see all available commands and workflows
```

### Available MCP Tools

The server provides tools prefixed with `mcp__eits__i-*`:

- `i-instructions` - Complete workflow and initialization guide
- `i-start-session` - Register a new agent session
- `i-end-session` - Complete an agent session
- `i-save-session-context` - Save session context (markdown) with history
- `i-save-agent-context` - Save agent‑specific context (markdown) per project
- `i-speak` - Text-to-speech notifications (macOS)
- `i-todo-*` - Task management tools
- `i-note-add` - Add notes to sessions
- `i-commits` - Track git commits
- And more...

## Database Location

The MCP server uses:
```
~/.config/eye-in-the-sky/eits.db
```

This is created automatically on first run. No manual setup required.

## Running Modes

The Eye in the Sky binary has two modes:

### 1. MCP Stdio Server Mode (for Claude Code)
```bash
# Run without arguments
./bin/eye-in-the-sky
```
Starts the MCP server and listens for JSON-RPC messages on stdin/stdout.

### 2. CLI Mode (for direct commands)
```bash
# Run with specific commands
./bin/eye-in-the-sky i-speak "Hello world"
./bin/eye-in-the-sky i-start-session -description "My session"
./bin/eye-in-the-sky -db /custom/path/eits.db i-speak "Using custom DB"
```
The `-db` flag only works in CLI mode.

## Removing the Server

If you need to remove the MCP server:

```bash
claude mcp remove eits -s local
```

## Additional Resources

- **Full Documentation:** See [MANUAL.md](MANUAL.md)
- **Quick Start:** See [QUICKSTART.md](QUICKSTART.md)
- **CLI Mode:** Use `./bin/eye-in-the-sky <command>` for direct CLI access
- **Web Dashboard:** Run `./bin/eye-ui` for the web interface at http://localhost:8080
