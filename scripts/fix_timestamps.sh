#!/bin/bash

# Fix bad timestamps in the database
# Converts Go-formatted timestamps to SQLite/Ecto-compatible format
# From: "2025-12-03 08:37:56.916653 -0600 CST m=+..."
# To:   "2025-12-03 08:37:56.916653"

DB_PATH="$HOME/.config/eye-in-the-sky/eits.db"

echo "Fixing bad timestamps in $DB_PATH..."

# Fix tasks.created_at - extract just the datetime part (first 26 chars)
sqlite3 "$DB_PATH" "
UPDATE tasks
SET created_at = substr(created_at, 1, 26)
WHERE created_at LIKE '%-%-%:%:%' AND length(created_at) > 26;
"

# Fix tasks.updated_at
sqlite3 "$DB_PATH" "
UPDATE tasks
SET updated_at = substr(updated_at, 1, 26)
WHERE updated_at LIKE '%-%-%:%:%' AND length(updated_at) > 26;
"

# Fix task_sessions.created_at
sqlite3 "$DB_PATH" "
UPDATE task_sessions
SET created_at = substr(created_at, 1, 26)
WHERE created_at LIKE '%-%-%:%:%' AND length(created_at) > 26;
"

# Fix task_notes.created_at
sqlite3 "$DB_PATH" "
UPDATE task_notes
SET created_at = substr(created_at, 1, 26)
WHERE created_at LIKE '%-%-%:%:%' AND length(created_at) > 26;
"

echo "Done! Timestamps fixed."
echo "Format: YYYY-MM-DD HH:MM:SS.microseconds"
