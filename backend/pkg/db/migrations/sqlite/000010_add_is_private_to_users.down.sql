-- Remove is_private column from users table
-- Note: SQLite does not directly support DROP COLUMN in older versions.
-- A common workaround is to recreate the table without the column.
-- However, for simplicity and assuming a modern SQLite version or context
-- where this is acceptable, we'll use the standard syntax.
-- If this fails, a table recreation strategy would be needed.
ALTER TABLE users DROP COLUMN is_private;
