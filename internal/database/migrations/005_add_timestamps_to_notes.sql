-- Migration 005: Add Ecto-compatible timestamp columns to notes table
-- This allows Phoenix/Elixir to use standard timestamps() macro

-- Add inserted_at and updated_at columns
ALTER TABLE notes ADD COLUMN inserted_at TIMESTAMP;
ALTER TABLE notes ADD COLUMN updated_at TIMESTAMP;

-- Populate existing rows with created_at values
UPDATE notes SET inserted_at = created_at WHERE inserted_at IS NULL;
UPDATE notes SET updated_at = created_at WHERE updated_at IS NULL;

-- Note: SQLite doesn't support adding columns with CURRENT_TIMESTAMP default
-- New rows will need to have these fields set explicitly by the application
