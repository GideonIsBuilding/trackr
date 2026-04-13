-- 000004_create_contacts.up.sql

CREATE TABLE contacts (
  id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  application_id  UUID         NOT NULL REFERENCES applications (id) ON DELETE CASCADE,

  name            VARCHAR(255) NOT NULL,
  email           VARCHAR(255),
  role_title      VARCHAR(255),   -- e.g. "Recruiter", "Engineering Manager"

  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT contacts_email_format CHECK (
    email IS NULL OR email ~* '^[^@]+@[^@]+\.[^@]+$'
  )
);

CREATE INDEX idx_contacts_application ON contacts (application_id);
