package services

import (
	"context"

	"github.com/robstave/link-manager/internal/models"
	"github.com/robstave/link-manager/internal/repositories"
)

type TagService struct{ repo *repositories.TagRepository }

func NewTagService(repo *repositories.TagRepository) *TagService { return &TagService{repo: repo} }

func (s *TagService) List(ctx context.Context, ownerID string) ([]models.Tag, error) {
	return s.repo.List(ctx, ownerID)
}
