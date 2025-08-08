package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/lucasg04/fyrss-server/internal/repository"
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

func (s *ArticleService) GetFeedPaginated(ctx context.Context, from, to int) ([]*model.Article, error) {
	fullFeed, err := s.repo.GetAllSortedByRecent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed: %w", err)
	}
	if len(fullFeed) == 0 {
		return []*model.Article{}, nil // Return empty slice if no articles found
	}

	sortedFeed := SortFeedArticles(fullFeed)

	fullArticles, err := s.GetArticlesFromMinimal(ctx, sortedFeed[from:to])
	if err != nil {
		return nil, fmt.Errorf("failed to get full articles from minimal feed: %w", err)
	}
	return fullArticles, nil
}

func SortFeedArticles(articles []*model.MinimalFeedArticle) []*model.MinimalFeedArticle {
	// 1) group articles in 12h blocks
	var grouped [][]*model.MinimalFeedArticle
	var currentBlock []*model.MinimalFeedArticle
	blockStart := articles[0].PublishedAt

	for _, article := range articles {
		if blockStart.Sub(article.PublishedAt) > 12*time.Hour {
			grouped = append(grouped, currentBlock)
			currentBlock = []*model.MinimalFeedArticle{}
			blockStart = article.PublishedAt
		}
		currentBlock = append(currentBlock, article)
	}
	if len(currentBlock) > 0 {
		grouped = append(grouped, currentBlock)
	}

	// 2) sort articles by priority and publishedAt
	for _, block := range grouped {
		slices.SortFunc(block, func(a, b *model.MinimalFeedArticle) int {
			// sort unknown priorities to the end
			if a.Priority == model.PriorityUnknown {
				return 1
			}
			if b.Priority == model.PriorityUnknown {
				return -1
			}

			// sort by priority first, then by PublishedAt
			if a.Priority != b.Priority {
				return a.Priority - b.Priority
			}
			return b.PublishedAt.Compare(a.PublishedAt)
		})
	}

	// 3) flatten the grouped articles back to a single slice
	var sortedArticles []*model.MinimalFeedArticle
	for _, block := range grouped {
		sortedArticles = append(sortedArticles, block...)
	}

	return sortedArticles
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
	fullArticlesMap, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get full articles by minimal articles: %w", err)
	}

	// Restore original order
	fullArticles := make([]*model.Article, 0, len(articles))
	for _, article := range articles {
		if fullArticle, exists := fullArticlesMap[article.ID]; exists {
			fullArticles = append(fullArticles, fullArticle)
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

	article.Tags = s.determineTags(ctx, article)
	article.Priority, err = s.determinePriority(ctx, article)
	if err != nil {
		fmt.Printf("failed to determine article priority: %v\n", err)
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

func (s *ArticleService) determinePriority(ctx context.Context, article *model.Article) (int, error) {
	systemPrompt := `You are an expert news analyst. Your task is to assess the public importance of a news article based on its title and description. Rate its importance on a scale from 1 to 5, where:
		1 = Very important (national impact, breaking news, major events)
		2 = Important (significant developments, policy decisions, large public interest)
		3 = Moderate (niche relevance, regional importance, medium impact)
		4 = Low importance (minor updates, limited audience)
		5 = Very unimportant (celebrity gossip, clickbait, trivial matters)
		Consider societal impact, urgency, and relevance. Respond with a single JSON object only, like: {"priority": 2}
		Do not include any text, explanation, or markdown. Only return the JSON.`
	prompt := "Title: " + article.Title + "\nDescription: " + article.Description
	response, err := s.aiService.Generate(ctx, systemPrompt, prompt)
	if err != nil {
		fmt.Printf("Error generating priority: %v\n", err)
		return model.PriorityUnknown, fmt.Errorf("failed to generate priority: %w", err)
	}
	priority, err := readJsonPriority(response)
	if err != nil {
		fmt.Printf("Error reading JSON priority: %v\n", err)
		return model.PriorityUnknown, fmt.Errorf("failed to read priority from JSON: %w", err)
	}
	return priority, nil
}

func readJsonPriority(response string) (int, error) {
	var result struct {
		Priority int `json:"priority"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return model.PriorityUnknown, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}
	if result.Priority < 1 || result.Priority > 5 {
		return model.PriorityUnknown, fmt.Errorf("priority must be between 1 and 5, got %d", result.Priority)
	}
	return result.Priority, nil
}
