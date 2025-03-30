CREATE TABLE IF NOT EXISTS group_members (
    group_id TEXT,
    user_id TEXT,
    role TEXT CHECK (role IN ('admin', 'member')) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups (id),
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE INDEX idx_group_members ON group_members (group_id, user_id);