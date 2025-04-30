-- +migrate Up
CREATE TABLE IF NOT EXISTS group_events (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    creator_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    event_time DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Trigger to update updated_at timestamp on update
CREATE TRIGGER IF NOT EXISTS update_group_events_updated_at
AFTER UPDATE ON group_events
FOR EACH ROW
BEGIN
    UPDATE group_events SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
