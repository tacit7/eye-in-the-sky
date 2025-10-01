-- Migration: Create personas table for expert agent templates
-- This stores predefined agent personas with their expertise and initial context

CREATE TABLE personas (
    id TEXT PRIMARY KEY,                    -- persona identifier (e.g., "frontend-specialist", "security-expert")
    name TEXT NOT NULL,                     -- human-readable name
    description TEXT NOT NULL,              -- brief description of the persona
    expertise TEXT NOT NULL,                -- JSON array of expertise areas
    initial_context TEXT NOT NULL,          -- the persona's initial instructions/context
    preferred_tools TEXT,                   -- JSON array of preferred MCP tools
    specialization TEXT,                    -- primary domain (e.g., "frontend", "backend", "security")
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups by specialization
CREATE INDEX idx_personas_specialization ON personas(specialization);

-- Allow agents to reference a persona
ALTER TABLE agents ADD COLUMN persona_id TEXT REFERENCES personas(id);
