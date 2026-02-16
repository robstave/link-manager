package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/robstave/link-manager/internal/db"
	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/models"
)

type TagHandler struct {
	DB *db.DB
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserClaims(r.Context())

	rows, err := h.DB.Pool.Query(r.Context(), `
		SELECT 
			t.id, t.owner_id, t.name, t.color, t.created_at,
			COUNT(lt.link_id) as link_count
		FROM tags t
		LEFT JOIN link_tags lt ON lt.tag_id = t.id
		WHERE t.owner_id = $1
		GROUP BY t.id
		ORDER BY t.name
	`, claims.UserID)

	if err != nil {
		http.Error(w, "failed to fetch tags", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.OwnerID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.LinkCount)
		if err != nil {
			http.Error(w, "failed to scan tag", http.StatusInternalServerError)
			return
		}
		tags = append(tags, tag)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}
