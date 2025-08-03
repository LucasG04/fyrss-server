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

func (r *ArticleRepository) GetAllUniqueTags(ctx context.Context) (tags []string, err error) {
	query := "SELECT DISTINCT unnest(tags) AS tag FROM articles"
	err = r.db.SelectContext(ctx, &tags, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all unique tags: %w", err)
	}
	return tags, nil
}

func (r *ArticleRepository) IsDuplicate(ctx context.Context, contentHash string) (bool, error) {
	query := "SELECT COUNT(*) FROM articles WHERE content_hash = $1"
	var count int
	err := r.db.GetContext(ctx, &count, query, contentHash)
	if err != nil {
		return false, fmt.Errorf("failed to check for duplicate article: %w", err)
	}
	return count > 0, nil
}

func (r *ArticleRepository) Save(ctx context.Context, article *model.Article) error {
	query := `
		INSERT INTO articles (id, title, description, content_hash, source_url, source_type, tags, published_at, last_read_at, save)
		VALUES (:id, :title, :description, :content_hash, :source_url, :source_type, :tags, :published_at, :last_read_at, :save)
		ON CONFLICT (id) DO NOTHING`
	_, err := r.db.NamedExecContext(ctx, query, article)
	if err != nil {
		return fmt.Errorf("failed to save article: %w", err)
	}
	return nil
}
