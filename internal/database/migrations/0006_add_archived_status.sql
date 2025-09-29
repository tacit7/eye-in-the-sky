-- Migration 0006: Add archived status support
-- Allow agents to be archived to keep them out of the main dashboard view

-- Note: We don't need to add a new column since we can use the existing status field
-- Just documenting that 'archived' is now a valid status value alongside:
-- 'active', 'idle', 'working', 'completed', 'failed'

-- Record this migration
INSERT OR IGNORE INTO schema_migrations (version) VALUES ('0006_add_archived_status');