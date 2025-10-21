-- Skip migration if session_id already exists (column may have been manually renamed)
-- This migration is a no-op since the column is already named session_id

-- Just ensure the index exists with the new name
DROP INDEX IF EXISTS idx_agents_current_session;
CREATE INDEX IF NOT EXISTS idx_agents_session_id ON agents(session_id);
