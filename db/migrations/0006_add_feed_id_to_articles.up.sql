-- Add feed_id column to articles table to establish 1:m relationship
ALTER TABLE articles ADD COLUMN feed_id UUID;

-- Add foreign key constraint (feeds must exist before being referenced)
ALTER TABLE articles ADD CONSTRAINT fk_articles_feed_id 
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE SET NULL;

-- Index for performance when querying articles by feed
CREATE INDEX idx_articles_feed_id ON articles(feed_id);