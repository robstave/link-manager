package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robstave/link-manager/internal/models"
)

type TagRepository struct{ pool *pgxpool.Pool }

func NewTagRepository(pool *pgxpool.Pool) *TagRepository { return &TagRepository{pool: pool} }

func (r *TagRepository) List(ctx context.Context, ownerID string) ([]models.Tag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			t.id, t.owner_id, t.name, t.color, t.created_at,
			COUNT(lt.link_id) as link_count
		FROM tags t
		LEFT JOIN link_tags lt ON lt.tag_id = t.id
		WHERE t.owner_id = $1
		GROUP BY t.id
		ORDER BY t.name
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.OwnerID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.LinkCount); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}
