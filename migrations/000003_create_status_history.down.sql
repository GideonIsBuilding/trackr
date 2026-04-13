-- 000003_create_status_history.down.sql
DROP TRIGGER IF EXISTS trg_status_history_sync_application ON status_history;
DROP FUNCTION IF EXISTS sync_application_last_activity();
DROP TABLE IF EXISTS status_history;
