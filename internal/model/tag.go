package model

import "github.com/google/uuid"

type Tag struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Priority bool      `json:"priority" db:"priority"`
}
