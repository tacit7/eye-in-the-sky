#!/usr/bin/env bash
# Debug hook to see what data is actually passed

LOG_FILE="/tmp/claude-hook-debug.log"

echo "=== Hook execution at $(date) ===" >> "$LOG_FILE"
echo "Working directory: $(pwd)" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

echo "Environment variables:" >> "$LOG_FILE"
env | sort >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

echo "STDIN content:" >> "$LOG_FILE"
cat >> "$LOG_FILE"
echo "" >> "$LOG_FILE"
echo "=== End ===" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"
