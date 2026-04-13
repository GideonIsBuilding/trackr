-- 000002_create_applications.up.sql

-- Valid application statuses.
-- Stored as VARCHAR (not ENUM) so new statuses never require a migration.
-- The CHECK constraint still enforces the allowed set at the DB level.
CREATE TABLE applications (
  id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,

  -- Role details
  company           VARCHAR(255) NOT NULL,
  role              VARCHAR(255) NOT NULL,
  job_url           TEXT,
  location          VARCHAR(255),

  -- Where did you find this role?
  source            VARCHAR(64)  NOT NULL DEFAULT 'other',

  -- Current status (denormalised for fast reads; history lives in status_history)
  status            VARCHAR(64)  NOT NULL DEFAULT 'applied',

  -- Dates
  applied_at        DATE         NOT NULL DEFAULT CURRENT_DATE,
  -- Bumped on every status change or note edit — used by reminder engine
  last_activity_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  notes             TEXT,

  created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT applications_source_valid CHECK (
    source IN ('linkedin', 'referral', 'company_site', 'job_board', 'recruiter', 'other')
  ),
  CONSTRAINT applications_status_valid CHECK (
    status IN ('applied', 'phone_screen', 'interview', 'technical_assessment',
               'offer', 'negotiating', 'accepted', 'rejected', 'withdrawn', 'ghosted')
  )
);

CREATE TRIGGER trg_applications_updated_at
  BEFORE UPDATE ON applications
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- All applications for a given user (dashboard query)
CREATE INDEX idx_applications_user_id ON applications (user_id);

-- Reminder engine: find stale apps quickly
CREATE INDEX idx_applications_last_activity ON applications (user_id, last_activity_at);

-- Filter by status
CREATE INDEX idx_applications_status ON applications (user_id, status);
