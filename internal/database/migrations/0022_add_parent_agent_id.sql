-- Add parent_agent_id column to agents table for tracking agent hierarchies
ALTER TABLE agents ADD COLUMN parent_agent_id TEXT DEFAULT NULL;

-- Add index for efficient parent lookups
CREATE INDEX IF NOT EXISTS idx_agents_parent_agent_id ON agents(parent_agent_id);

-- Add comment explaining the column
-- parent_agent_id: References the agent ID that spawned this agent (for subagents)
-- NULL indicates a top-level/root agent with no parent
