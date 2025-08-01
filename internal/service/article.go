package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lucasg04/nuntium/internal/model"
	"github.com/lucasg04/nuntium/internal/repository"
)

type ArticleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo *repository.ArticleRepository) *ArticleService {
	return &ArticleService{repo: *repo}
}

func (s *ArticleService) GetAll(ctx context.Context) ([]*model.Article, error) {
	articles, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all articles: %w", err)
	}
	return articles, nil
}

func (s *ArticleService) GetByID(ctx context.Context, id uuid.UUID) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get article with ID %s: %w", id, err)
	}
	return article, nil
}
