package services

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/robstave/link-manager/internal/models"
	"github.com/robstave/link-manager/internal/repositories"
)

type LinkService struct{ repo *repositories.LinkRepository }

func NewLinkService(repo *repositories.LinkRepository) *LinkService { return &LinkService{repo: repo} }

func (s *LinkService) List(ctx context.Context, ownerID string, f repositories.LinkFilters) ([]repositories.LinkWithMeta, int, error) {
	return s.repo.List(ctx, ownerID, f)
}

func (s *LinkService) Create(ctx context.Context, ownerID string, req CreateLinkInput) (models.Link, error) {
	projectID := req.ProjectID
	categoryID := req.CategoryID
	if projectID == "" {
		id, err := s.repo.DefaultProjectID(ctx, ownerID)
		if err != nil {
			return models.Link{}, err
		}
		projectID = id
	}
	if categoryID == "" {
		id, err := s.repo.DefaultCategoryID(ctx, projectID)
		if err != nil {
			return models.Link{}, err
		}
		categoryID = id
	}
	return s.repo.Create(ctx, ownerID, projectID, categoryID, normalizeURL(req.URL), req.Title, req.Description, req.UserNotes, req.Stars, req.Tags)
}

func (s *LinkService) Get(ctx context.Context, linkID, ownerID string) (repositories.LinkWithMeta, error) {
	return s.repo.Get(ctx, linkID, ownerID)
}
func (s *LinkService) Click(ctx context.Context, linkID, ownerID string) (string, error) {
	return s.repo.Click(ctx, linkID, ownerID)
}

func (s *LinkService) Update(ctx context.Context, ownerID, linkID string, req CreateLinkInput) error {
	projectID := req.ProjectID
	categoryID := req.CategoryID
	if projectID == "" {
		id, err := s.repo.DefaultProjectID(ctx, ownerID)
		if err != nil {
			return err
		}
		projectID = id
	}
	if categoryID == "" {
		id, err := s.repo.DefaultCategoryID(ctx, projectID)
		if err != nil {
			return err
		}
		categoryID = id
	}
	return s.repo.Update(ctx, ownerID, linkID, projectID, categoryID, normalizeURL(req.URL), req.Title, req.Description, req.UserNotes, req.Stars, req.Tags)
}

func (s *LinkService) UpdateStars(ctx context.Context, linkID, ownerID string, stars int) error {
	return s.repo.UpdateStars(ctx, linkID, ownerID, stars)
}
func (s *LinkService) ToggleCart(ctx context.Context, linkID, ownerID string, cart bool) error {
	return s.repo.ToggleCart(ctx, linkID, ownerID, cart)
}
func (s *LinkService) Delete(ctx context.Context, linkID, ownerID string) error {
	return s.repo.Delete(ctx, linkID, ownerID)
}
func (s *LinkService) Export(ctx context.Context, ownerID string) ([]repositories.LinkWithMeta, error) {
	return s.repo.Export(ctx, ownerID)
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || err == pgx.ErrNoRows
}

type CreateLinkInput struct {
	URL, Title, Description, UserNotes string
	ProjectID, CategoryID              string
	Tags                               []string
	Stars                              int
}

func normalizeURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return value
	}
	if strings.Contains(value, "://") {
		return value
	}
	return "https://" + value
}
