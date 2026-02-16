package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/services"
)

type ProjectController struct{ service *services.ProjectService }

func NewProjectController(service *services.ProjectService) *ProjectController {
	return &ProjectController{service: service}
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *ProjectController) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	projects, err := c.service.List(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "failed to fetch projects", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (c *ProjectController) Create(w http.ResponseWriter, r *http.Request) {
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
	project, err := c.service.Create(r.Context(), claims.UserID, req.Name, req.Description)
	if err != nil {
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (c *ProjectController) Delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	err := c.service.Delete(r.Context(), claims.UserID, r.PathValue("id"))
	if errors.Is(err, services.ErrDefaultProjectDelete) {
		http.Error(w, "cannot delete default project", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
