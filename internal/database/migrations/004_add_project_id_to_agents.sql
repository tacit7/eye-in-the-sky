-- Migration: Add project_id foreign key to agents table
-- This allows proper relational linking between agents and projects

-- Add project_id column to agents
ALTER TABLE agents ADD COLUMN project_id TEXT REFERENCES projects(id);

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_agents_project_id ON agents(project_id);

-- Backfill project_id from project_name for existing agents
-- This will link agents to projects based on the project name match
UPDATE agents
SET project_id = (
    SELECT id FROM projects
    WHERE name = agents.project_name
    LIMIT 1
)
WHERE project_name IS NOT NULL
AND project_id IS NULL;

-- Note: We keep project_name column for backward compatibility
-- It can be dropped in a future migration if needed
