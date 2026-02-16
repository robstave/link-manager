package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/robstave/link-manager/internal/db"
	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/models"
)

type LinkHandler struct {
	DB *db.DB
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

type UpdateLinkRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	UserNotes   *string  `json:"user_notes"`
	Stars       *int     `json:"stars"`
	Cart        *bool    `json:"cart"`
	Tags        []string `json:"tags"`
}

type LinkResponse struct {
	models.Link
	Project  *ProjectInfo  `json:"project,omitempty"`
	Category *CategoryInfo `json:"category,omitempty"`
}

type ProjectInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CategoryInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LinksListResponse struct {
	Links  []LinkResponse `json:"links"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func (h *LinkHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())

	// Parse query parameters
	projectID := r.URL.Query().Get("project_id")
	categoryID := r.URL.Query().Get("category_id")
	tag := r.URL.Query().Get("tag")
	cart := r.URL.Query().Get("cart")
	search := r.URL.Query().Get("q")
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "stars"
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Build query
	query := `
		SELECT 
			l.id, l.owner_id, l.project_id, l.category_id, l.url, l.title, l.description,
			l.icon_url, l.user_notes, l.generated_notes, l.generated_notes_size,
			l.stars, l.click_count, l.last_clicked_at, l.cart, l.created_at, l.updated_at,
			p.name as project_name, c.name as category_name,
			ARRAY_AGG(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
		FROM links l
		LEFT JOIN projects p ON p.id = l.project_id
		LEFT JOIN categories c ON c.id = l.category_id
		LEFT JOIN link_tags lt ON lt.link_id = l.id
		LEFT JOIN tags t ON t.id = lt.tag_id
		WHERE l.owner_id = $1
	`

	args := []interface{}{claims.UserID}
	argCount := 1

	if projectID != "" {
		argCount++
		query += ` AND l.project_id = $` + strconv.Itoa(argCount)
		args = append(args, projectID)
	}

	if categoryID != "" {
		argCount++
		query += ` AND l.category_id = $` + strconv.Itoa(argCount)
		args = append(args, categoryID)
	}

	if cart != "" {
		argCount++
		query += ` AND l.cart = $` + strconv.Itoa(argCount)
		args = append(args, cart == "true")
	}

	if tag != "" {
		argCount++
		query += ` AND EXISTS (
			SELECT 1 FROM link_tags lt2 
			JOIN tags t2 ON t2.id = lt2.tag_id 
			WHERE lt2.link_id = l.id AND t2.name = $` + strconv.Itoa(argCount) + `
		)`
		args = append(args, tag)
	}

	if search != "" {
		argCount++
		query += ` AND l.fts @@ plainto_tsquery('english', $` + strconv.Itoa(argCount) + `)`
		args = append(args, search)
	}

	query += ` GROUP BY l.id, p.name, c.name`

	// Add sorting
	switch sortBy {
	case "clicks":
		query += ` ORDER BY l.click_count DESC, l.created_at DESC`
	case "recent":
		query += ` ORDER BY l.last_clicked_at DESC NULLS LAST, l.created_at DESC`
	case "created":
		query += ` ORDER BY l.created_at DESC`
	default: // stars
		query += ` ORDER BY l.stars DESC, l.created_at DESC`
	}

	// Add pagination
	argCount++
	query += ` LIMIT $` + strconv.Itoa(argCount)
	args = append(args, limit)

	argCount++
	query += ` OFFSET $` + strconv.Itoa(argCount)
	args = append(args, offset)

	rows, err := h.DB.Pool.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "failed to fetch links: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	links := []LinkResponse{}
	for rows.Next() {
		var link models.Link
		var projectName, categoryName string
		var tags []string

		err := rows.Scan(
			&link.ID, &link.OwnerID, &link.ProjectID, &link.CategoryID, &link.URL,
			&link.Title, &link.Description, &link.IconURL, &link.UserNotes,
			&link.GeneratedNotes, &link.GeneratedNotesSize, &link.Stars,
			&link.ClickCount, &link.LastClickedAt, &link.Cart, &link.CreatedAt, &link.UpdatedAt,
			&projectName, &categoryName, &tags,
		)
		if err != nil {
			http.Error(w, "failed to scan link", http.StatusInternalServerError)
			return
		}

		link.Tags = tags

		resp := LinkResponse{
			Link: link,
			Project: &ProjectInfo{
				ID:   link.ProjectID,
				Name: projectName,
			},
			Category: &CategoryInfo{
				ID:   link.CategoryID,
				Name: categoryName,
			},
		}

		links = append(links, resp)
	}

	// Get total count (simplified - could be optimized)
	var total int
	err = h.DB.Pool.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM links WHERE owner_id = $1
	`, claims.UserID).Scan(&total)
	if err != nil {
		total = len(links)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LinksListResponse{
		Links:  links,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	// Get default project/category if not provided
	projectID := req.ProjectID
	categoryID := req.CategoryID

	if projectID == nil {
		var defaultProjectID string
		err := h.DB.Pool.QueryRow(r.Context(), `
			SELECT id FROM projects WHERE owner_id = $1 AND is_default = true
		`, claims.UserID).Scan(&defaultProjectID)
		if err != nil {
			http.Error(w, "no default project found", http.StatusInternalServerError)
			return
		}
		projectID = &defaultProjectID
	}

	if categoryID == nil {
		var defaultCategoryID string
		err := h.DB.Pool.QueryRow(r.Context(), `
			SELECT id FROM categories WHERE project_id = $1 AND is_default = true
		`, *projectID).Scan(&defaultCategoryID)
		if err != nil {
			http.Error(w, "no default category found", http.StatusInternalServerError)
			return
		}
		categoryID = &defaultCategoryID
	}

	// Create link
	var link models.Link
	err := h.DB.Pool.QueryRow(r.Context(), `
		INSERT INTO links (owner_id, project_id, category_id, url, title, description, stars)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, owner_id, project_id, category_id, url, title, description, icon_url,
			user_notes, generated_notes, generated_notes_size, stars, click_count, 
			last_clicked_at, cart, created_at, updated_at
	`, claims.UserID, *projectID, *categoryID, req.URL, req.Title, req.Description, req.Stars).Scan(
		&link.ID, &link.OwnerID, &link.ProjectID, &link.CategoryID, &link.URL,
		&link.Title, &link.Description, &link.IconURL, &link.UserNotes,
		&link.GeneratedNotes, &link.GeneratedNotesSize, &link.Stars,
		&link.ClickCount, &link.LastClickedAt, &link.Cart, &link.CreatedAt, &link.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "failed to create link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add tags if provided
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			// Get or create tag
			var tagID string
			err := h.DB.Pool.QueryRow(r.Context(), `
				INSERT INTO tags (owner_id, name)
				VALUES ($1, $2)
				ON CONFLICT (owner_id, name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id
			`, claims.UserID, tagName).Scan(&tagID)

			if err != nil {
				continue
			}

			// Link tag to link
			_, _ = h.DB.Pool.Exec(r.Context(), `
				INSERT INTO link_tags (link_id, tag_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, link.ID, tagID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

func (h *LinkHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	linkID := r.PathValue("id")

	var link models.Link
	var projectName, categoryName string
	var tags []string

	err := h.DB.Pool.QueryRow(r.Context(), `
		SELECT 
			l.id, l.owner_id, l.project_id, l.category_id, l.url, l.title, l.description,
			l.icon_url, l.user_notes, l.generated_notes, l.generated_notes_size,
			l.stars, l.click_count, l.last_clicked_at, l.cart, l.created_at, l.updated_at,
			p.name as project_name, c.name as category_name,
			ARRAY_AGG(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
		FROM links l
		LEFT JOIN projects p ON p.id = l.project_id
		LEFT JOIN categories c ON c.id = l.category_id
		LEFT JOIN link_tags lt ON lt.link_id = l.id
		LEFT JOIN tags t ON t.id = lt.tag_id
		WHERE l.id = $1 AND l.owner_id = $2
		GROUP BY l.id, p.name, c.name
	`, linkID, claims.UserID).Scan(
		&link.ID, &link.OwnerID, &link.ProjectID, &link.CategoryID, &link.URL,
		&link.Title, &link.Description, &link.IconURL, &link.UserNotes,
		&link.GeneratedNotes, &link.GeneratedNotesSize, &link.Stars,
		&link.ClickCount, &link.LastClickedAt, &link.Cart, &link.CreatedAt, &link.UpdatedAt,
		&projectName, &categoryName, &tags,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to fetch link", http.StatusInternalServerError)
		return
	}

	link.Tags = tags

	resp := LinkResponse{
		Link: link,
		Project: &ProjectInfo{
			ID:   link.ProjectID,
			Name: projectName,
		},
		Category: &CategoryInfo{
			ID:   link.CategoryID,
			Name: categoryName,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *LinkHandler) Click(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	linkID := r.PathValue("id")

	var url string
	err := h.DB.Pool.QueryRow(r.Context(), `
		UPDATE links 
		SET click_count = click_count + 1, last_clicked_at = NOW()
		WHERE id = $1 AND owner_id = $2
		RETURNING url
	`, linkID, claims.UserID).Scan(&url)

	if err == pgx.ErrNoRows {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to record click", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"redirect_url": url,
	})
}

func (h *LinkHandler) UpdateStars(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	linkID := r.PathValue("id")

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

	_, err := h.DB.Pool.Exec(r.Context(), `
		UPDATE links SET stars = $1, updated_at = NOW()
		WHERE id = $2 AND owner_id = $3
	`, req.Stars, linkID, claims.UserID)

	if err != nil {
		http.Error(w, "failed to update stars", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *LinkHandler) ToggleCart(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	linkID := r.PathValue("id")

	var req struct {
		Cart bool `json:"cart"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Pool.Exec(r.Context(), `
		UPDATE links SET cart = $1, updated_at = NOW()
		WHERE id = $2 AND owner_id = $3
	`, req.Cart, linkID, claims.UserID)

	if err != nil {
		http.Error(w, "failed to update cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *LinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())
	linkID := r.PathValue("id")

	_, err := h.DB.Pool.Exec(r.Context(), `
		DELETE FROM links WHERE id = $1 AND owner_id = $2
	`, linkID, claims.UserID)

	if err != nil {
		http.Error(w, "failed to delete link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (h *LinkHandler) Export(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())

	// For export, we reuse the List query but without pagination limit
	query := `
		SELECT 
			l.id, l.owner_id, l.project_id, l.category_id, l.url, l.title, l.description,
			l.icon_url, l.user_notes, l.generated_notes, l.generated_notes_size,
			l.stars, l.click_count, l.last_clicked_at, l.cart, l.created_at, l.updated_at,
			p.name as project_name, c.name as category_name,
			ARRAY_AGG(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
		FROM links l
		LEFT JOIN projects p ON p.id = l.project_id
		LEFT JOIN categories c ON c.id = l.category_id
		LEFT JOIN link_tags lt ON lt.link_id = l.id
		LEFT JOIN tags t ON t.id = lt.tag_id
		WHERE l.owner_id = $1
		GROUP BY l.id, p.name, c.name
		ORDER BY l.created_at DESC
	`

	rows, err := h.DB.Pool.Query(r.Context(), query, claims.UserID)
	if err != nil {
		http.Error(w, "failed to fetch links for export: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	links := []LinkResponse{}
	for rows.Next() {
		var link models.Link
		var projectName, categoryName string
		var tags []string

		err := rows.Scan(
			&link.ID, &link.OwnerID, &link.ProjectID, &link.CategoryID, &link.URL,
			&link.Title, &link.Description, &link.IconURL, &link.UserNotes,
			&link.GeneratedNotes, &link.GeneratedNotesSize, &link.Stars,
			&link.ClickCount, &link.LastClickedAt, &link.Cart, &link.CreatedAt, &link.UpdatedAt,
			&projectName, &categoryName, &tags,
		)
		if err != nil {
			http.Error(w, "failed to scan link", http.StatusInternalServerError)
			return
		}

		link.Tags = tags
		resp := LinkResponse{
			Link: link,
			Project: &ProjectInfo{
				ID:   link.ProjectID,
				Name: projectName,
			},
			Category: &CategoryInfo{
				ID:   link.CategoryID,
				Name: categoryName,
			},
		}
		links = append(links, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=links_export.json")
	json.NewEncoder(w).Encode(links)
}
