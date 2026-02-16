package services

import (
	"context"
	"errors"

	"github.com/robstave/link-manager/internal/models"
	"github.com/robstave/link-manager/internal/repositories"
)

var ErrDefaultProjectDelete = errors.New("cannot delete default project")

type ProjectService struct {
	repo *repositories.ProjectRepository
}

func NewProjectService(repo *repositories.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) List(ctx context.Context, ownerID string) ([]repositories.ProjectWithCounts, error) {
	return s.repo.List(ctx, ownerID)
}
func (s *ProjectService) Create(ctx context.Context, ownerID, name, description string) (models.Project, error) {
	return s.repo.Create(ctx, ownerID, name, description)
}
func (s *ProjectService) Delete(ctx context.Context, ownerID, projectID string) error {
	isDefault, err := s.repo.IsDefault(ctx, projectID, ownerID)
	if err != nil {
		return err
	}
	if isDefault {
		return ErrDefaultProjectDelete
	}
	return s.repo.Delete(ctx, projectID, ownerID)
}
