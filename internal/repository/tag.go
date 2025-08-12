package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lucasg04/fyrss-server/internal/model"
)

type TagRepository struct {
	db *sqlx.DB
}

func NewTagRepository(db *sqlx.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (t *TagRepository) GetAllTags(ctx context.Context) (tags []string, err error) {
	query := "SELECT DISTINCT unnest(tags) AS tag FROM articles"
	err = t.db.SelectContext(ctx, &tags, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all unique tags: %w", err)
	}
	return tags, nil
}

func (t *TagRepository) GetTagsWithWeights(ctx context.Context) ([]*model.WeightedTag, error) {
	var weightedTags []*model.WeightedTag
	query := "SELECT name, weight FROM weighted_tags"
	err := t.db.SelectContext(ctx, &weightedTags, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag weight: %w", err)
	}
	return weightedTags, nil
}

func (t *TagRepository) GetWeightedTag(ctx context.Context, tag string) (*model.WeightedTag, error) {
	weightedTag := &model.WeightedTag{}
	query := "SELECT name, weight FROM weighted_tags WHERE name = $1"
	err := t.db.GetContext(ctx, weightedTag, query, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get weight for tag %q: %w", tag, err)
	}
	return weightedTag, nil
}

func (t *TagRepository) CreateWeightedTag(ctx context.Context, tag string, weight int) error {
	query := "INSERT INTO weighted_tags (name, weight) VALUES ($1, $2)"
	_, err := t.db.ExecContext(ctx, query, tag, weight)
	if err != nil {
		return fmt.Errorf("failed to create tag weight: %w", err)
	}
	return nil
}

func (t *TagRepository) SetTagWeight(ctx context.Context, tag string, weight int) error {
	query := "UPDATE weighted_tags SET weight = $1 WHERE name = $2"
	_, err := t.db.ExecContext(ctx, query, weight, tag)
	if err != nil {
		return fmt.Errorf("failed to set tag weight: %w", err)
	}
	return nil
}

func (t *TagRepository) RemoveWeight(ctx context.Context, tag string) error {
	query := "DELETE FROM weighted_tags WHERE name = $1"
	_, err := t.db.ExecContext(ctx, query, tag)
	if err != nil {
		return fmt.Errorf("failed to remove tag weight: %w", err)
	}
	return nil
}
