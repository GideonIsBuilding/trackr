-- 000005_create_reminders.up.sql

CREATE TABLE reminders (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  application_id      UUID        NOT NULL REFERENCES applications (id) ON DELETE CASCADE,

  -- How many days of silence before an alert fires
  trigger_after_days  INT         NOT NULL DEFAULT 14,

  -- Whether this reminder rule is active at all
  is_active           BOOLEAN     NOT NULL DEFAULT TRUE,

  -- When the last notification was actually sent to the user
  last_sent_at        TIMESTAMPTZ,

  -- Snooze: if set and in the future, the engine skips this reminder until then
  snoozed_until       TIMESTAMPTZ,

  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- One reminder rule per application
  CONSTRAINT reminders_application_unique UNIQUE (application_id),

  CONSTRAINT reminders_trigger_days_positive CHECK (trigger_after_days > 0)
);

-- Reminder engine query: active, un-snoozed reminders
CREATE INDEX idx_reminders_active ON reminders (application_id)
  WHERE is_active = TRUE;
