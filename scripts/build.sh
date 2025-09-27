#!/bin/bash

# Build script for Eye in the Sky MCP Server

set -e

echo "🔨 Building Eye in the Sky MCP Server..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build the MCP server
echo "📦 Building MCP server binary..."
go build -o bin/eye-in-the-sky ./cmd/server

echo "✅ Build complete!"
echo "🚀 Run with: ./bin/eye-in-the-sky"
echo "📖 Help: ./bin/eye-in-the-sky -help"