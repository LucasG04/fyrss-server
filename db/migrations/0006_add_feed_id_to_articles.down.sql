-- Remove feed_id relationship from articles table
DROP INDEX IF EXISTS idx_articles_feed_id;
ALTER TABLE articles DROP CONSTRAINT IF EXISTS fk_articles_feed_id;
ALTER TABLE articles DROP COLUMN IF EXISTS feed_id;