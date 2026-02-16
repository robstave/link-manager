package services

import (
	"context"
	"errors"

	"github.com/robstave/link-manager/internal/models"
	"github.com/robstave/link-manager/internal/repositories"
)

var ErrDefaultCategoryDelete = errors.New("cannot delete default category")

type CategoryService struct {
	repo *repositories.CategoryRepository
}

func NewCategoryService(repo *repositories.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context, ownerID, projectID string) ([]repositories.CategoryWithCount, error) {
	pidOwner, err := s.repo.ProjectOwnerID(ctx, projectID)
	if err != nil || pidOwner != ownerID {
		return nil, errors.New("project not found")
	}
	return s.repo.List(ctx, projectID)
}

func (s *CategoryService) Create(ctx context.Context, ownerID, projectID, name string) (models.Category, error) {
	pidOwner, err := s.repo.ProjectOwnerID(ctx, projectID)
	if err != nil || pidOwner != ownerID {
		return models.Category{}, errors.New("project not found")
	}
	return s.repo.Create(ctx, projectID, name)
}

func (s *CategoryService) Delete(ctx context.Context, ownerID, categoryID string) error {
	isDefault, projectID, owner, err := s.repo.CategoryMetadata(ctx, categoryID)
	if err != nil || owner != ownerID {
		return errors.New("category not found")
	}
	if isDefault {
		return ErrDefaultCategoryDelete
	}
	defaultID, err := s.repo.DefaultCategoryID(ctx, projectID)
	if err != nil {
		return err
	}
	if err := s.repo.MoveLinks(ctx, defaultID, categoryID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, categoryID)
}
