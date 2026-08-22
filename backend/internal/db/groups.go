package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devkrishna/link-groupy/backend/internal/models"
)

// ListGroupsWithCounts returns all groups with their link counts.
func ListGroupsWithCounts(ctx context.Context, pool *pgxpool.Pool) ([]*models.GroupWithCount, error) {
	rows, err := pool.Query(ctx,
		`SELECT g.id, g.name, g.color, g.sort_order, g.created_at,
		        COUNT(lg.link_id) AS link_count
		 FROM groups g
		 LEFT JOIN link_groups lg ON lg.group_id = g.id
		 GROUP BY g.id, g.name, g.color, g.sort_order, g.created_at
		 ORDER BY g.sort_order ASC, g.name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("ListGroupsWithCounts: %w", err)
	}
	defer rows.Close()

	groups, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*models.GroupWithCount, error) {
		g := &models.GroupWithCount{}
		err := row.Scan(
			&g.ID, &g.Name, &g.Color, &g.SortOrder, &g.CreatedAt,
			&g.LinkCount,
		)
		return g, err
	})
	if err != nil {
		return nil, fmt.Errorf("ListGroupsWithCounts scan: %w", err)
	}
	return groups, nil
}

// CreateGroup creates a new group.
func CreateGroup(ctx context.Context, pool *pgxpool.Pool, name, color string, sortOrder int) (*models.Group, error) {
	g := &models.Group{}
	err := pool.QueryRow(ctx,
		`INSERT INTO groups (name, color, sort_order) VALUES ($1, $2, $3)
		 RETURNING id, name, color, sort_order, created_at`,
		name, color, sortOrder,
	).Scan(&g.ID, &g.Name, &g.Color, &g.SortOrder, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("CreateGroup: %w", err)
	}
	return g, nil
}

// UpdateGroup updates a group's name, color, and/or sort_order.
func UpdateGroup(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, input models.GroupInput) (*models.Group, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Color != nil {
		setClauses = append(setClauses, fmt.Sprintf("color = $%d", argIdx))
		args = append(args, *input.Color)
		argIdx++
	}
	if input.SortOrder != nil {
		setClauses = append(setClauses, fmt.Sprintf("sort_order = $%d", argIdx))
		args = append(args, *input.SortOrder)
		argIdx++
	}

	if len(setClauses) == 0 {
		// No changes, just return current
		g := &models.Group{}
		err := pool.QueryRow(ctx,
			`SELECT id, name, color, sort_order, created_at FROM groups WHERE id = $1`, id,
		).Scan(&g.ID, &g.Name, &g.Color, &g.SortOrder, &g.CreatedAt)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, nil
			}
			return nil, fmt.Errorf("UpdateGroup get: %w", err)
		}
		return g, nil
	}

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE groups SET %s WHERE id = $%d
		 RETURNING id, name, color, sort_order, created_at`,
		stringsJoin(setClauses, ", "), argIdx,
	)

	g := &models.Group{}
	err := pool.QueryRow(ctx, query, args...).Scan(
		&g.ID, &g.Name, &g.Color, &g.SortOrder, &g.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("UpdateGroup: %w", err)
	}
	return g, nil
}

// DeleteGroup deletes a group by ID. Link-group associations are removed via ON DELETE CASCADE.
func DeleteGroup(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM groups WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("DeleteGroup: %w", err)
	}
	return nil
}

// GetGroupByName returns a group by its name.
func GetGroupByName(ctx context.Context, pool *pgxpool.Pool, name string) (*models.Group, error) {
	g := &models.Group{}
	err := pool.QueryRow(ctx,
		`SELECT id, name, color, sort_order, created_at FROM groups WHERE name = $1`, name,
	).Scan(&g.ID, &g.Name, &g.Color, &g.SortOrder, &g.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetGroupByName: %w", err)
	}
	return g, nil
}

// GetOrCreateGroupByName returns an existing group by name, or creates one with defaults.
func GetOrCreateGroupByName(ctx context.Context, pool *pgxpool.Pool, name, color string, sortOrder int) (*models.Group, error) {
	g, err := GetGroupByName(ctx, pool, name)
	if err != nil {
		return nil, err
	}
	if g != nil {
		return g, nil
	}
	return CreateGroup(ctx, pool, name, color, sortOrder)
}

// stringsJoin is a helper to avoid importing strings for the simple group query builder.
func stringsJoin(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	result := elems[0]
	for _, e := range elems[1:] {
		result += sep + e
	}
	return result
}
