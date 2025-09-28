-- Add window_id column for Claude Desktop agents
ALTER TABLE agents ADD COLUMN window_id TEXT;

-- Create index for window_id lookups
CREATE INDEX IF NOT EXISTS idx_agents_window_id ON agents(window_id);