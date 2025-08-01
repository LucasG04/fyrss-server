-- Extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    summary TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL UNIQUE,
    source_url TEXT NOT NULL,
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('rss', 'scraped')),
    category VARCHAR(100) NOT NULL,
    tags TEXT[] DEFAULT '{}',
    priority INTEGER CHECK (priority >= 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_read_at TIMESTAMP WITH TIME ZONE
);

-- index for content deduplication
CREATE UNIQUE INDEX idx_articles_content_hash ON articles(content_hash);