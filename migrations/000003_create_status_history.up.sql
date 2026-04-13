-- 000003_create_status_history.up.sql

CREATE TABLE status_history (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  application_id  UUID        NOT NULL REFERENCES applications (id) ON DELETE CASCADE,

  -- NULL from_status means this is the very first entry (initial "applied" record)
  from_status     VARCHAR(64),
  to_status       VARCHAR(64) NOT NULL,

  -- Optional note explaining the transition (e.g. "Got a call from recruiter")
  note            TEXT,

  changed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT status_history_to_status_valid CHECK (
    to_status IN ('applied', 'phone_screen', 'interview', 'technical_assessment',
                  'offer', 'negotiating', 'accepted', 'rejected', 'withdrawn', 'ghosted')
  )
);

-- Timeline view for a single application (ORDER BY changed_at ASC)
CREATE INDEX idx_status_history_application ON status_history (application_id, changed_at);


-- Automatically update applications.last_activity_at whenever a new
-- status history row is inserted.
CREATE OR REPLACE FUNCTION sync_application_last_activity()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE applications
  SET
    status           = NEW.to_status,
    last_activity_at = NEW.changed_at,
    updated_at       = NOW()
  WHERE id = NEW.application_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_status_history_sync_application
  AFTER INSERT ON status_history
  FOR EACH ROW EXECUTE FUNCTION sync_application_last_activity();
