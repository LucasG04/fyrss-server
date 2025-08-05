package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const PriorityUnknown = 0

var DefaultNilTime time.Time

type Article struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	ContentHash string    `json:"-" db:"content_hash"`
	SourceUrl   string    `json:"source_url" db:"source_url"`
	// SourceType indicates the type of source. "rss" or "scraped"
	SourceType  string         `json:"source_type" db:"source_type"`
	Priority    int            `json:"priority" db:"priority"`
	Tags        pq.StringArray `json:"tags" db:"tags"`
	PublishedAt time.Time      `json:"published_at" db:"published_at"`
	LastReadAt  time.Time      `json:"last_read_at" db:"last_read_at"`
	Save        bool           `json:"save" db:"save"`
}

type MinimalFeedArticle struct {
	ID          uuid.UUID `db:"id"`
	PublishedAt time.Time `db:"published_at"`
	Priority    int       `db:"priority"`
}
