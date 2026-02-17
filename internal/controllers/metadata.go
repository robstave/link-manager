package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/robstave/link-manager/internal/services"
)

type MetadataController struct {
	service *services.MetadataService
}

func NewMetadataController(service *services.MetadataService) *MetadataController {
	return &MetadataController{service: service}
}

func (c *MetadataController) FetchTitle(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	title, err := c.service.FetchTitle(url)
	if err != nil {
		http.Error(w, "failed to fetch title: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"title": title})
}
