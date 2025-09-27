#!/bin/bash

set -e

echo "🔨 Building Eye in the Sky MCP Server..."

mkdir -p bin

echo "📦 Building MCP server binary..."
go build -o bin/eye-in-the-sky ./cmd/server

echo "✅ Build complete!"
echo "🚀 Run with: ./bin/eye-in-the-sky"
echo "📖 Help: ./bin/eye-in-the-sky -help"