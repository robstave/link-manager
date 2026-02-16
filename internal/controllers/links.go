package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/repositories"
	"github.com/robstave/link-manager/internal/services"
)

type LinkController struct{ service *services.LinkService }

func NewLinkController(service *services.LinkService) *LinkController {
	return &LinkController{service: service}
}

type CreateLinkRequest struct {
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ProjectID   *string  `json:"project_id"`
	CategoryID  *string  `json:"category_id"`
	Tags        []string `json:"tags"`
	Stars       int      `json:"stars"`
}

type ProjectInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type CategoryInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LinkResponse struct {
	ID                 string        `json:"id"`
	OwnerID            string        `json:"owner_id"`
	ProjectID          string        `json:"project_id"`
	CategoryID         string        `json:"category_id"`
	URL                string        `json:"url"`
	Title              string        `json:"title"`
	Description        string        `json:"description"`
	IconURL            string        `json:"icon_url"`
	UserNotes          string        `json:"user_notes"`
	GeneratedNotes     string        `json:"generated_notes"`
	GeneratedNotesSize string        `json:"generated_notes_size"`
	Stars              int           `json:"stars"`
	ClickCount         int           `json:"click_count"`
	LastClickedAt      any           `json:"last_clicked_at,omitempty"`
	Cart               bool          `json:"cart"`
	CreatedAt          any           `json:"created_at"`
	UpdatedAt          any           `json:"updated_at"`
	Tags               []string      `json:"tags,omitempty"`
	Project            *ProjectInfo  `json:"project,omitempty"`
	Category           *CategoryInfo `json:"category,omitempty"`
}

type LinksListResponse struct {
	Links                []LinkResponse `json:"links"`
	Total, Limit, Offset int
}

func toResponse(item repositories.LinkWithMeta) LinkResponse {
	return LinkResponse{ID: item.ID, OwnerID: item.OwnerID, ProjectID: item.ProjectID, CategoryID: item.CategoryID, URL: item.URL, Title: item.Title, Description: item.Description, IconURL: item.IconURL, UserNotes: item.UserNotes, GeneratedNotes: item.GeneratedNotes, GeneratedNotesSize: item.GeneratedNotesSize, Stars: item.Stars, ClickCount: item.ClickCount, LastClickedAt: item.LastClickedAt, Cart: item.Cart, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, Tags: item.Tags, Project: &ProjectInfo{ID: item.ProjectID, Name: item.ProjectName}, Category: &CategoryInfo{ID: item.CategoryID, Name: item.CategoryName}}
}

func (c *LinkController) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "stars"
	}
	limit, offset := 50, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	items, total, err := c.service.List(r.Context(), claims.UserID, repositories.LinkFilters{ProjectID: r.URL.Query().Get("project_id"), CategoryID: r.URL.Query().Get("category_id"), Tag: r.URL.Query().Get("tag"), Cart: r.URL.Query().Get("cart"), Search: r.URL.Query().Get("q"), SortBy: sortBy, Limit: limit, Offset: offset})
	if err != nil {
		http.Error(w, "failed to fetch links: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]LinkResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, toResponse(it))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LinksListResponse{Links: resp, Total: total, Limit: limit, Offset: offset})
}

func (c *LinkController) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	projectID, categoryID := "", ""
	if req.ProjectID != nil {
		projectID = *req.ProjectID
	}
	if req.CategoryID != nil {
		categoryID = *req.CategoryID
	}
	link, err := c.service.Create(r.Context(), claims.UserID, services.CreateLinkInput{URL: req.URL, Title: req.Title, Description: req.Description, ProjectID: projectID, CategoryID: categoryID, Tags: req.Tags, Stars: req.Stars})
	if err != nil {
		http.Error(w, "failed to create link: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

func (c *LinkController) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	item, err := c.service.Get(r.Context(), r.PathValue("id"), claims.UserID)
	if services.IsNotFound(err) {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to fetch link", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toResponse(item))
}

func (c *LinkController) Click(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	url, err := c.service.Click(r.Context(), r.PathValue("id"), claims.UserID)
	if services.IsNotFound(err) {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to record click", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"redirect_url": url})
}

func (c *LinkController) UpdateStars(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	var req struct {
		Stars int `json:"stars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Stars < 0 || req.Stars > 10 {
		http.Error(w, "stars must be between 0 and 10", http.StatusBadRequest)
		return
	}
	if err := c.service.UpdateStars(r.Context(), r.PathValue("id"), claims.UserID, req.Stars); err != nil {
		http.Error(w, "failed to update stars", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *LinkController) ToggleCart(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	var req struct {
		Cart bool `json:"cart"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := c.service.ToggleCart(r.Context(), r.PathValue("id"), claims.UserID, req.Cart); err != nil {
		http.Error(w, "failed to update cart", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *LinkController) Delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	if err := c.service.Delete(r.Context(), r.PathValue("id"), claims.UserID); err != nil {
		http.Error(w, "failed to delete link", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *LinkController) Export(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	items, err := c.service.Export(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "failed to fetch links for export: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]LinkResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, toResponse(it))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=links_export.json")
	json.NewEncoder(w).Encode(resp)
}
