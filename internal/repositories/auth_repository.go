package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robstave/link-manager/internal/auth"
	"github.com/robstave/link-manager/internal/models"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository { return &AuthRepository{pool: pool} }

func (r *AuthRepository) Authenticate(ctx context.Context, username, password string) (*models.User, error) {
	return auth.Authenticate(ctx, r.pool, username, password)
}

func (r *AuthRepository) UserExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)
	return exists, err
}

func (r *AuthRepository) CreateUser(ctx context.Context, username, password, role string) (*models.User, error) {
	return auth.CreateUser(ctx, r.pool, username, password, role)
}

func (r *AuthRepository) DefaultProjectID(ctx context.Context, userID string) (string, error) {
	var projectID string
	err := r.pool.QueryRow(ctx, `SELECT id FROM projects WHERE owner_id = $1 AND is_default = true`, userID).Scan(&projectID)
	return projectID, err
}

func (r *AuthRepository) DefaultCategoryID(ctx context.Context, projectID string) (string, error) {
	var categoryID string
	err := r.pool.QueryRow(ctx, `SELECT id FROM categories WHERE project_id = $1 AND is_default = true`, projectID).Scan(&categoryID)
	return categoryID, err
}

func (r *AuthRepository) InsertLink(ctx context.Context, userID, projectID, categoryID, url, title string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO links (owner_id, project_id, category_id, url, title)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, projectID, categoryID, url, title)
	return err
}
