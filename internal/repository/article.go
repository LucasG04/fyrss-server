package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lucasg04/nuntium/internal/model"
)

type ArticleRepository struct {
	db *sqlx.DB
}

func NewArticleRepository(db *sqlx.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) GetAll(ctx context.Context) ([]*model.Article, error) {
	query := "SELECT * FROM articles"
	var articles []*model.Article
	err := r.db.SelectContext(ctx, &articles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all articles: %w", err)
	}
	// Ensure empty slice, not nil, if no results
	if articles == nil {
		articles = []*model.Article{}
	}
	return articles, nil
}

func (r *ArticleRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Article, error) {
	query := "SELECT * FROM articles WHERE id = $1"
	var article model.Article
	err := r.db.GetContext(ctx, &article, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get article by ID: %w", err)
	}
	return &article, nil
}
