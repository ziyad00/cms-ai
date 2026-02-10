-- Migration: Add progress tracking to jobs
-- Run: psql -d cms_ai -f server/migrations/005_job_progress.sql

ALTER TABLE jobs ADD COLUMN IF NOT EXISTS progress_step TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS progress_pct INTEGER DEFAULT 0;
