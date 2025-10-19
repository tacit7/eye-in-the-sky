-- Refactor compactions table to use single session_id instead of old/new pattern
-- A compaction is just a backup of the conversation at that point in time

-- This migration has already been applied manually.
-- The table now has: id, agent_id, session_id (instead of old_session_id/new_session_id)
-- This file is kept as documentation of the schema change.

-- Migration was:
-- 1. Created compactions_new with session_id column
-- 2. Migrated data: new_session_id -> session_id
-- 3. Dropped old compactions table
-- 4. Renamed compactions_new to compactions
-- 5. Recreated indexes

-- No-op: Migration already applied
SELECT 'Migration 0020: Already applied - no action needed' AS status;
