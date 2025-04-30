-- Add the group_id column to the posts table
ALTER TABLE posts ADD COLUMN group_id TEXT;

-- Add the foreign key constraint
-- Note: SQLite doesn't directly support adding foreign key constraints via ALTER TABLE
-- in a single step if the table already has data or other constraints.
-- A common workaround is to recreate the table, but for simplicity here,
-- we'll just add the column. Proper enforcement would require a more complex migration
-- (create new table, copy data, drop old, rename new) or rely on application logic.
-- For now, we add the column and will add the FK logic if needed later or assume application-level checks.
-- If using PostgreSQL, you would add:
-- ADD CONSTRAINT fk_posts_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;

-- Optional: Add an index for faster lookups of group posts
CREATE INDEX IF NOT EXISTS idx_posts_group_id ON posts (group_id);
