package model

import "github.com/google/uuid"

type Tag struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Priority bool      `json:"priority" db:"priority"`
}

// TagWithArticleID is used when fetching tags for multiple articles in a single query.
// It flattens the join between article_tags and tags.
type TagWithArticleID struct {
	ArticleID uuid.UUID `db:"article_id"`
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Priority  bool      `db:"priority"`
}
