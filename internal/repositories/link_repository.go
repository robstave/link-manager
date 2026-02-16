package repositories

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robstave/link-manager/internal/models"
)

type LinkRepository struct{ pool *pgxpool.Pool }

func NewLinkRepository(pool *pgxpool.Pool) *LinkRepository { return &LinkRepository{pool: pool} }

type LinkFilters struct {
	ProjectID  string
	CategoryID string
	Tag        string
	Cart       string
	Search     string
	SortBy     string
	Limit      int
	Offset     int
}

type LinkWithMeta struct {
	models.Link
	ProjectName  string
	CategoryName string
}

func (r *LinkRepository) List(ctx context.Context, ownerID string, f LinkFilters) ([]LinkWithMeta, int, error) {
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
	args := []interface{}{ownerID}
	argCount := 1

	if f.ProjectID != "" {
		argCount++
		query += ` AND l.project_id = $` + strconv.Itoa(argCount)
		args = append(args, f.ProjectID)
	}
	if f.CategoryID != "" {
		argCount++
		query += ` AND l.category_id = $` + strconv.Itoa(argCount)
		args = append(args, f.CategoryID)
	}
	if f.Cart != "" {
		argCount++
		query += ` AND l.cart = $` + strconv.Itoa(argCount)
		args = append(args, f.Cart == "true")
	}
	if f.Tag != "" {
		argCount++
		query += ` AND EXISTS (SELECT 1 FROM link_tags lt2 JOIN tags t2 ON t2.id = lt2.tag_id WHERE lt2.link_id = l.id AND t2.name = $` + strconv.Itoa(argCount) + `)`
		args = append(args, f.Tag)
	}
	if f.Search != "" {
		argCount++
		query += ` AND l.fts @@ plainto_tsquery('english', $` + strconv.Itoa(argCount) + `)`
		args = append(args, f.Search)
	}
	query += ` GROUP BY l.id, p.name, c.name`

	switch f.SortBy {
	case "clicks":
		query += ` ORDER BY l.click_count DESC, l.created_at DESC`
	case "recent":
		query += ` ORDER BY l.last_clicked_at DESC NULLS LAST, l.created_at DESC`
	case "created":
		query += ` ORDER BY l.created_at DESC`
	default:
		query += ` ORDER BY l.stars DESC, l.created_at DESC`
	}

	argCount++
	query += ` LIMIT $` + strconv.Itoa(argCount)
	args = append(args, f.Limit)
	argCount++
	query += ` OFFSET $` + strconv.Itoa(argCount)
	args = append(args, f.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	links := []LinkWithMeta{}
	for rows.Next() {
		var item LinkWithMeta
		var tags []string
		if err := rows.Scan(
			&item.ID, &item.OwnerID, &item.ProjectID, &item.CategoryID, &item.URL,
			&item.Title, &item.Description, &item.IconURL, &item.UserNotes,
			&item.GeneratedNotes, &item.GeneratedNotesSize, &item.Stars,
			&item.ClickCount, &item.LastClickedAt, &item.Cart, &item.CreatedAt, &item.UpdatedAt,
			&item.ProjectName, &item.CategoryName, &tags,
		); err != nil {
			return nil, 0, err
		}
		item.Tags = tags
		links = append(links, item)
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM links WHERE owner_id = $1`, ownerID).Scan(&total); err != nil {
		total = len(links)
	}
	return links, total, nil
}

func (r *LinkRepository) DefaultProjectID(ctx context.Context, ownerID string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT id FROM projects WHERE owner_id = $1 AND is_default = true`, ownerID).Scan(&id)
	return id, err
}

func (r *LinkRepository) DefaultCategoryID(ctx context.Context, projectID string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT id FROM categories WHERE project_id = $1 AND is_default = true`, projectID).Scan(&id)
	return id, err
}

func (r *LinkRepository) Create(ctx context.Context, ownerID, projectID, categoryID, url, title, description string, stars int, tags []string) (models.Link, error) {
	var link models.Link
	err := r.pool.QueryRow(ctx, `
		INSERT INTO links (owner_id, project_id, category_id, url, title, description, stars)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, owner_id, project_id, category_id, url, title, description, icon_url,
			user_notes, generated_notes, generated_notes_size, stars, click_count, 
			last_clicked_at, cart, created_at, updated_at
	`, ownerID, projectID, categoryID, url, title, description, stars).Scan(
		&link.ID, &link.OwnerID, &link.ProjectID, &link.CategoryID, &link.URL,
		&link.Title, &link.Description, &link.IconURL, &link.UserNotes,
		&link.GeneratedNotes, &link.GeneratedNotesSize, &link.Stars,
		&link.ClickCount, &link.LastClickedAt, &link.Cart, &link.CreatedAt, &link.UpdatedAt,
	)
	if err != nil {
		return models.Link{}, err
	}

	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		var tagID string
		err := r.pool.QueryRow(ctx, `
			INSERT INTO tags (owner_id, name)
			VALUES ($1, $2)
			ON CONFLICT (owner_id, name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, ownerID, tagName).Scan(&tagID)
		if err != nil {
			continue
		}
		_, _ = r.pool.Exec(ctx, `INSERT INTO link_tags (link_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, link.ID, tagID)
	}
	return link, nil
}

func (r *LinkRepository) Get(ctx context.Context, linkID, ownerID string) (LinkWithMeta, error) {
	var item LinkWithMeta
	var tags []string
	err := r.pool.QueryRow(ctx, `
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
	`, linkID, ownerID).Scan(
		&item.ID, &item.OwnerID, &item.ProjectID, &item.CategoryID, &item.URL,
		&item.Title, &item.Description, &item.IconURL, &item.UserNotes,
		&item.GeneratedNotes, &item.GeneratedNotesSize, &item.Stars,
		&item.ClickCount, &item.LastClickedAt, &item.Cart, &item.CreatedAt, &item.UpdatedAt,
		&item.ProjectName, &item.CategoryName, &tags,
	)
	if err != nil {
		return LinkWithMeta{}, err
	}
	item.Tags = tags
	return item, nil
}

func (r *LinkRepository) Click(ctx context.Context, linkID, ownerID string) (string, error) {
	var url string
	err := r.pool.QueryRow(ctx, `UPDATE links SET click_count = click_count + 1, last_clicked_at = NOW() WHERE id = $1 AND owner_id = $2 RETURNING url`, linkID, ownerID).Scan(&url)
	return url, err
}

func (r *LinkRepository) UpdateStars(ctx context.Context, linkID, ownerID string, stars int) error {
	_, err := r.pool.Exec(ctx, `UPDATE links SET stars = $1, updated_at = NOW() WHERE id = $2 AND owner_id = $3`, stars, linkID, ownerID)
	return err
}

func (r *LinkRepository) ToggleCart(ctx context.Context, linkID, ownerID string, cart bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE links SET cart = $1, updated_at = NOW() WHERE id = $2 AND owner_id = $3`, cart, linkID, ownerID)
	return err
}

func (r *LinkRepository) Delete(ctx context.Context, linkID, ownerID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM links WHERE id = $1 AND owner_id = $2`, linkID, ownerID)
	return err
}

func (r *LinkRepository) Export(ctx context.Context, ownerID string) ([]LinkWithMeta, error) {
	rows, err := r.pool.Query(ctx, `
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
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []LinkWithMeta{}
	for rows.Next() {
		var item LinkWithMeta
		var tags []string
		if err := rows.Scan(
			&item.ID, &item.OwnerID, &item.ProjectID, &item.CategoryID, &item.URL,
			&item.Title, &item.Description, &item.IconURL, &item.UserNotes,
			&item.GeneratedNotes, &item.GeneratedNotesSize, &item.Stars,
			&item.ClickCount, &item.LastClickedAt, &item.Cart, &item.CreatedAt, &item.UpdatedAt,
			&item.ProjectName, &item.CategoryName, &tags,
		); err != nil {
			return nil, err
		}
		item.Tags = tags
		links = append(links, item)
	}
	return links, nil
}

func IsNoRows(err error) bool { return err == pgx.ErrNoRows }
