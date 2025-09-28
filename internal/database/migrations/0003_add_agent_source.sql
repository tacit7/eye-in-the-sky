-- Migration 0003: Add source field to agents table
-- Distinguishes between worktree agents and Claude Desktop agents

-- Add source column to agents table
ALTER TABLE agents ADD COLUMN source TEXT NOT NULL DEFAULT 'worktree';

-- Update index to include source for better filtering
CREATE INDEX IF NOT EXISTS idx_agents_source ON agents(source);

-- Record this migration
INSERT OR IGNORE INTO schema_migrations (version) VALUES ('0003_add_agent_source');