package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/services"
)

type TagController struct{ service *services.TagService }

func NewTagController(service *services.TagService) *TagController {
	return &TagController{service: service}
}

func (c *TagController) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	tags, err := c.service.List(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "failed to fetch tags", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}
