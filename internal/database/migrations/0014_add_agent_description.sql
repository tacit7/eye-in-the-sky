-- Add description column to agents table for agent name/label
ALTER TABLE agents ADD COLUMN description TEXT;

-- Create index for description searches
CREATE INDEX idx_agents_description ON agents(description);
