-- Add parent_session_id column to agents table
ALTER TABLE agents ADD COLUMN parent_session_id TEXT;

-- Create index for parent_session_id lookups
CREATE INDEX IF NOT EXISTS idx_agents_parent_session_id ON agents(parent_session_id);
