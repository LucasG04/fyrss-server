package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/lucasg04/fyrss-server/internal/repository"
	"github.com/mmcdole/gofeed"
)

var (
	ErrFeedNotFound       = errors.New("feed not found")
	ErrDuplicateFeedURL   = errors.New("feed URL already exists")
	ErrInvalidFeedURL     = errors.New("invalid feed URL")
	ErrInvalidFeedName    = errors.New("feed name cannot be empty")
	ErrInvalidRSSFeed     = errors.New("URL does not return a valid RSS/Atom feed")
	ErrFeedValidationFail = errors.New("feed validation failed")
)

type FeedService struct {
	repo           *repository.FeedRepository
	rssReader      *RssArticleReader
	articleService *ArticleService
}

func NewFeedService(repo *repository.FeedRepository, rssReader *RssArticleReader, articleService *ArticleService) *FeedService {
	return &FeedService{
		repo:           repo,
		rssReader:      rssReader,
		articleService: articleService,
	}
}

func (s *FeedService) GetAll(ctx context.Context) ([]*model.Feed, error) {
	feeds, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all feeds: %w", err)
	}
	return feeds, nil
}

func (s *FeedService) GetByID(ctx context.Context, id uuid.UUID) (*model.Feed, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid feed ID: %s", id)
	}

	feed, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed with ID %s: %w", id, err)
	}
	return feed, nil
}

func (s *FeedService) Create(ctx context.Context, req *model.CreateFeedRequest) (*model.Feed, error) {
	if err := s.validateFeedRequest(req.Name, req.URL); err != nil {
		return nil, err
	}

	// Validate that the URL actually returns a valid RSS feed
	if err := s.validateRSSFeed(ctx, req.URL); err != nil {
		return nil, err
	}

	// Check if URL already exists
	exists, err := s.repo.IsURLExists(ctx, req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate feed URL: %w", err)
	}
	if exists {
		return nil, ErrDuplicateFeedURL
	}

	now := time.Now()
	feed := &model.Feed{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(req.Name),
		URL:       strings.TrimSpace(req.URL),
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdFeed, err := s.repo.Create(ctx, feed)
	if err != nil {
		return nil, fmt.Errorf("failed to create feed: %w", err)
	}

	// Automatically fetch and process the feed after creation
	go s.processFeedAsync(context.Background(), createdFeed)

	return createdFeed, nil
}

func (s *FeedService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateFeedRequest) (*model.Feed, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid feed ID: %s", id)
	}

	if err := s.validateFeedRequest(req.Name, req.URL); err != nil {
		return nil, err
	}

	// Validate that the URL actually returns a valid RSS feed
	if err := s.validateRSSFeed(ctx, req.URL); err != nil {
		return nil, err
	}

	// Check if URL already exists for a different feed
	exists, err := s.repo.IsURLExists(ctx, req.URL, &id)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate feed URL: %w", err)
	}
	if exists {
		return nil, ErrDuplicateFeedURL
	}

	feed := &model.Feed{
		Name: strings.TrimSpace(req.Name),
		URL:  strings.TrimSpace(req.URL),
	}

	updatedFeed, err := s.repo.Update(ctx, id, feed)
	if err != nil {
		return nil, fmt.Errorf("failed to update feed with ID %s: %w", id, err)
	}

	// Automatically fetch and process the feed after update
	go s.processFeedAsync(context.Background(), updatedFeed)

	return updatedFeed, nil
}

func (s *FeedService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid feed ID: %s", id)
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete feed with ID %s: %w", id, err)
	}

	return nil
}

func (s *FeedService) GetAllURLs(ctx context.Context) ([]string, error) {
	feeds, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all feeds: %w", err)
	}

	urls := make([]string, len(feeds))
	for i, feed := range feeds {
		urls[i] = feed.URL
	}
	return urls, nil
}

// GetByURL retrieves a feed by its URL
func (s *FeedService) GetByURL(ctx context.Context, url string) (*model.Feed, error) {
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	feed, err := s.repo.GetByURL(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed with URL %s: %w", url, err)
	}
	return feed, nil
}

// ValidateFeedURL provides a way to validate an RSS feed URL without creating a feed
// This can be useful for testing or admin purposes
func (s *FeedService) ValidateFeedURL(ctx context.Context, feedURL string) error {
	// First validate URL format
	if strings.TrimSpace(feedURL) == "" {
		return ErrInvalidFeedURL
	}

	parsedURL, err := url.Parse(feedURL)
	if err != nil {
		return ErrInvalidFeedURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidFeedURL
	}

	if parsedURL.Host == "" {
		return ErrInvalidFeedURL
	}

	// Then validate RSS feed
	return s.validateRSSFeed(ctx, feedURL)
}

func (s *FeedService) validateFeedRequest(name, feedURL string) error {
	// Validate name
	if strings.TrimSpace(name) == "" {
		return ErrInvalidFeedName
	}

	// Validate URL
	if strings.TrimSpace(feedURL) == "" {
		return ErrInvalidFeedURL
	}

	parsedURL, err := url.Parse(feedURL)
	if err != nil {
		return ErrInvalidFeedURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidFeedURL
	}

	if parsedURL.Host == "" {
		return ErrInvalidFeedURL
	}

	return nil
}

// validateRSSFeed checks if the given URL returns a valid RSS/Atom feed
func (s *FeedService) validateRSSFeed(ctx context.Context, feedURL string) error {
	// Create a context with timeout for the RSS validation
	validateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use gofeed parser to attempt to parse the feed
	fp := gofeed.NewParser()
	fp.UserAgent = "Fyrss-Server/1.0 (+https://github.com/LucasG04/fyrss-server)"

	// Parse the feed URL with context
	feed, err := fp.ParseURLWithContext(feedURL, validateCtx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRSSFeed, err)
	}

	// Check if we got a valid feed response
	if feed == nil {
		return fmt.Errorf("%w: feed is empty", ErrInvalidRSSFeed)
	}

	// Check if the feed has a title (basic requirement for valid feeds)
	if strings.TrimSpace(feed.Title) == "" {
		return fmt.Errorf("%w: feed has no title", ErrInvalidRSSFeed)
	}

	// Additional validation: ensure the feed has at least basic structure
	// We don't require items as some feeds might be empty but still valid
	if feed.FeedType == "" {
		return fmt.Errorf("%w: unable to determine feed type", ErrInvalidRSSFeed)
	}

	return nil
}

// processFeedAsync automatically processes a feed in the background
// This method runs asynchronously to avoid blocking API responses
func (s *FeedService) processFeedAsync(ctx context.Context, feed *model.Feed) {
	// Set a timeout for feed processing to avoid hanging
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := s.ProcessFeedNow(processCtx, feed)
	if err != nil {
		// Log error but don't fail the API response since this is async
		fmt.Printf("Background feed processing failed for feed %s (%s): %v\n", feed.Name, feed.URL, err)
	}
}

// ProcessFeedNow immediately fetches and processes articles from a feed
func (s *FeedService) ProcessFeedNow(ctx context.Context, feed *model.Feed) error {
	if feed == nil {
		return fmt.Errorf("feed cannot be nil")
	}

	// Read articles from the feed
	articles, err := s.rssReader.ReadFeed(ctx, feed)
	if err != nil {
		return fmt.Errorf("failed to read feed %s: %w", feed.URL, err)
	}

	if len(articles) == 0 {
		return fmt.Errorf("no articles found in feed %s", feed.URL)
	}

	// Save articles to database
	savedCount := 0
	duplicateCount := 0

	for _, article := range articles {
		err := s.articleService.Save(ctx, article)
		if err == ErrDuplicateArticle {
			duplicateCount++
			continue
		}
		if err != nil {
			// Log individual article save errors but continue processing
			fmt.Printf("Failed to save article '%s' from feed %s: %v\n", article.Title, feed.URL, err)
			continue
		}
		savedCount++
	}

	fmt.Printf("Processed feed %s (%s): saved %d new articles, skipped %d duplicates\n",
		feed.Name, feed.URL, savedCount, duplicateCount)

	return nil
}

// ProcessFeedByID processes a feed by its ID
func (s *FeedService) ProcessFeedByID(ctx context.Context, feedID uuid.UUID) error {
	feed, err := s.GetByID(ctx, feedID)
	if err != nil {
		return fmt.Errorf("failed to get feed for processing: %w", err)
	}

	return s.ProcessFeedNow(ctx, feed)
}
