package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lucasg04/fyrss-server/internal/model"
)

type TagRepository struct {
	db *sqlx.DB
}

func NewTagRepository(db *sqlx.DB) *TagRepository {
	return &TagRepository{db: db}
}

// GetAllTags returns all tag names from normalized tags table
func (t *TagRepository) GetAllTags(ctx context.Context) (tags []string, err error) {
	query := "SELECT name FROM tags ORDER BY name"
	err = t.db.SelectContext(ctx, &tags, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all unique tags: %w", err)
	}
	return tags, nil
}

func (t *TagRepository) GetTagByID(ctx context.Context, id uuid.UUID) (*model.Tag, error) {
	query := "SELECT * FROM tags WHERE id = $1"
	var tag model.Tag
	if err := t.db.GetContext(ctx, &tag, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tag %q: %w", id, err)
	}
	return &tag, nil
}

func (t *TagRepository) ExistsByID(ctx context.Context, id uuid.UUID) (exists bool, err error) {
	query := "SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)"
	if err = t.db.GetContext(ctx, &exists, query, id); err != nil {
		return false, fmt.Errorf("failed to check if tag exists %q: %w", id, err)
	}
	return exists, nil
}

func (t *TagRepository) GetByName(ctx context.Context, name string) (*model.Tag, error) {
	query := "SELECT * FROM tags WHERE name = $1"
	var tag model.Tag
	if err := t.db.GetContext(ctx, &tag, query, name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tag %q: %w", name, err)
	}
	return &tag, nil
}

func (t *TagRepository) GetTagsOfArticle(ctx context.Context, articleID uuid.UUID) ([]*model.Tag, error) {
	query := "SELECT t.* FROM tags t INNER JOIN article_tags at ON t.id = at.tag_id WHERE at.article_id = $1"
	var tags []*model.Tag
	if err := t.db.SelectContext(ctx, &tags, query, articleID); err != nil {
		return nil, fmt.Errorf("failed to get tags for article %q: %w", articleID, err)
	}
	return tags, nil
}

func (t *TagRepository) CreateTag(ctx context.Context, name string) (tag *model.Tag, err error) {
	query := "INSERT INTO tags (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING"
	tagId := uuid.New()
	if _, err := t.db.ExecContext(ctx, query, tagId, name); err != nil {
		return nil, fmt.Errorf("failed to create tag %q: %w", name, err)
	}
	return &model.Tag{ID: tagId, Name: name, Priority: false}, nil
}

func (t *TagRepository) UpdateTag(ctx context.Context, tag *model.Tag) error {
	query := "UPDATE tags SET name = $2, priority = $3 WHERE id = $1"
	_, err := t.db.ExecContext(ctx, query, tag.ID, tag.Name, tag.Priority)
	if err != nil {
		return fmt.Errorf("failed to update tag %q: %w", tag.ID, err)
	}
	return nil
}

func (t *TagRepository) AssignTagsToArticle(ctx context.Context, articleID uuid.UUID, tagIDs []uuid.UUID) error {
	query := "INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2)"
	for _, tagID := range tagIDs {
		if _, err := t.db.ExecContext(ctx, query, articleID, tagID); err != nil {
			return fmt.Errorf("failed to assign tag %q to article %q: %w", tagID, articleID, err)
		}
	}
	return nil
}
