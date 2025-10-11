package service

import (
	"context"
	"testing"
	"time"

	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/lucasg04/fyrss-server/internal/repository"
)

// TestValidateRSSFeed_ValidFeed tests RSS validation with a known working RSS feed
func TestValidateRSSFeed_ValidFeed(t *testing.T) {
	// This test requires internet connectivity
	if testing.Short() {
		t.Skip("Skipping RSS validation test in short mode")
	}

	feedService := &FeedService{}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test with a known working RSS feed
	validURL := "https://www.tagesschau.de/index~rss2.xml"
	err := feedService.validateRSSFeed(ctx, validURL)
	if err != nil {
		t.Errorf("Expected valid RSS feed to pass validation, got error: %v", err)
	}
}

// TestValidateRSSFeed_InvalidFeed tests RSS validation with an invalid URL
func TestValidateRSSFeed_InvalidFeed(t *testing.T) {
	// This test requires internet connectivity
	if testing.Short() {
		t.Skip("Skipping RSS validation test in short mode")
	}

	feedService := &FeedService{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with a URL that doesn't return RSS
	invalidURL := "https://www.google.com"
	err := feedService.validateRSSFeed(ctx, invalidURL)
	if err == nil {
		t.Error("Expected invalid RSS URL to fail validation, but it passed")
	}
	if err != ErrInvalidRSSFeed && err.Error() == "" {
		t.Errorf("Expected ErrInvalidRSSFeed or wrapped error, got: %v", err)
	}
}

// TestValidateRSSFeed_NonExistentURL tests RSS validation with a non-existent URL
func TestValidateRSSFeed_NonExistentURL(t *testing.T) {
	// This test requires internet connectivity
	if testing.Short() {
		t.Skip("Skipping RSS validation test in short mode")
	}

	feedService := &FeedService{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with a non-existent URL
	nonExistentURL := "https://this-domain-definitely-does-not-exist-12345.com/rss.xml"
	err := feedService.validateRSSFeed(ctx, nonExistentURL)
	if err == nil {
		t.Error("Expected non-existent URL to fail validation, but it passed")
	}
}

// Example of how the full Create flow works with RSS validation
func ExampleFeedService_Create_withValidation() {
	// Mock repository for example - in real usage this would be a database
	mockRepo := &repository.FeedRepository{}
	mockRssReader := &RssArticleReader{}
	mockArticleService := &ArticleService{}
	feedService := NewFeedService(mockRepo, mockRssReader, mockArticleService)

	req := &model.CreateFeedRequest{
		Name: "Example RSS Feed",
		URL:  "https://www.tagesschau.de/index~rss2.xml",
	}

	ctx := context.Background()

	// This will:
	// 1. Validate the name and URL format
	// 2. Validate that the URL actually returns a valid RSS feed
	// 3. Check for duplicate URLs
	// 4. Create the feed in the database
	feed, err := feedService.Create(ctx, req)
	if err != nil {
		// Handle validation errors or other issues
		return
	}

	// Feed was successfully created and validated
	_ = feed
}
