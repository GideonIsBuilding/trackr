-- 000007_add_application_checklist.down.sql
ALTER TABLE applications
  DROP COLUMN IF EXISTS cover_letter,
  DROP COLUMN IF EXISTS cv_tailored,
  DROP COLUMN IF EXISTS referral,
  DROP COLUMN IF EXISTS portfolio_link,
  DROP COLUMN IF EXISTS video_intro,
  DROP COLUMN IF EXISTS linkedin_connect;
