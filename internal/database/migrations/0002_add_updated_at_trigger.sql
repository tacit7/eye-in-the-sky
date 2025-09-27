-- Migration 0002: Add automatic updated_at trigger
-- Automatically update updated_at timestamp when agents table is modified

-- Trigger for agents table to automatically update updated_at on UPDATE
CREATE TRIGGER IF NOT EXISTS update_agents_updated_at
    AFTER UPDATE ON agents
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at  -- Only update if updated_at wasn't explicitly set
BEGIN
    UPDATE agents SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Record this migration
INSERT OR IGNORE INTO schema_migrations (version) VALUES ('0002_add_updated_at_trigger');