-- Remove featured column and its index
DROP INDEX IF EXISTS idx_posts_featured;
ALTER TABLE posts DROP COLUMN IF EXISTS featured;