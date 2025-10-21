-- Rename current_session_id to session_id in agents table
ALTER TABLE agents RENAME COLUMN current_session_id TO session_id;

-- Update indexes
DROP INDEX IF EXISTS idx_agents_current_session;
CREATE INDEX IF NOT EXISTS idx_agents_session_id ON agents(session_id);
