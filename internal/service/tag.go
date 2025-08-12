package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/lucasg04/fyrss-server/internal/repository"
)

var ErrInvalidTagWeight = errors.New("invalid weight")
var ErrInvalidTag = errors.New("invalid tag")

type TagService struct {
	repo repository.TagRepository
}

func NewTagService(repo *repository.TagRepository) *TagService {
	return &TagService{repo: *repo}
}

func (s *TagService) GetAll(ctx context.Context) ([]string, error) {
	return s.repo.GetAllTags(ctx)
}

func (s *TagService) GetTagsWithWeights(ctx context.Context) ([]*model.WeightedTag, error) {
	return s.repo.GetTagsWithWeights(ctx)
}

func (s *TagService) SetTagWeight(ctx context.Context, tag string, weight int) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ErrInvalidTag
	}
	if weight < 0 {
		return ErrInvalidTagWeight
	}

	weightedTag, _ := s.repo.GetWeightedTag(ctx, tag)
	if weightedTag == nil {
		if err := s.repo.CreateWeightedTag(ctx, tag, weight); err != nil {
			return fmt.Errorf("create weight for tag %q: %w", tag, err)
		}
		return nil
	}

	if err := s.repo.SetTagWeight(ctx, tag, weight); err != nil {
		return fmt.Errorf("set weight for tag %q: %w", tag, err)
	}
	return nil
}

func (s *TagService) RemoveTagWeight(ctx context.Context, tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ErrInvalidTag
	}
	if err := s.repo.RemoveWeight(ctx, tag); err != nil {
		return fmt.Errorf("remove weight for tag %q: %w", tag, err)
	}
	return nil
}
