package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lucasg04/nuntium/internal/model"
)

func TestSortFeedArticles(t *testing.T) {
	t.Run("no change with articles sorted by publishedAt with large time diffs", func(t *testing.T) {
		articles := []*model.MinimalFeedArticle{
			mockArticle(1, time.Now().Add(-2*time.Hour)),
			mockArticle(1, time.Now().Add(-8*time.Hour)),
			mockArticle(1, time.Now().Add(-15*time.Hour)),
			mockArticle(1, time.Now().Add(-17*time.Hour)),
			mockArticle(1, time.Now().Add(-27*time.Hour)),
			mockArticle(1, time.Now().Add(-30*time.Hour)),
		}
		sortedArticles := SortFeedArticles(articles)
		assertDateEquals(t, articles[0].PublishedAt, sortedArticles[0].PublishedAt)
		assertDateEquals(t, articles[1].PublishedAt, sortedArticles[1].PublishedAt)
		assertDateEquals(t, articles[2].PublishedAt, sortedArticles[2].PublishedAt)
		assertDateEquals(t, articles[3].PublishedAt, sortedArticles[3].PublishedAt)
		assertDateEquals(t, articles[4].PublishedAt, sortedArticles[4].PublishedAt)
		assertDateEquals(t, articles[5].PublishedAt, sortedArticles[5].PublishedAt)
	})

	t.Run("articles by priority and publishedAt", func(t *testing.T) {

		articles := []*model.MinimalFeedArticle{
			mockArticle(2, time.Now()),
			mockArticle(model.PriorityUnknown, time.Now()),
			mockArticle(5, time.Now()),
			mockArticle(model.PriorityUnknown, time.Now().Add(-2*time.Hour)),
			mockArticle(1, time.Now()),
			mockArticle(2, time.Now().Add(-2*time.Hour)),
		}

		sortedArticles := SortFeedArticles(articles)

		if len(sortedArticles) != 6 {
			t.Fatalf("Expected 6 articles, got %d", len(sortedArticles))
		}
		assertSameID(t, articles[0], sortedArticles[1])
		assertSameID(t, articles[1], sortedArticles[4])
		assertSameID(t, articles[2], sortedArticles[3])
		assertSameID(t, articles[3], sortedArticles[5])
		assertSameID(t, articles[4], sortedArticles[0])
		assertSameID(t, articles[5], sortedArticles[2])
	})
}

func mockArticle(priority int, publishedAt time.Time) *model.MinimalFeedArticle {
	return &model.MinimalFeedArticle{
		ID:          uuid.New(),
		Priority:    priority,
		PublishedAt: publishedAt,
	}
}

func assertSameID(t *testing.T, a, b *model.MinimalFeedArticle) {
	if a.ID != b.ID {
		t.Errorf("Expected article ID %v, got %v", a.ID, b.ID)
	}
}

func assertDateEquals(t *testing.T, a, b time.Time) {
	if !a.Equal(b) {
		t.Errorf("Expected date %v, got %v", a, b)
	}
}
