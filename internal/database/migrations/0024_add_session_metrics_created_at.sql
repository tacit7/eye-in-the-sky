-- Add created_at column to session_metrics table for better timestamp tracking
ALTER TABLE session_metrics ADD COLUMN created_at DATETIME DEFAULT CURRENT_TIMESTAMP;
