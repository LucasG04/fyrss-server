-- Restore priority and tags columns to articles table
ALTER TABLE articles ADD COLUMN IF NOT EXISTS priority INT NOT NULL DEFAULT 0 CHECK (priority >= 0 AND priority <= 5);
ALTER TABLE articles ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';