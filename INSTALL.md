# 🔧 Installation Guide

## Quick Install (Recommended)

**One-line install for macOS/Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/eye-in-the-sky/main/install.sh | bash
```

## Manual Installation

### 1. Prerequisites
- Go 1.21 or later
- Git
- Claude Code CLI or Claude Desktop

### 2. Clone and Build
```bash
git clone https://github.com/yourusername/eye-in-the-sky.git
cd eye-in-the-sky
chmod +x install.sh
./install.sh
```

### 3. Manual Setup (Alternative)

If you prefer manual setup:

```bash
# Build the applications
go build -o ~/.local/bin/eye-in-the-sky ./cmd/server/main.go
go build -o ~/.local/bin/eye-in-the-sky-integrated ./main.go

# Create data directory
mkdir -p ~/.eye-in-the-sky/data

# Add to PATH (add to ~/.zshrc or ~/.bashrc)
export PATH="$HOME/.local/bin:$PATH"
```

### 4. Configure Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "eye-in-the-sky": {
      "command": "/Users/YOUR_USERNAME/.local/bin/eye-in-the-sky",
      "args": [
        "-db",
        "/Users/YOUR_USERNAME/.eye-in-the-sky/data/agents.db"
      ],
      "env": {
        "PATH": "/usr/local/bin:/usr/bin:/bin",
        "HOME": "/Users/YOUR_USERNAME"
      }
    }
  }
}
```

Replace `YOUR_USERNAME` with your actual username.

### 5. Start the Dashboard

```bash
eye-in-the-sky-start
```

Visit: http://localhost:8080

## Verification

1. **Check MCP server** (for Claude Code):
   ```bash
   claude mcp list
   ```
   Should show `eye-in-the-sky` in the list.

2. **Test dashboard:**
   ```bash
   curl -I http://localhost:8080
   ```
   Should return `200 OK`.

3. **Register test agent:**
   ```bash
   # In Claude Code or Claude Desktop
   Register as Claude Code agent working on test project
   ```

## What the Install Script Does

1. ✅ Checks Go installation
2. ✅ Creates `~/.eye-in-the-sky/` directory
3. ✅ Builds MCP and integrated servers
4. ✅ Installs binaries to `~/.local/bin/`
5. ✅ Configures Claude Desktop MCP integration
6. ✅ Creates start script and desktop shortcut
7. ✅ Updates your shell PATH
8. ✅ Tests the installation

## Troubleshooting

### MCP Server Not Found
- Restart Claude Desktop after installation
- Check `claude_desktop_config.json` configuration
- Verify binary exists: `ls ~/.local/bin/eye-in-the-sky`

### Port Already in Use
```bash
# Use different port
eye-in-the-sky-start 8081
```

### Permission Denied
```bash
# Fix permissions
chmod +x ~/.local/bin/eye-in-the-sky*
```

### PATH Issues
```bash
# Reload shell configuration
source ~/.zshrc  # or ~/.bashrc
```

## Uninstall

```bash
# Remove binaries
rm ~/.local/bin/eye-in-the-sky*

# Remove installation directory
rm -rf ~/.eye-in-the-sky

# Remove from PATH (edit ~/.zshrc or ~/.bashrc)
# Remove this line: export PATH="$HOME/.local/bin:$PATH"

# Remove Claude Desktop config (optional)
# Edit: ~/Library/Application Support/Claude/claude_desktop_config.json
```

## Support

- 📖 Full documentation: [README.md](README.md)
- 🚀 Quick start: [QUICKSTART.md](QUICKSTART.md)
- 🛠️ Technical details: [CLAUDE.md](CLAUDE.md)
- 🐛 Issues: [GitHub Issues](https://github.com/yourusername/eye-in-the-sky/issues)