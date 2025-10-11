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
	repo *repository.FeedRepository
}

func NewFeedService(repo *repository.FeedRepository) *FeedService {
	return &FeedService{repo: repo}
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
