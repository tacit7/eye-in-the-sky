-- Remove FTS5 triggers and handle sync in application code
-- This avoids "unsafe use of virtual table" errors with modernc.org/sqlite
DROP TRIGGER IF EXISTS sync_task_search_insert;
DROP TRIGGER IF EXISTS sync_task_search_update;
DROP TRIGGER IF EXISTS sync_task_search_delete;
