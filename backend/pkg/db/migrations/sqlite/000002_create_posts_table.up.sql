CREATE TABLE IF NOT EXISTS posts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT,
    content TEXT,
    image_url TEXT,
    privacy TEXT CHECK (privacy IN ('public', 'almost_private', 'private')) NOT NULL DEFAULT 'public', -- Fixed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_posts_user_id ON posts (user_id);