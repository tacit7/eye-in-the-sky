-- Migration 003: Consolidate CCUsage tables into eits.db
-- This migration adds the usage_entries and file_metadata tables to the main database

-- Create usage_entries table
CREATE TABLE IF NOT EXISTS usage_entries (
    id INTEGER PRIMARY KEY,
    session_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    project TEXT NOT NULL,
    model TEXT NOT NULL,
    input_tokens INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,
    cache_creation_tokens INTEGER NOT NULL,
    cache_read_tokens INTEGER NOT NULL,
    total_cost REAL NOT NULL,
    message_id TEXT NOT NULL,
    request_id TEXT NOT NULL,
    unique_hash TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create file_metadata table
CREATE TABLE IF NOT EXISTS file_metadata (
    file_path TEXT PRIMARY KEY,
    last_mtime INTEGER NOT NULL,
    last_parsed_at TEXT NOT NULL
);

-- Create indexes for usage_entries
CREATE INDEX IF NOT EXISTS idx_usage_timestamp ON usage_entries(timestamp);
CREATE INDEX IF NOT EXISTS idx_usage_project ON usage_entries(project);
CREATE INDEX IF NOT EXISTS idx_usage_model ON usage_entries(model);
CREATE INDEX IF NOT EXISTS idx_usage_session_id ON usage_entries(session_id);
