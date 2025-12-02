-- Migration: Add subagent_prompts table for reusable prompt templates
-- Created: 2025-12-02
-- Description: Stores global and project-scoped prompt templates for spawning subagents

CREATE TABLE IF NOT EXISTS subagent_prompts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT,
    prompt_text TEXT NOT NULL,
    project_id TEXT,
    active BOOLEAN DEFAULT 1,
    version INTEGER DEFAULT 1,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CHECK (slug GLOB '[a-z][a-z0-9-]*')
);

-- Partial unique indexes: Allow same slug for global + per-project
CREATE UNIQUE INDEX idx_subagent_prompts_slug_global
  ON subagent_prompts(slug) WHERE project_id IS NULL;

CREATE UNIQUE INDEX idx_subagent_prompts_slug_project
  ON subagent_prompts(slug, project_id) WHERE project_id IS NOT NULL;

-- Performance indexes
CREATE INDEX idx_subagent_prompts_project_id ON subagent_prompts(project_id);
CREATE INDEX idx_subagent_prompts_active ON subagent_prompts(active);

-- BEFORE UPDATE trigger to auto-update timestamps and version
CREATE TRIGGER update_subagent_prompts_timestamp
BEFORE UPDATE ON subagent_prompts
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    SELECT CASE
        WHEN NEW.prompt_text != OLD.prompt_text THEN
            (SELECT NEW.version + 1)
        ELSE
            OLD.version
    END;
    SELECT CURRENT_TIMESTAMP;
END;
