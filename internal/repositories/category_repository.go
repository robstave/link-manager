package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robstave/link-manager/internal/models"
)

type CategoryWithCount struct {
	models.Category
	LinkCount int `json:"link_count"`
}

type CategoryRepository struct{ pool *pgxpool.Pool }

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) ProjectOwnerID(ctx context.Context, projectID string) (string, error) {
	var ownerID string
	err := r.pool.QueryRow(ctx, `SELECT owner_id FROM projects WHERE id = $1`, projectID).Scan(&ownerID)
	return ownerID, err
}

func (r *CategoryRepository) List(ctx context.Context, projectID string) ([]CategoryWithCount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			c.id, c.project_id, c.name, c.is_default, c.display_order, c.created_at, c.updated_at,
			COUNT(l.id) as link_count
		FROM categories c
		LEFT JOIN links l ON l.category_id = c.id
		WHERE c.project_id = $1
		GROUP BY c.id
		ORDER BY c.display_order, c.created_at
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []CategoryWithCount{}
	for rows.Next() {
		var cat CategoryWithCount
		if err := rows.Scan(&cat.ID, &cat.ProjectID, &cat.Name, &cat.IsDefault, &cat.DisplayOrder, &cat.CreatedAt, &cat.UpdatedAt, &cat.LinkCount); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *CategoryRepository) Create(ctx context.Context, projectID, name string) (models.Category, error) {
	var category models.Category
	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (project_id, name)
		VALUES ($1, $2)
		RETURNING id, project_id, name, is_default, display_order, created_at, updated_at
	`, projectID, name).Scan(
		&category.ID, &category.ProjectID, &category.Name, &category.IsDefault,
		&category.DisplayOrder, &category.CreatedAt, &category.UpdatedAt,
	)
	return category, err
}

func (r *CategoryRepository) CategoryMetadata(ctx context.Context, categoryID string) (bool, string, string, error) {
	var isDefault bool
	var projectID, ownerID string
	err := r.pool.QueryRow(ctx, `
		SELECT c.is_default, c.project_id, p.owner_id
		FROM categories c
		JOIN projects p ON p.id = c.project_id
		WHERE c.id = $1
	`, categoryID).Scan(&isDefault, &projectID, &ownerID)
	return isDefault, projectID, ownerID, err
}

func (r *CategoryRepository) DefaultCategoryID(ctx context.Context, projectID string) (string, error) {
	var defaultCategoryID string
	err := r.pool.QueryRow(ctx, `SELECT id FROM categories WHERE project_id = $1 AND is_default = true`, projectID).Scan(&defaultCategoryID)
	return defaultCategoryID, err
}

func (r *CategoryRepository) MoveLinks(ctx context.Context, toCategoryID, fromCategoryID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE links SET category_id = $1 WHERE category_id = $2`, toCategoryID, fromCategoryID)
	return err
}

func (r *CategoryRepository) Delete(ctx context.Context, categoryID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
	return err
}
