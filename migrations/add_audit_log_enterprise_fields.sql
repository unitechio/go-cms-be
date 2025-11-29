-- Add new columns to audit_logs table for enterprise-level audit trail
-- Migration: Add request_body, response_body, and finished_at columns

-- Add request_body column (CLOB/TEXT)
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS request_body TEXT;

-- Add response_body column (CLOB/TEXT)
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS response_body TEXT;

-- Add finished_at column with index
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS finished_at TIMESTAMP;

-- Create index on finished_at for performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_finished_at ON audit_logs(finished_at DESC);

-- Add comments for documentation
COMMENT ON COLUMN audit_logs.request_body IS 'Full HTTP request body (sanitized, sensitive data removed)';
COMMENT ON COLUMN audit_logs.response_body IS 'Full HTTP response body (limited to 10KB)';
COMMENT ON COLUMN audit_logs.finished_at IS 'Timestamp when the request finished processing';
COMMENT ON COLUMN audit_logs.created_at IS 'Timestamp when the request started processing';
COMMENT ON COLUMN audit_logs.duration IS 'Request processing duration in milliseconds';
