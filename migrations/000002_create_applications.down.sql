-- 000002_create_applications.down.sql
DROP TRIGGER IF EXISTS trg_applications_updated_at ON applications;
DROP TABLE IF EXISTS applications;
