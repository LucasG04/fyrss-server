package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/lucasg04/fyrss-server/internal/repository"
)

var ErrInvalidTag = errors.New("invalid tag")

type TagService struct {
	repo repository.TagRepository
}

func NewTagService(repo *repository.TagRepository) *TagService {
	return &TagService{repo: *repo}
}

func (s *TagService) GetAll(ctx context.Context) ([]*model.Tag, error) {
	return s.repo.GetAllTags(ctx)
}

func (s *TagService) GetTagByID(ctx context.Context, id uuid.UUID) (*model.Tag, error) {
	tag, err := s.repo.GetTagByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get tag by ID %q: %w", id, err)
	}
	return tag, nil
}

func (s *TagService) CreateTag(ctx context.Context, tagId uuid.UUID, name string) error {
	if _, err := s.repo.CreateTag(ctx, name); err != nil {
		return fmt.Errorf("create tag %q: %w", tagId, err)
	}
	return nil
}

func (s *TagService) UpdateTag(ctx context.Context, tag *model.Tag) error {
	exists, err := s.repo.ExistsByID(ctx, tag.ID)
	if err != nil {
		return fmt.Errorf("check if tag exists %q: %w", tag.ID, err)
	}
	if !exists {
		return ErrInvalidTag
	}
	if err := s.repo.UpdateTag(ctx, tag); err != nil {
		return fmt.Errorf("failed to update for tag %q: %w", tag.ID, err)
	}
	return nil
}

func (s *TagService) AssignTagsToArticle(ctx context.Context, articleID uuid.UUID, tags []string) error {
	tagIDs := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		tag, err := s.repo.GetByName(ctx, t)
		if err != nil {
			return fmt.Errorf("check if tag exists %q: %w", t, err)
		}
		if tag == nil {
			created, err := s.repo.CreateTag(ctx, t)
			if err != nil {
				return fmt.Errorf("create tag %q: %w", t, err)
			}
			tag = created
		}
		tagIDs[i] = tag.ID
	}

	if err := s.repo.AssignTagsToArticle(ctx, articleID, tagIDs); err != nil {
		return fmt.Errorf("assign tags to article %q: %w", articleID, err)
	}
	return nil
}

// GetTagsOfArticles returns a map[articleID][]*model.Tag for a slice of minimal articles.
func (s *TagService) GetTagsOfArticles(ctx context.Context, articles []*model.MinimalFeedArticle) (map[uuid.UUID][]*model.Tag, error) {
	ids := make([]uuid.UUID, 0, len(articles))
	for _, a := range articles {
		if a != nil {
			ids = append(ids, a.ID)
		}
	}
	rows, err := s.repo.GetTagsByArticleIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	res := make(map[uuid.UUID][]*model.Tag, len(ids))
	for _, r := range rows {
		res[r.ArticleID] = append(res[r.ArticleID], &model.Tag{ID: r.ID, Name: r.Name, Priority: r.Priority})
	}
	return res, nil
}
