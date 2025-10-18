-- Migration: Add current_session_id to agents table
-- This allows us to track which session is currently active for each agent

-- Add current_session_id column to agents table
ALTER TABLE agents ADD COLUMN current_session_id TEXT;

-- Add index for faster lookups
CREATE INDEX idx_agents_current_session ON agents(current_session_id);

-- Add foreign key constraint
-- Note: SQLite doesn't support adding foreign keys to existing tables,
-- so we'll enforce this in the application layer
