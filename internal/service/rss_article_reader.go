package service

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/mmcdole/gofeed"
)

type RssArticleReader struct {
	articleService *ArticleService
}

func NewRssArticleReader(articleService *ArticleService) *RssArticleReader {
	return &RssArticleReader{articleService: articleService}
}

func (r *RssArticleReader) ReadFeed(ctx context.Context, feed *model.Feed) ([]*model.Article, error) {
	fp := gofeed.NewParser()
	rssFeed, err := fp.ParseURLWithContext(feed.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed URL %s: %w", feed.URL, err)
	}
	if rssFeed == nil || len(rssFeed.Items) == 0 {
		return nil, fmt.Errorf("no elements found in feed URL %s", feed.URL)
	}

	feedLength := len(rssFeed.Items)
	articles := make([]*model.Article, feedLength)
	for i, item := range rssFeed.Items {
		articles[i] = &model.Article{
			ID:          uuid.New(),
			Title:       item.Title,
			Description: item.Description,
			ContentHash: generateContentHash(item),
			SourceUrl:   item.Link,
			PublishedAt: *item.PublishedParsed,
			SourceType:  "rss",
			Save:        false,
			FeedID:      &feed.ID, // Associate with feed if provided
		}
	}

	return articles, nil
}

func generateContentHash(item *gofeed.Item) string {
	combinedContent := item.Title + item.Description + item.Link
	hash := sha256.New()
	hash.Write([]byte(combinedContent))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
