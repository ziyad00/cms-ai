-- Migration 004: Decks + deck versions
-- Run: psql -d cms_ai -f server/migrations/004_decks.sql

CREATE TABLE IF NOT EXISTS decks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  owner_user_id UUID REFERENCES users(id),
  name TEXT NOT NULL,
  source_template_version_id UUID REFERENCES template_versions(id),
  content TEXT NOT NULL,
  current_version_id UUID,
  latest_version_no INT DEFAULT 1,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS deck_versions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  deck_id UUID REFERENCES decks(id) ON DELETE CASCADE,
  org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  version_no INT NOT NULL,
  spec_json JSONB NOT NULL,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ DEFAULT NOW()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_decks_current_version'
  ) THEN
    ALTER TABLE decks
      ADD CONSTRAINT fk_decks_current_version
      FOREIGN KEY (current_version_id) REFERENCES deck_versions(id);
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_decks_org ON decks(org_id);
CREATE INDEX IF NOT EXISTS idx_deck_versions_org_deck ON deck_versions(org_id, deck_id);
