-- Create the table to store user responses to group events
CREATE TABLE IF NOT EXISTS group_event_responses (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    response TEXT NOT NULL CHECK(response IN ('going', 'not_going')), -- Ensure valid response values
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (event_id) REFERENCES group_events(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    -- Ensure only one response per user per event
    UNIQUE (event_id, user_id)
);

-- Trigger to automatically update 'updated_at' timestamp
CREATE TRIGGER IF NOT EXISTS update_group_event_responses_updated_at
AFTER UPDATE ON group_event_responses
FOR EACH ROW
BEGIN
    UPDATE group_event_responses SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
