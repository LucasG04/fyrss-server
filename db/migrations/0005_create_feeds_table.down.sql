-- Remove feeds table
DROP INDEX IF EXISTS idx_feeds_name;
DROP INDEX IF EXISTS idx_feeds_url;
DROP TABLE IF EXISTS feeds;