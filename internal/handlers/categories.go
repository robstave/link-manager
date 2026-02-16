package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/robstave/link-manager/internal/db"
	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/models"
)

type CategoryHandler struct {
	DB *db.DB
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type CategoryWithCount struct {
	models.Category
	LinkCount int `json:"link_count"`
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	projectID := r.PathValue("project_id")

	// Verify project ownership
	var ownerID string
	err := h.DB.Pool.QueryRow(r.Context(), `
		SELECT owner_id FROM projects WHERE id = $1
	`, projectID).Scan(&ownerID)

	if err != nil || ownerID != claims.UserID {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	rows, err := h.DB.Pool.Query(r.Context(), `
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
		http.Error(w, "failed to fetch categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	categories := []CategoryWithCount{}
	for rows.Next() {
		var cat CategoryWithCount
		err := rows.Scan(
			&cat.ID, &cat.ProjectID, &cat.Name, &cat.IsDefault, &cat.DisplayOrder,
			&cat.CreatedAt, &cat.UpdatedAt, &cat.LinkCount,
		)
		if err != nil {
			http.Error(w, "failed to scan category", http.StatusInternalServerError)
			return
		}
		categories = append(categories, cat)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	projectID := r.PathValue("project_id")

	// Verify project ownership
	var ownerID string
	err := h.DB.Pool.QueryRow(r.Context(), `
		SELECT owner_id FROM projects WHERE id = $1
	`, projectID).Scan(&ownerID)

	if err != nil || ownerID != claims.UserID {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var category models.Category
	err = h.DB.Pool.QueryRow(r.Context(), `
		INSERT INTO categories (project_id, name)
		VALUES ($1, $2)
		RETURNING id, project_id, name, is_default, display_order, created_at, updated_at
	`, projectID, req.Name).Scan(
		&category.ID, &category.ProjectID, &category.Name, &category.IsDefault,
		&category.DisplayOrder, &category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "failed to create category", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	categoryID := r.PathValue("id")

	// Check if it's the default category and verify ownership
	var isDefault bool
	var projectID, ownerID string
	err := h.DB.Pool.QueryRow(r.Context(), `
		SELECT c.is_default, c.project_id, p.owner_id
		FROM categories c
		JOIN projects p ON p.id = c.project_id
		WHERE c.id = $1
	`, categoryID).Scan(&isDefault, &projectID, &ownerID)

	if err != nil || ownerID != claims.UserID {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}

	if isDefault {
		http.Error(w, "cannot delete default category", http.StatusBadRequest)
		return
	}

	// Move links to default category before deleting
	var defaultCategoryID string
	err = h.DB.Pool.QueryRow(r.Context(), `
		SELECT id FROM categories WHERE project_id = $1 AND is_default = true
	`, projectID).Scan(&defaultCategoryID)

	if err != nil {
		http.Error(w, "no default category found", http.StatusInternalServerError)
		return
	}

	// Move links
	_, err = h.DB.Pool.Exec(r.Context(), `
		UPDATE links SET category_id = $1 WHERE category_id = $2
	`, defaultCategoryID, categoryID)

	if err != nil {
		http.Error(w, "failed to move links", http.StatusInternalServerError)
		return
	}

	// Delete category
	_, err = h.DB.Pool.Exec(r.Context(), `
		DELETE FROM categories WHERE id = $1
	`, categoryID)

	if err != nil {
		http.Error(w, "failed to delete category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
