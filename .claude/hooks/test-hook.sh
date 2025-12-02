#!/usr/bin/env bash
# Simple test hook - just write to a file
echo "Hook executed at $(date)" >> /tmp/claude-hook-test.log
echo "Tool: ${tool_name:-unknown}" >> /tmp/claude-hook-test.log
echo "---" >> /tmp/claude-hook-test.log
