package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/robstave/link-manager/internal/models"
	"github.com/robstave/link-manager/internal/repositories"
)

type LinkService struct {
	repo    *repositories.LinkRepository
	metaSvc *MetadataService
}

func NewLinkService(repo *repositories.LinkRepository, metaSvc *MetadataService) *LinkService {
	return &LinkService{repo: repo, metaSvc: metaSvc}
}

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
	normURL := normalizeURL(req.URL)
	title := req.Title
	iconURL := req.IconURL

	// Auto-fetch title and icon when title is empty
	if title == "" && normURL != "" && s.metaSvc != nil {
		slog.Info("link-create: title empty, auto-fetching metadata", "url", normURL)
		meta, err := s.metaSvc.FetchPageMeta(normURL)
		if err != nil {
			slog.Error("link-create: auto-fetch failed", "url", normURL, "error", err)
		} else {
			slog.Info("link-create: auto-fetch succeeded", "url", normURL, "title", meta.Title, "description", meta.Description, "iconURL", meta.IconURL)
			if meta.Title != "" {
				title = meta.Title
			}
			if meta.Description != "" && req.Description == "" {
				req.Description = meta.Description
			}
			if meta.IconURL != "" {
				iconURL = meta.IconURL
			}
		}
	} else if title != "" {
		slog.Info("link-create: title provided, skipping auto-fetch", "title", title)
	}

	return s.repo.Create(ctx, ownerID, projectID, categoryID, normURL, title, req.Description, req.UserNotes, iconURL, req.Stars, req.Tags)
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
	normURL := normalizeURL(req.URL)
	title := req.Title
	iconURL := req.IconURL

	// Auto-fetch title and icon when title is empty
	if title == "" && normURL != "" && s.metaSvc != nil {
		slog.Info("link-update: title empty, auto-fetching metadata", "url", normURL, "linkID", linkID)
		meta, err := s.metaSvc.FetchPageMeta(normURL)
		if err != nil {
			slog.Error("link-update: auto-fetch failed", "url", normURL, "linkID", linkID, "error", err)
		} else {
			slog.Info("link-update: auto-fetch succeeded", "url", normURL, "linkID", linkID, "title", meta.Title, "description", meta.Description, "iconURL", meta.IconURL)
			if meta.Title != "" {
				title = meta.Title
			}
			if meta.Description != "" && req.Description == "" {
				req.Description = meta.Description
			}
			if meta.IconURL != "" {
				iconURL = meta.IconURL
			}
		}
	} else if title != "" {
		slog.Info("link-update: title provided, skipping auto-fetch", "title", title, "linkID", linkID)
	}

	return s.repo.Update(ctx, ownerID, linkID, projectID, categoryID, normURL, title, req.Description, req.UserNotes, iconURL, req.Stars, req.Tags)
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
	IconURL                            string
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
