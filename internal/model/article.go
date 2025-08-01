package model

import (
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Summary     string    `json:"summary" db:"summary"`
	ContentHash string    `json:"-" db:"content_hash"`
	SourceUrl   string    `json:"source_url" db:"source_url"`
	// SourceType indicates the type of source. "rss" or "scraped"
	SourceType string     `json:"source_type" db:"source_type"`
	Category   string     `json:"category" db:"category"`
	Tags       []string   `json:"tags" db:"tags"`
	Priority   *int       `json:"priority,omitempty" db:"priority"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	LastReadAt *time.Time `json:"last_read_at,omitempty" db:"last_read_at"`
}
