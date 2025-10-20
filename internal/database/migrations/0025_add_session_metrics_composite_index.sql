-- Add composite index for session_id and timestamp queries
CREATE INDEX IF NOT EXISTS idx_session_metrics_session_ts
ON session_metrics(session_id, timestamp);
