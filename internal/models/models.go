package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email        *string   `json:"email,omitempty"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Project struct {
	ID           string    `json:"id"`
	OwnerID      string    `json:"owner_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	IsDefault    bool      `json:"is_default"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Category struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	Name         string    `json:"name"`
	IsDefault    bool      `json:"is_default"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Link struct {
	ID                  string     `json:"id"`
	OwnerID             string     `json:"owner_id"`
	ProjectID           string     `json:"project_id"`
	CategoryID          string     `json:"category_id"`
	URL                 string     `json:"url"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	IconURL             string     `json:"icon_url"`
	UserNotes           string     `json:"user_notes"`
	GeneratedNotes      string     `json:"generated_notes"`
	GeneratedNotesSize  string     `json:"generated_notes_size"`
	Stars               int        `json:"stars"`
	ClickCount          int        `json:"click_count"`
	LastClickedAt       *time.Time `json:"last_clicked_at,omitempty"`
	Cart                bool       `json:"cart"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	Tags                []string   `json:"tags,omitempty"`
	ProjectName         string     `json:"project_name,omitempty"`
	CategoryName        string     `json:"category_name,omitempty"`
}

type Tag struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	LinkCount int       `json:"link_count,omitempty"`
}
