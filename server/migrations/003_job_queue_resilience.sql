-- Migration 003: Job queue resilience features
-- Run: psql -d cms_ai -f server/migrations/003_job_queue_resilience.sql

-- Add new columns to jobs table for retry and deduplication support
ALTER TABLE jobs 
ADD COLUMN retry_count INT DEFAULT 0,
ADD COLUMN max_retries INT DEFAULT 3,
ADD COLUMN last_retry_at TIMESTAMPTZ,
ADD COLUMN deduplication_id TEXT,
ADD COLUMN metadata JSONB;

-- Update status check constraint to include new statuses
ALTER TABLE jobs DROP CONSTRAINT jobs_status_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_status_check 
CHECK (status IN ('Queued', 'Running', 'Done', 'Failed', 'Retry', 'DeadLetter'));

-- Add indexes for new queue features
CREATE INDEX idx_jobs_deduplication ON jobs(org_id, deduplication_id) WHERE deduplication_id IS NOT NULL;
CREATE INDEX idx_jobs_retry_status ON jobs(status) WHERE status IN ('Retry', 'DeadLetter');
CREATE INDEX idx_jobs_next_retry ON jobs(status, last_retry_at) WHERE status = 'Retry';

-- Add comments for documentation
COMMENT ON COLUMN jobs.retry_count IS 'Number of times this job has been retried';
COMMENT ON COLUMN jobs.max_retries IS 'Maximum number of retry attempts allowed';
COMMENT ON COLUMN jobs.last_retry_at IS 'Timestamp of the last retry attempt';
COMMENT ON COLUMN jobs.deduplication_id IS 'ID used to prevent duplicate jobs';
COMMENT ON COLUMN jobs.metadata IS 'Additional metadata for job processing';