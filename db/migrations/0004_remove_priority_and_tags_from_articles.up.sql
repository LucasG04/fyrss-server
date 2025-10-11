-- Remove priority and tags columns from articles table
ALTER TABLE articles DROP COLUMN IF EXISTS priority;
ALTER TABLE articles DROP COLUMN IF EXISTS tags;