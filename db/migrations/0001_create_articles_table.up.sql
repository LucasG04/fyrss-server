-- Extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL UNIQUE,
    source_url TEXT NOT NULL,
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('rss', 'scraped')),
    priority INT NOT NULL DEFAULT 0 CHECK (priority >= 0 AND priority <= 5),
    tags TEXT[] DEFAULT '{}',
    published_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_read_at TIMESTAMP WITH TIME ZONE,
    save BOOLEAN DEFAULT FALSE
);

-- index for content deduplication
CREATE UNIQUE INDEX idx_articles_content_hash ON articles(content_hash);