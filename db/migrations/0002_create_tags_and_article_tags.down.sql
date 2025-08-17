-- Recreate tags column in articles and drop normalized tables

ALTER TABLE articles ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';

-- Rehydrate articles.tags from normalized tables (best-effort)
UPDATE articles a
SET tags = (
  SELECT COALESCE(ARRAY_AGG(t.name ORDER BY t.name), '{}')
  FROM article_tags at
  JOIN tags t ON t.id = at.tag_id
  WHERE at.article_id = a.id
);

DROP TABLE IF EXISTS article_tags;
DROP TABLE IF EXISTS tags;
