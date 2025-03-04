CREATE TABLE
    IF NOT EXISTS post_allowed_users (
        post_id TEXT,
        user_id TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (post_id, user_id),
        FOREIGN KEY (post_id) REFERENCES posts (id),
        FOREIGN KEY (user_id) REFERENCES users (id)
    );

CREATE INDEX idx_post_allowed_users ON post_allowed_users (post_id, user_id);