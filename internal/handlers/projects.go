package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/robstave/link-manager/internal/db"
	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/models"
)

type ProjectHandler struct {
	DB *db.DB
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectWithCounts struct {
	models.Project
	CategoryCount int `json:"category_count"`
	LinkCount     int `json:"link_count"`
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())

	rows, err := h.DB.Pool.Query(r.Context(), `
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
	`, claims.UserID)

	if err != nil {
		http.Error(w, "failed to fetch projects", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	projects := []ProjectWithCounts{}
	for rows.Next() {
		var p ProjectWithCounts
		err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.IsDefault, &p.DisplayOrder,
			&p.CreatedAt, &p.UpdatedAt, &p.CategoryCount, &p.LinkCount,
		)
		if err != nil {
			http.Error(w, "failed to scan project", http.StatusInternalServerError)
			return
		}
		projects = append(projects, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var project models.Project
	err := h.DB.Pool.QueryRow(r.Context(), `
		INSERT INTO projects (owner_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, owner_id, name, description, is_default, display_order, created_at, updated_at
	`, claims.UserID, req.Name, req.Description).Scan(
		&project.ID, &project.OwnerID, &project.Name, &project.Description,
		&project.IsDefault, &project.DisplayOrder, &project.CreatedAt, &project.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	// Create default category for this project
	_, err = h.DB.Pool.Exec(r.Context(), `
		INSERT INTO categories (project_id, name, is_default)
		VALUES ($1, 'Unsorted', true)
	`, project.ID)

	if err != nil {
		http.Error(w, "failed to create default category", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	projectID := r.PathValue("id")

	// Check if it's the default project
	var isDefault bool
	err := h.DB.Pool.QueryRow(r.Context(), `
		SELECT is_default FROM projects WHERE id = $1 AND owner_id = $2
	`, projectID, claims.UserID).Scan(&isDefault)

	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	if isDefault {
		http.Error(w, "cannot delete default project", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Pool.Exec(r.Context(), `
		DELETE FROM projects WHERE id = $1 AND owner_id = $2
	`, projectID, claims.UserID)

	if err != nil {
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
