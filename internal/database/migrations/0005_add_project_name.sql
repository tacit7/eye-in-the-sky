-- Migration 0005: Add project_name to agents table
-- Add project tracking to better identify what project each agent is working on

ALTER TABLE agents ADD COLUMN project_name TEXT;

-- Record this migration
INSERT OR IGNORE INTO schema_migrations (version) VALUES ('0005_add_project_name');