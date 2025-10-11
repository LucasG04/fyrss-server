package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/lucasg04/fyrss-server/internal/repository"
)

var ErrDuplicateArticle = errors.New("duplicate article found")

type ArticleService struct {
	repo *repository.ArticleRepository
}

func NewArticleService(repo *repository.ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
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

func (s *ArticleService) GetFeedPaginated(ctx context.Context, from, to int) ([]*model.Article, error) {
	fullFeed, err := s.repo.GetAllSortedByRecent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed: %w", err)
	}
	if len(fullFeed) == 0 {
		return []*model.Article{}, nil // Return empty slice if no articles found
	}
	// adjust from & to params if articles length doesnt match
	if from < 0 || to > len(fullFeed) || from >= to {
		from = 0
		to = len(fullFeed)
	}

	sortedFeed := s.SortFeedArticles(ctx, fullFeed)

	fullArticles, err := s.GetArticlesFromMinimal(ctx, sortedFeed[from:to])
	if err != nil {
		return nil, fmt.Errorf("failed to get full articles from minimal feed: %w", err)
	}
	return fullArticles, nil
}

func (s *ArticleService) SortFeedArticles(ctx context.Context, articles []*model.MinimalFeedArticle) []*model.MinimalFeedArticle {
	if len(articles) == 0 {
		return []*model.MinimalFeedArticle{}
	}

	// Simple sort by published date (newest first)
	sort.SliceStable(articles, func(i, j int) bool {
		if !articles[i].PublishedAt.Equal(articles[j].PublishedAt) {
			return articles[i].PublishedAt.After(articles[j].PublishedAt)
		}
		// Deterministic fallback by ID
		return articles[i].ID.String() > articles[j].ID.String()
	})

	return articles
}

func (s *ArticleService) DeleteOneWeekOldArticles(ctx context.Context) error {
	err := s.repo.DeleteOneWeekOldArticles(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete articles older than one week: %w", err)
	}
	return nil
}

func (s *ArticleService) GetHistoryPaginated(ctx context.Context, from, to int) ([]*model.Article, error) {
	articles, err := s.repo.GetFullHistorySorted(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get article history: %w", err)
	}

	if len(articles) == 0 {
		return []*model.Article{}, nil // Return empty slice if no articles found
	}
	// adjust from & to params if articles length doesnt match
	if from < 0 || to > len(articles) || from >= to {
		from = 0
		to = len(articles)
	}

	fullArticles, err := s.GetArticlesFromMinimal(ctx, articles[from:to])
	if err != nil {
		return nil, fmt.Errorf("failed to get full articles from minimal history: %w", err)
	}
	return fullArticles, nil
}

func (s *ArticleService) GetSavedPaginated(ctx context.Context, from, to int) ([]*model.Article, error) {
	articles, err := s.repo.GetAllSavedSorted(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get saved articles: %w", err)
	}

	if len(articles) == 0 {
		return []*model.Article{}, nil // Return empty slice if no articles found
	}
	// adjust from & to params if articles length doesnt match
	if from < 0 || to > len(articles) || from >= to {
		from = 0
		to = len(articles)
	}

	fullArticles, err := s.GetArticlesFromMinimal(ctx, articles[from:to])
	if err != nil {
		return nil, fmt.Errorf("failed to get full articles from minimal saved: %w", err)
	}
	return fullArticles, nil
}

func (s *ArticleService) GetArticlesFromMinimal(ctx context.Context, articles []*model.MinimalFeedArticle) ([]*model.Article, error) {
	if len(articles) == 0 {
		return []*model.Article{}, nil // Return empty slice if no articles provided
	}

	// Extract IDs from minimal articles
	ids := make([]uuid.UUID, len(articles))
	for i, article := range articles {
		ids[i] = article.ID
	}

	// Fetch full articles by IDs
	fullUnsortedArticles, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get full articles by minimal articles: %w", err)
	}
	fullArticlesMap := make(map[uuid.UUID]*model.Article, len(articles))
	for _, article := range fullUnsortedArticles {
		fullArticlesMap[article.ID] = article
	}

	// Restore original order
	fullArticles := make([]*model.Article, len(articles))
	for i, article := range articles {
		if fullArticle, exists := fullArticlesMap[article.ID]; exists {
			fullArticles[i] = fullArticle
		}
	}

	return fullArticles, nil
}

func (s *ArticleService) UpdateSavedByID(ctx context.Context, id uuid.UUID, saved bool) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid article ID: %s", id)
	}

	err := s.repo.UpdateSavedByID(ctx, id, saved)
	if err != nil {
		return fmt.Errorf("failed to update saved status for article ID %s: %w", id, err)
	}
	return nil
}

func (s *ArticleService) Save(ctx context.Context, article *model.Article) error {
	if article == nil {
		return fmt.Errorf("article cannot be nil")
	}

	// Check if the article is a duplicate
	isDuplicate, err := s.repo.IsDuplicate(ctx, article.ContentHash)
	if err != nil {
		return fmt.Errorf("failed to check for duplicate article: %w", err)
	}
	if isDuplicate {
		return ErrDuplicateArticle
	}

	_, err = s.repo.Save(ctx, article)
	if err != nil {
		return fmt.Errorf("failed to save article: %w", err)
	}

	return nil
}

func (s *ArticleService) UpdateReadByID(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid article ID: %s", id)
	}

	err := s.repo.UpdateReadByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to update read status for article ID %s: %w", id, err)
	}
	return nil
}
