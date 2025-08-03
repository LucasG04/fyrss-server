package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lucasg04/nuntium/internal/model"
	"github.com/lucasg04/nuntium/internal/repository"
)

var ErrDuplicateArticle = errors.New("duplicate article found")

type ArticleService struct {
	repo      repository.ArticleRepository
	aiService *AiService
}

func NewArticleService(repo *repository.ArticleRepository, aiService *AiService) *ArticleService {
	return &ArticleService{repo: *repo, aiService: aiService}
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

func (s *ArticleService) GetAllUniqueTags(ctx context.Context) ([]string, error) {
	tags, err := s.repo.GetAllUniqueTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all unique tags: %w", err)
	}
	return tags, nil
}

func (s *ArticleService) DeleteOneWeekOldArticles(ctx context.Context) error {
	err := s.repo.DeleteOneWeekOldArticles(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete articles older than one week: %w", err)
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

	if len(article.Tags) == 0 {
		// If no tags are provided, generate them using AI
		article.Tags = s.determineTags(ctx, article)
	}

	err = s.repo.Save(ctx, article)
	if err != nil {
		return fmt.Errorf("failed to save article: %w", err)
	}
	return nil
}

func (s *ArticleService) determineTags(ctx context.Context, item *model.Article) []string {
	// 1) get current tags from db
	tags, err := s.GetAllUniqueTags(ctx)
	if err != nil {
		fmt.Printf("Error getting unique tags: %v\n", err)
		return []string{}
	}

	// 2) use AI to generate tags based on the item
	systemPrompt := "You are an expert in categorizing news articles. Your task is to assign 1–3 very general, high-level tags (like topic categories or news sections) to each article. Use only tags from this predefined list: " + fmt.Sprintf("%v", tags) + ". If no tags apply, generate 1–3 new general tags. Important: The tags must be written in the same language as the article (e.g., 'Politik' for German, 'Politics' for English). Do not translate. Respond strictly as a JSON object. Do not include any text or markdown. Tags must be written in the same language as the article (e.g., use German tags for German articles)."
	prompt := "Title: '" + item.Title + "' Description: '" + item.Description + "'\nReturn a plain JSON object like: '{\"tags\": [\"tag1\", \"tag2\"]}' No other output. No markdown. Only JSON."
	response, err := s.aiService.Generate(ctx, systemPrompt, prompt)
	if err != nil {
		fmt.Printf("Error generating tags: %v\n", err)
		return []string{}
	}
	tags, err = readJsonTags(response)
	if err != nil {
		fmt.Printf("Error reading JSON tags: %v\n", err)
		return []string{}
	}
	return tags
}

func readJsonTags(response string) ([]string, error) {
	var result struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}
	return result.Tags, nil
}
