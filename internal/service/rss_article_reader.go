package service

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/lucasg04/fyrss/internal/model"
	"github.com/mmcdole/gofeed"
)

type RssArticleReader struct {
	articleService *ArticleService
}

func NewRssArticleReader(articleService *ArticleService) *RssArticleReader {
	return &RssArticleReader{articleService: articleService}
}

func (r *RssArticleReader) ReadArticleFeed(ctx context.Context, feedURL string) ([]*model.Article, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed URL %s: %w", feedURL, err)
	}
	if feed == nil || len(feed.Items) == 0 {
		return nil, fmt.Errorf("no elements found in feed URL %s", feedURL)
	}

	feedLength := len(feed.Items)
	articles := make([]*model.Article, feedLength)
	for i, item := range feed.Items {
		articles[i] = &model.Article{
			ID:          uuid.New(),
			Title:       item.Title,
			Description: item.Description,
			ContentHash: generateContentHash(item),
			Tags:        []string{},
			SourceUrl:   item.Link,
			PublishedAt: *item.PublishedParsed,
			SourceType:  "rss",
			Save:        false,
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
