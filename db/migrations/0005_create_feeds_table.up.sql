-- Create feeds table for managing RSS feed URLs
CREATE TABLE feeds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for unique URL constraint and fast lookups
CREATE UNIQUE INDEX idx_feeds_url ON feeds(url);

-- Index for searching by name
CREATE INDEX idx_feeds_name ON feeds(name);