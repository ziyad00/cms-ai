-- Migration 006: Comprehensive Job Table Sync
-- This ensures the jobs table has all columns and constraints required by the latest Go models.

-- 1. Add missing columns if they don't exist
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS retry_count INT DEFAULT 0;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS max_retries INT DEFAULT 3;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS last_retry_at TIMESTAMPTZ;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS deduplication_id TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS progress_step TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS progress_pct INT DEFAULT 0;

-- 2. Update Constraints to support new Types and Statuses
ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_type_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_type_check CHECK (type IN ('render', 'preview', 'export', 'generate', 'bind'));

ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_status_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_status_check CHECK (status IN ('Queued', 'Running', 'Done', 'Failed', 'Retry', 'DeadLetter'));

-- 3. Add index for performance on input_ref (used for listing exports)
CREATE INDEX IF NOT EXISTS idx_jobs_input_ref ON jobs(input_ref);
