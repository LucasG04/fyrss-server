-- Create normalized tag tables and migrate existing data from articles.tags (TEXT[])

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL,
    priority BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS article_tags (
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, tag_id)
);

-- Backfill tags table from existing articles.tags array
INSERT INTO tags (name)
SELECT DISTINCT TRIM(t) AS name
FROM articles a
       CROSS JOIN LATERAL UNNEST(a.tags) AS t
WHERE t IS NOT NULL AND TRIM(t) <> ''
ON CONFLICT (name) DO NOTHING;

-- Backfill article_tags relations
INSERT INTO article_tags (article_id, tag_id)
SELECT a.id, tg.id
FROM articles a
       CROSS JOIN LATERAL UNNEST(a.tags) AS t(name)
       JOIN tags tg ON tg.name = TRIM(t.name)
ON CONFLICT DO NOTHING;

-- Drop old denormalized column
ALTER TABLE articles DROP COLUMN IF EXISTS tags;
