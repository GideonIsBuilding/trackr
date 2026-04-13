-- 000007_add_application_checklist.up.sql
-- Adds submission checklist fields to the applications table.
-- All boolean fields default to FALSE (not done).

ALTER TABLE applications
  ADD COLUMN cover_letter      BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN cv_tailored       BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN referral          BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN portfolio_link    TEXT,
  ADD COLUMN video_intro       BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN linkedin_connect  BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN applications.cover_letter     IS 'Was a cover letter included?';
COMMENT ON COLUMN applications.cv_tailored      IS 'Was the CV tailored to this JD?';
COMMENT ON COLUMN applications.referral         IS 'Was there an internal referral?';
COMMENT ON COLUMN applications.portfolio_link   IS 'Link to portfolio/project submitted';
COMMENT ON COLUMN applications.video_intro      IS 'Was a video introduction included?';
COMMENT ON COLUMN applications.linkedin_connect IS 'Did you connect with the hiring manager on LinkedIn?';
