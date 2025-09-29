#!/bin/bash

echo "=== Eye in the Sky MCP Diagnostics ==="
echo "Date: $(date)"
echo

# Check if binary exists and is executable
echo "1. Checking binary..."
BINARY_PATH="/Users/urielmaldonado/projects/eye-in-the-sky/bin/eye-in-the-sky"
if [ -f "$BINARY_PATH" ]; then
    echo "✅ Binary exists: $BINARY_PATH"
    echo "   Permissions: $(ls -la "$BINARY_PATH")"
    echo "   Size: $(du -h "$BINARY_PATH" | cut -f1)"
else
    echo "❌ Binary not found: $BINARY_PATH"
    exit 1
fi

# Check if database directory exists
echo
echo "2. Checking database..."
DB_PATH="/Users/urielmaldonado/projects/eye-in-the-sky/data/agents.db"
DB_DIR=$(dirname "$DB_PATH")
if [ -d "$DB_DIR" ]; then
    echo "✅ Database directory exists: $DB_DIR"
    if [ -f "$DB_PATH" ]; then
        echo "✅ Database file exists: $DB_PATH"
        echo "   Size: $(du -h "$DB_PATH" | cut -f1)"
    else
        echo "ℹ️  Database file will be created: $DB_PATH"
    fi
else
    echo "❌ Database directory not found: $DB_DIR"
    mkdir -p "$DB_DIR"
    echo "✅ Created database directory: $DB_DIR"
fi

# Check Claude Desktop config
echo
echo "3. Checking Claude Desktop configuration..."
CONFIG_PATH="/Users/urielmaldonado/Library/Application Support/Claude/claude_desktop_config.json"
if [ -f "$CONFIG_PATH" ]; then
    echo "✅ Config file exists: $CONFIG_PATH"
    echo "   Content:"
    cat "$CONFIG_PATH" | jq . 2>/dev/null || cat "$CONFIG_PATH"
else
    echo "❌ Config file not found: $CONFIG_PATH"
fi

# Test basic MCP functionality
echo
echo "4. Testing basic MCP functionality..."
echo "Testing MCP initialize..."

# Create a temporary test script
INIT_JSON='{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {"roots": {"listChanged": true}}, "clientInfo": {"name": "claude-desktop", "version": "0.7.1"}}}'

echo "Sending: $INIT_JSON"
echo

# Test the MCP server
RESPONSE=$(echo "$INIT_JSON" | timeout 5s "$BINARY_PATH" -db "$DB_PATH" 2>/dev/null)
if [ $? -eq 0 ] && [ -n "$RESPONSE" ]; then
    echo "✅ MCP server responds correctly"
    echo "Response: $RESPONSE"
else
    echo "❌ MCP server failed to respond"
    echo "Trying with stderr output..."
    echo "$INIT_JSON" | timeout 5s "$BINARY_PATH" -db "$DB_PATH"
fi

# Test tools list
echo
echo "5. Testing tools list..."
TOOLS_JSON='{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}'
echo "Testing tools/list after initialization..."

# Full MCP sequence
(echo "$INIT_JSON"; sleep 0.1; echo "$TOOLS_JSON"; sleep 0.1) | timeout 10s "$BINARY_PATH" -db "$DB_PATH" 2>/dev/null | tail -1 | jq '.result.tools[] | select(.name == "register_claude_desktop_agent") | .name' 2>/dev/null

if [ $? -eq 0 ]; then
    echo "✅ register_claude_desktop_agent tool found"
else
    echo "❌ register_claude_desktop_agent tool not found or parsing failed"
fi

# Check for common issues
echo
echo "6. Checking for common issues..."

# Check if Claude Desktop is running
if pgrep -f "Claude" > /dev/null; then
    echo "✅ Claude Desktop appears to be running"
else
    echo "⚠️  Claude Desktop does not appear to be running"
fi

# Check permissions
if [ -r "$CONFIG_PATH" ]; then
    echo "✅ Config file is readable"
else
    echo "❌ Config file is not readable"
fi

if [ -x "$BINARY_PATH" ]; then
    echo "✅ Binary is executable"
else
    echo "❌ Binary is not executable"
fi

echo
echo "=== Diagnostics Complete ==="
echo
echo "If you're still seeing errors, please:"
echo "1. Restart Claude Desktop completely"
echo "2. Try running: /mcp"
echo "3. Look for 'eye-in-the-sky' in the server list"
echo "4. If you see it, try: /mcp help eye-in-the-sky"
echo "5. Share the exact error message you see"