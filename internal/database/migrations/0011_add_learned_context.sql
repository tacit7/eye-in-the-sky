-- Migration: Add learned_context to session_context table
-- This stores the agent's initial expertise/context at session start

ALTER TABLE session_context ADD COLUMN learned_context TEXT;
