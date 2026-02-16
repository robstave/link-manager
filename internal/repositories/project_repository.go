package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robstave/link-manager/internal/models"
)

type ProjectWithCounts struct {
	models.Project
	CategoryCount int `json:"category_count"`
	LinkCount     int `json:"link_count"`
}

type ProjectRepository struct{ pool *pgxpool.Pool }

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

func (r *ProjectRepository) List(ctx context.Context, ownerID string) ([]ProjectWithCounts, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			p.id, p.owner_id, p.name, p.description, p.is_default, p.display_order, 
			p.created_at, p.updated_at,
			COUNT(DISTINCT c.id) as category_count,
			COUNT(DISTINCT l.id) as link_count
		FROM projects p
		LEFT JOIN categories c ON c.project_id = p.id
		LEFT JOIN links l ON l.project_id = p.id
		WHERE p.owner_id = $1
		GROUP BY p.id
		ORDER BY p.display_order, p.created_at
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []ProjectWithCounts{}
	for rows.Next() {
		var p ProjectWithCounts
		if err := rows.Scan(&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.IsDefault, &p.DisplayOrder, &p.CreatedAt, &p.UpdatedAt, &p.CategoryCount, &p.LinkCount); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *ProjectRepository) Create(ctx context.Context, ownerID, name, description string) (models.Project, error) {
	var project models.Project
	err := r.pool.QueryRow(ctx, `
		INSERT INTO projects (owner_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, owner_id, name, description, is_default, display_order, created_at, updated_at
	`, ownerID, name, description).Scan(
		&project.ID, &project.OwnerID, &project.Name, &project.Description,
		&project.IsDefault, &project.DisplayOrder, &project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		return models.Project{}, err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO categories (project_id, name, is_default) VALUES ($1, 'Unsorted', true)`, project.ID)
	return project, err
}

func (r *ProjectRepository) IsDefault(ctx context.Context, projectID, ownerID string) (bool, error) {
	var isDefault bool
	err := r.pool.QueryRow(ctx, `SELECT is_default FROM projects WHERE id = $1 AND owner_id = $2`, projectID, ownerID).Scan(&isDefault)
	return isDefault, err
}

func (r *ProjectRepository) Delete(ctx context.Context, projectID, ownerID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1 AND owner_id = $2`, projectID, ownerID)
	return err
}
