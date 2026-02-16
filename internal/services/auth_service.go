package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/robstave/link-manager/internal/auth"
	"github.com/robstave/link-manager/internal/models"
	"github.com/robstave/link-manager/internal/repositories"
)

type AuthService struct {
	repo   *repositories.AuthRepository
	logger *slog.Logger
}

func NewAuthService(repo *repositories.AuthRepository, logger *slog.Logger) *AuthService {
	return &AuthService{repo: repo, logger: logger}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := s.repo.Authenticate(ctx, username, password)
	if err != nil {
		return "", "", err
	}
	token, expiresAt, err := auth.GenerateToken(user)
	if err != nil {
		return "", "", err
	}
	return token, expiresAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (s *AuthService) EnsureAdminUser(ctx context.Context) error {
	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminUsername == "" {
		adminUsername = "admin"
	}
	if adminPassword == "" {
		adminPassword = "admin"
	}

	exists, err := s.repo.UserExists(ctx, adminUsername)
	if err != nil {
		return fmt.Errorf("failed to check for admin user: %w", err)
	}
	if exists {
		s.logger.Info("admin user already exists", "username", adminUsername)
		return nil
	}

	s.logger.Info("creating admin user", "username", adminUsername)
	user, err := s.repo.CreateUser(ctx, adminUsername, adminPassword, "admin")
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	if err := s.seedSampleLinks(ctx, user); err != nil {
		s.logger.Warn("failed to seed sample links", "error", err)
	}
	return nil
}

func (s *AuthService) seedSampleLinks(ctx context.Context, user *models.User) error {
	projectID, err := s.repo.DefaultProjectID(ctx, user.ID)
	if err != nil {
		return err
	}
	categoryID, err := s.repo.DefaultCategoryID(ctx, projectID)
	if err != nil {
		return err
	}

	sampleLinks := []struct{ url, title string }{{"https://www.google.com", "Google"}, {"https://www.github.com", "GitHub"}, {"https://www.msn.com", "MSN"}}
	for _, l := range sampleLinks {
		if err := s.repo.InsertLink(ctx, user.ID, projectID, categoryID, l.url, l.title); err != nil {
			return err
		}
	}
	s.logger.Info("seeded sample links", "count", len(sampleLinks))
	return nil
}
