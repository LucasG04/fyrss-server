package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lucasg04/fyrss-server/internal/model"
)

type FeedRepository struct {
	db *sqlx.DB
}

func NewFeedRepository(db *sqlx.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

func (r *FeedRepository) GetAll(ctx context.Context) ([]*model.Feed, error) {
	query := "SELECT * FROM feeds ORDER BY created_at DESC"
	var feeds []*model.Feed
	err := r.db.SelectContext(ctx, &feeds, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all feeds: %w", err)
	}
	// Ensure empty slice, not nil, if no results
	if feeds == nil {
		feeds = []*model.Feed{}
	}
	return feeds, nil
}

func (r *FeedRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Feed, error) {
	query := "SELECT * FROM feeds WHERE id = $1"
	var feed model.Feed
	err := r.db.GetContext(ctx, &feed, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed by ID: %w", err)
	}
	return &feed, nil
}

func (r *FeedRepository) Create(ctx context.Context, feed *model.Feed) (*model.Feed, error) {
	query := `
		INSERT INTO feeds (id, name, url, created_at, updated_at)
		VALUES (:id, :name, :url, :created_at, :updated_at)
		RETURNING id`
	var returnedID uuid.UUID
	rows, err := r.db.NamedQueryContext(ctx, query, feed)
	if err != nil {
		return nil, fmt.Errorf("failed to create feed: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&returnedID); err != nil {
			return nil, fmt.Errorf("failed to scan returned id: %w", err)
		}
		feed.ID = returnedID
		return feed, nil
	}
	return nil, fmt.Errorf("failed to create feed: no ID returned")
}

func (r *FeedRepository) Update(ctx context.Context, id uuid.UUID, feed *model.Feed) (*model.Feed, error) {
	query := `
		UPDATE feeds
		SET name = $2, url = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, url, created_at, updated_at`
	var updatedFeed model.Feed
	err := r.db.GetContext(ctx, &updatedFeed, query, id, feed.Name, feed.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to update feed with ID %s: %w", id, err)
	}
	return &updatedFeed, nil
}

func (r *FeedRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM feeds WHERE id = $1"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete feed with ID %s: %w", id, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for feed deletion: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("feed with ID %s not found", id)
	}
	return nil
}

func (r *FeedRepository) IsURLExists(ctx context.Context, url string, excludeID *uuid.UUID) (bool, error) {
	var query string
	var args []interface{}

	if excludeID != nil {
		query = "SELECT COUNT(*) FROM feeds WHERE url = $1 AND id != $2"
		args = []interface{}{url, *excludeID}
	} else {
		query = "SELECT COUNT(*) FROM feeds WHERE url = $1"
		args = []interface{}{url}
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to check if feed URL exists: %w", err)
	}
	return count > 0, nil
}
