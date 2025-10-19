#!/bin/bash

set -e

echo "🔨 Building Eye in the Sky..."

mkdir -p bin

echo "📦 Building MCP server binary..."
go build -o bin/eye-in-the-sky ./cmd/server

echo "📦 Building TUI dashboard binary..."
go build -o bin/dashboard ./cmd/dashboard

echo "✅ Build complete!"
echo ""
echo "Binaries available:"
echo "  🔧 MCP Server: ./bin/eye-in-the-sky"
echo "  📊 Dashboard:  ./bin/dashboard"
echo ""
echo "Usage:"
echo "  MCP Server: ./bin/eye-in-the-sky -db /path/to/agents.db"
echo "  Dashboard:  ./bin/dashboard"