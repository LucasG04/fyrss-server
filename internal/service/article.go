package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/lucasg04/fyrss-server/internal/repository"
)

var ErrDuplicateArticle = errors.New("duplicate article found")

type ArticleService struct {
	repo       *repository.ArticleRepository
	tagService *TagService
	aiService  *AiService
}

func NewArticleService(repo *repository.ArticleRepository, tagService *TagService, aiService *AiService) *ArticleService {
	return &ArticleService{repo: repo, tagService: tagService, aiService: aiService}
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

	tagMap, err := s.GetTagLookupByArticles(ctx, fullFeed)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag lookup by articles: %w", err)
	}
	sortedFeed := s.SortFeedArticles(ctx, fullFeed, tagMap)

	fullArticles, err := s.GetArticlesFromMinimal(ctx, sortedFeed[from:to], tagMap)
	if err != nil {
		return nil, fmt.Errorf("failed to get full articles from minimal feed: %w", err)
	}
	return fullArticles, nil
}

func (s *ArticleService) SortFeedArticles(ctx context.Context, articles []*model.MinimalFeedArticle, tagMap map[uuid.UUID][]*model.Tag) []*model.MinimalFeedArticle {
	if len(articles) == 0 {
		return []*model.MinimalFeedArticle{}
	}
	now := time.Now()

	type scored struct {
		art   *model.MinimalFeedArticle
		score float64
	}

	// Tunable parameters
	const (
		halfLifeHours = 24.0
		descIdealMin  = 80
		descIdealMax  = 300
		wTag          = 0.4
		wRecency      = 0.4
		wDesc         = 0.2
	)

	items := make([]scored, 0, len(articles))
	for _, a := range articles {
		if a == nil {
			continue
		}

		tags := tagMap[a.ID]
		var priorityCount float64
		for _, t := range tags {
			if t.Priority {
				priorityCount += 1
			}
		}
		// Compress impact of many priority tags (0..~1)
		tagScore := math.Tanh(priorityCount)

		// Recency exponential decay (0..1)
		var recencyScore float64
		ageHours := now.Sub(a.PublishedAt).Hours()
		if ageHours < 0 {
			ageHours = 0
		}
		recencyScore = math.Exp(-ageHours / halfLifeHours)

		// Description score: encourage medium length, penalize extremes
		descLen := len(strings.TrimSpace(a.Description))
		var descScore float64
		switch {
		case descLen < descIdealMin:
			descScore = (float64(descLen) / float64(descIdealMin)) * 0.7
		case descLen <= descIdealMax:
			descScore = 1
		default:
			overflow := float64(descLen - descIdealMax)
			descScore = 1 / (1 + overflow/400) // saturating penalty
		}

		total := wTag*tagScore + wRecency*recencyScore + wDesc*descScore
		items = append(items, scored{art: a, score: total})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score == items[j].score {
			if !items[i].art.PublishedAt.Equal(items[j].art.PublishedAt) {
				return items[i].art.PublishedAt.After(items[j].art.PublishedAt)
			}
			// Deterministic fallback by ID
			return items[i].art.ID.String() > items[j].art.ID.String()
		}
		return items[i].score > items[j].score
	})

	out := make([]*model.MinimalFeedArticle, len(items))
	for i, item := range items {
		out[i] = item.art
	}
	return out
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

	tagMap, err := s.GetTagLookupByArticles(ctx, articles)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag lookup by articles: %w", err)
	}

	fullArticles, err := s.GetArticlesFromMinimal(ctx, articles[from:to], tagMap)
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

	tagMap, err := s.GetTagLookupByArticles(ctx, articles)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag lookup by articles: %w", err)
	}

	fullArticles, err := s.GetArticlesFromMinimal(ctx, articles[from:to], tagMap)
	if err != nil {
		return nil, fmt.Errorf("failed to get full articles from minimal saved: %w", err)
	}
	return fullArticles, nil
}

func (s *ArticleService) GetArticlesFromMinimal(ctx context.Context, articles []*model.MinimalFeedArticle, tagMap map[uuid.UUID][]*model.Tag) ([]*model.Article, error) {
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

	// fill tags for json
	for i, article := range fullArticles {
		if tags, exists := tagMap[article.ID]; exists {
			// Convert model.Tag to string slice
			fullArticles[i].Tags = make([]string, len(tags))
			for j, tag := range tags {
				fullArticles[i].Tags[j] = tag.Name
			}
			slices.Sort(fullArticles[i].Tags)
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

func (s *ArticleService) GetTagLookupByArticles(ctx context.Context, articles []*model.MinimalFeedArticle) (map[uuid.UUID][]*model.Tag, error) {
	// Batch fetch tags to avoid N+1 queries
	tagMap, err := s.tagService.GetTagsOfArticles(ctx, articles)
	if err != nil {
		// Fallback: continue without tag influence
		tagMap = map[uuid.UUID][]*model.Tag{}
	}
	return tagMap, nil
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

	savedArticle, err := s.repo.Save(ctx, article)
	if err != nil {
		return fmt.Errorf("failed to save article: %w", err)
	}

	tags := s.determineTags(ctx, savedArticle)
	if len(tags) > 0 {
		err = s.tagService.AssignTagsToArticle(ctx, savedArticle.ID, tags)
		if err != nil {
			return fmt.Errorf("failed to assign tags to article: %w", err)
		}
	}
	return nil
}

func (s *ArticleService) determineTags(ctx context.Context, item *model.Article) []string {
	// 1) get current tags from db
	tags, err := s.tagService.GetAll(ctx)
	if err != nil {
		fmt.Printf("Error getting unique tags: %v\n", err)
		return []string{}
	}

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	// 2) use AI to generate tags based on the item
	systemPrompt := `You are an expert in categorizing news articles.
Assign 1–3 high-level, general tags (broad topics or news sections) to the article.
Use only tags from this list: ` + fmt.Sprintf("%v", tagNames) + `.
If none apply, create 1–3 new general tags.

Rules:
- Tags must be in the article's original language (do not translate).
- No specific events or names; keep tags broad (e.g., "Politics" not "US Election").
- Return ONLY a valid JSON object: {"tags": ["tag1", "tag2"]}
- No text, notes, or formatting outside the JSON.
- Invalid output is not allowed.`
	prompt := "Title: " + item.Title + "\nDescription: " + item.Description
	response, err := s.aiService.Generate(ctx, systemPrompt, prompt)
	if err != nil {
		fmt.Printf("Error generating tags: %v\n", err)
		return []string{}
	}
	tagNames, err = readJsonTags(response)
	if err != nil {
		fmt.Printf("Error reading JSON tags: %v\n", err)
		return []string{}
	}
	return tagNames
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
