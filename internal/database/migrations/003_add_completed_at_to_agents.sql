-- Migration: Add completed_at column to agents table
-- Purpose: Track when agent sessions are marked as completed
-- Date: 2025-11-09

ALTER TABLE agents ADD COLUMN completed_at DATETIME;
