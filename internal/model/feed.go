package model

import (
	"time"

	"github.com/google/uuid"
)

type Feed struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	URL       string    `json:"url" db:"url"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type CreateFeedRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type UpdateFeedRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
