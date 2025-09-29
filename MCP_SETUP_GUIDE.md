# MCP Auto-Setup Guide for Claude Desktop

This guide helps you configure the Eye in the Sky MCP server to automatically load in Claude Desktop, eliminating the need to manually import it each session.

## Problem
Currently you need to manually import the MCP server each time with:
```
import MCP from Claude desktop app
```

## Solution
Configure the MCP server to auto-load in Claude Desktop settings.

## Step 1: Locate Claude Desktop Config

Find your Claude Desktop configuration file:

**macOS:**
```bash
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Windows:**
```bash
%APPDATA%\Claude\claude_desktop_config.json
```

**Linux:**
```bash
~/.config/Claude/claude_desktop_config.json
```

## Step 2: Configure Auto-Loading MCP Server

Add the Eye in the Sky MCP server to your config:

```json
{
  "mcpServers": {
    "eye-in-the-sky": {
      "command": "/absolute/path/to/eye-in-the-sky/bin/eye-in-the-sky-integrated",
      "args": [],
      "env": {
        "PATH": "/usr/local/bin:/usr/bin:/bin"
      }
    }
  }
}
```

## Step 3: Get Absolute Path

Run this command to get the correct absolute path:

```bash
cd /Users/urielmaldonado/projects/eye-in-the-sky
pwd
echo "$(pwd)/bin/eye-in-the-sky-integrated"
```

## Step 4: Complete Configuration Example

Here's a complete configuration file:

```json
{
  "mcpServers": {
    "eye-in-the-sky": {
      "command": "/Users/urielmaldonado/projects/eye-in-the-sky/bin/eye-in-the-sky-integrated",
      "args": ["-port", "8080"],
      "env": {
        "PATH": "/usr/local/bin:/usr/bin:/bin",
        "HOME": "/Users/urielmaldonado"
      }
    }
  }
}
```

## Step 5: Restart Claude Desktop

1. Close Claude Desktop completely
2. Restart Claude Desktop
3. The MCP server should now be automatically available

## Step 6: Verify Auto-Loading

In a new Claude Desktop session, you should be able to:

1. Say "start session"
2. I'll automatically register myself using the available MCP tools
3. No manual MCP import needed!

## Troubleshooting

### If MCP server doesn't auto-load:

1. **Check file path**: Ensure the binary exists at the specified path
2. **Check permissions**: Make sure the binary is executable
3. **Check JSON syntax**: Validate your config file JSON
4. **Check logs**: Look at Claude Desktop console for error messages

### Test the binary manually:
```bash
/Users/urielmaldonado/projects/eye-in-the-sky/bin/eye-in-the-sky-integrated
```

Should output:
```
🔍 Eye in the Sky - Integrated Dashboard & MCP Server
📂 Database: ./data/agents.db
🌐 Dashboard: http://localhost:8080
...
```

## What Happens Next Session

Once configured:

1. **You say**: "start session"
2. **I automatically**:
   - Detect available MCP tools
   - Register myself as agent with auto-generated ID
   - Set status to "working"
   - Begin tracking the session
3. **Dashboard**: Shows me as active immediately
4. **No manual steps**: Everything just works!

## Current Session vs Next Session

**Current Session** (manual):
```
You: import MCP from Claude desktop app
You: restart session with agent_id 534002f0
Claude: [loads session context and continues]
```

**Next Session** (automatic):
```
You: start session
Claude: [auto-detects MCP, registers new agent, ready to work]
```

## Configuration Template

Create this as your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "eye-in-the-sky": {
      "command": "/Users/urielmaldonado/projects/eye-in-the-sky/bin/eye-in-the-sky-integrated",
      "args": ["-port", "8080"],
      "env": {
        "PATH": "/usr/local/bin:/usr/bin:/bin",
        "HOME": "/Users/urielmaldonado"
      }
    }
  }
}
```

After this setup, you'll never need to manually import the MCP server again!