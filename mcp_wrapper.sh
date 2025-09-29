#!/bin/bash

# MCP Wrapper Script for Eye in the Sky
# This wrapper helps ensure proper connection with Claude Desktop

# Log the start
echo "🚀 MCP Wrapper starting at $(date)" >> /tmp/eye-in-the-sky-mcp.log

# Change to the project directory
cd "/Users/urielmaldonado/projects/eye-in-the-sky"

# Set up environment
export PATH="/usr/local/bin:/usr/bin:/bin:$PATH"

# Log the command being executed
echo "📝 Executing: ./bin/eye-in-the-sky $@" >> /tmp/eye-in-the-sky-mcp.log

# Execute the MCP server with all passed arguments
exec "./bin/eye-in-the-sky" "$@"