-- Remove the index first (if it exists)
DROP INDEX IF EXISTS idx_posts_group_id;

-- Remove the group_id column
-- Note: SQLite's support for DROP COLUMN is relatively recent (version 3.35.0+).
-- If using an older version, this might fail, requiring the table recreation workaround.
ALTER TABLE posts DROP COLUMN group_id;

-- If using PostgreSQL, you would drop the constraint first:
-- ALTER TABLE posts DROP CONSTRAINT fk_posts_group;
