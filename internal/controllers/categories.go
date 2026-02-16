package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/services"
)

type CategoryController struct{ service *services.CategoryService }

func NewCategoryController(service *services.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

func (c *CategoryController) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	items, err := c.service.List(r.Context(), claims.UserID, r.PathValue("project_id"))
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (c *CategoryController) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	item, err := c.service.Create(r.Context(), claims.UserID, r.PathValue("project_id"), req.Name)
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (c *CategoryController) Delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	err := c.service.Delete(r.Context(), claims.UserID, r.PathValue("id"))
	if errors.Is(err, services.ErrDefaultCategoryDelete) {
		http.Error(w, "cannot delete default category", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
