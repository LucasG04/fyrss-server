package model

import (
	"time"

	"github.com/google/uuid"
)

var DefaultNilTime time.Time

type Article struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	ContentHash string    `json:"-" db:"content_hash"`
	SourceUrl   string    `json:"sourceUrl" db:"source_url"`
	// SourceType indicates the type of source. "rss" or "scraped"
	SourceType  string     `json:"sourceType" db:"source_type"`
	PublishedAt time.Time  `json:"publishedAt" db:"published_at"`
	LastReadAt  time.Time  `json:"lastReadAt" db:"last_read_at"`
	Save        bool       `json:"save" db:"save"`
	FeedID      *uuid.UUID `json:"feedId,omitempty" db:"feed_id"`
}

type MinimalFeedArticle struct {
	ID          uuid.UUID `db:"id"`
	Description string    `db:"description"`
	PublishedAt time.Time `db:"published_at"`
}
