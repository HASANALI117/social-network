-- +migrate Up
CREATE TABLE
    IF NOT EXISTS followers (
        follower_id TEXT NOT NULL,
        following_id TEXT NOT NULL,
        status TEXT NOT NULL CHECK (status IN ('pending', 'accepted')),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (follower_id, following_id),
        FOREIGN KEY (follower_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (following_id) REFERENCES users (id) ON DELETE CASCADE
    );

CREATE INDEX idx_followers_follower_id ON followers (follower_id);

CREATE INDEX idx_followers_following_id ON followers (following_id);
