# 🔧 Installation Guide

## Prerequisites
- Go 1.21 or later
- Git
- Claude Code CLI or Claude Desktop
- NATS server running (for message queue operations)

## Installation Steps

### 1. Clone the Repository
```bash
git clone https://github.com/yourusername/eye-in-the-sky.git
cd eye-in-the-sky
```

### 2. Build Binaries

Create directories and build the applications:
```bash
mkdir -p ~/.eye-in-the-sky/data
mkdir -p ~/.local/bin

cd ~/projects/eye-in-the-sky
go build -o ~/.local/bin/eye-in-the-sky ./cmd/server/main.go
go build -o ~/.local/bin/eye-in-the-sky-integrated ./main.go
```

### 3. Create Start Script

```bash
cat > ~/.local/bin/eye-in-the-sky-start << 'EOF'
#!/bin/bash

INSTALL_DIR="$HOME/.eye-in-the-sky"
PORT=${1:-8080}

echo "🔍 Starting Eye in the Sky Dashboard..."
echo "📊 Dashboard will be available at: http://localhost:$PORT"
echo ""
echo "Press Ctrl+C to stop"

cd "$INSTALL_DIR"
exec "$HOME/.local/bin/eye-in-the-sky-integrated" -port "$PORT"
EOF

chmod +x ~/.local/bin/eye-in-the-sky-start
```

### 4. Configure Claude Code (Per Project)

For each project using eye-in-the-sky, use the Claude Code CLI to register it:

```bash
cd /path/to/your/project
claude mcp add --transport stdio eye-in-the-sky --scope project \
  -- ~/.local/bin/eye-in-the-sky
```

This creates `.mcp.json` in your project root with the proper configuration.

Alternatively, configure it globally (available in all projects):
```bash
claude mcp add --transport stdio eye-in-the-sky --scope user \
  -- ~/.local/bin/eye-in-the-sky
```

### 5. Configure Claude Desktop (Optional)

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "eye-in-the-sky": {
      "command": "/Users/YOUR_USERNAME/.local/bin/eye-in-the-sky",
      "args": [],
      "env": {
        "PATH": "/usr/local/bin:/usr/bin:/bin",
        "HOME": "/Users/YOUR_USERNAME"
      }
    }
  }
}
```

Replace `YOUR_USERNAME` with your actual username, then restart Claude Desktop.

### 6. Start the Dashboard

```bash
eye-in-the-sky-start
```

Visit: http://localhost:8080

## Verification

### Claude Code
```bash
# List all configured MCP servers
claude mcp list
```
Should show `eye-in-the-sky` in the list.

### Dashboard
```bash
# Test if dashboard responds
curl -I http://localhost:8080
```
Should return `200 OK`.

### MCP Server
Once registered, you can use eye-in-the-sky tools:
```bash
# In Claude Code
/mcp  # Lists loaded MCP servers
```

## What Gets Installed

- **eye-in-the-sky** - MCP server binary (~17 MB)
- **eye-in-the-sky-integrated** - Dashboard server binary (~16 MB)
- **eye-in-the-sky-start** - Convenience script to start the dashboard
- **~/.eye-in-the-sky/data** - Data directory for SQLite database

## Troubleshooting

### MCP Server Not Showing in Claude Code
1. Verify the binary exists and is executable:
   ```bash
   ls -lh ~/.local/bin/eye-in-the-sky
   ```

2. Verify the `.mcp.json` file exists in your project:
   ```bash
   cat /path/to/project/.mcp.json
   ```

3. Exit and restart Claude Code for the configuration to load

4. Check if it's registered:
   ```bash
   claude mcp list
   ```
   If you see “No MCP servers configured”, add it:
   ```bash
   claude mcp add --transport stdio eye-in-the-sky -- ~/.local/bin/eye-in-the-sky
   claude mcp list
   ```

### Binary Not Found
```bash
# Check if binaries were built
ls ~/.local/bin/eye-in-the-sky*

# Rebuild if missing
cd ~/projects/eye-in-the-sky
go build -o ~/.local/bin/eye-in-the-sky ./cmd/server/main.go
go build -o ~/.local/bin/eye-in-the-sky-integrated ./main.go
```

### Port Already in Use
```bash
# Use different port
eye-in-the-sky-start 8081
```

### NATS Connection Issues
eye-in-the-sky requires NATS to be running for full functionality:
```bash
# Check if NATS is running
nats-server --version

# Start NATS if not running
nats-server
```

## Uninstall

### Remove MCP Server from Projects

For each project:
```bash
cd /path/to/project
claude mcp remove eye-in-the-sky
```

Or remove globally:
```bash
claude mcp remove eye-in-the-sky --scope user
```

### Remove Binaries and Data

```bash
# Remove binaries
rm ~/.local/bin/eye-in-the-sky*

# Remove data directory
rm -rf ~/.eye-in-the-sky

# Remove Claude Desktop config (optional)
# Edit: ~/Library/Application Support/Claude/claude_desktop_config.json
```

## Documentation

- 📖 [README.md](README.md) - Project overview and features
- 🚀 [QUICKSTART.md](QUICKSTART.md) - Quick start guide
- 🛠️ [CLAUDE.md](CLAUDE.md) - Technical architecture details
- 📊 [MCP_SETUP_GUIDE.md](MCP_SETUP_GUIDE.md) - MCP integration guide
- 🐛 [GitHub Issues](https://github.com/yourusername/eye-in-the-sky/issues) - Report issues
