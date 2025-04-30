-- +migrate Down
DROP TRIGGER IF EXISTS update_group_events_updated_at;
DROP TABLE IF EXISTS group_events;
