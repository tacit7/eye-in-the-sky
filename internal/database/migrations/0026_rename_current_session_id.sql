-- Rename current_session_id to session_id
ALTER TABLE agents RENAME COLUMN current_session_id TO session_id;

-- Recreate index with new name
DROP INDEX IF EXISTS idx_agents_current_session;
CREATE INDEX idx_agents_session_id ON agents(session_id);
