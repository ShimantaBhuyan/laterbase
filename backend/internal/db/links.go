package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devkrishna/laterbase/backend/internal/models"
)

// CreateLink inserts a new link and returns it.
func CreateLink(ctx context.Context, pool *pgxpool.Pool, url string) (*models.Link, error) {
	link := &models.Link{}
	err := pool.QueryRow(ctx,
		`INSERT INTO links (url, status) VALUES ($1, $2)
		 RETURNING id, url, title, summary, content_text, source_type,
		           notes, status, error_message, favicon_url, og_image_url,
		           input_tokens, output_tokens, llm_cost, created_at, processed_at`,
		url, models.StatusNew,
	).Scan(
		&link.ID, &link.URL, &link.Title, &link.Summary, &link.ContentText,
		&link.SourceType, &link.Notes,
		&link.Status, &link.ErrorMessage,
		&link.FaviconURL, &link.OgImageURL,
		&link.InputTokens, &link.OutputTokens, &link.LLMCost,
		&link.CreatedAt, &link.ProcessedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("CreateLink: %w", err)
	}
	return link, nil
}

// GetLink returns a single link by ID, including its group_ids.
func GetLink(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.Link, error) {
	link := &models.Link{}
	err := pool.QueryRow(ctx,
	`SELECT id, url, title, summary, content_text, source_type,
	        notes, status, error_message, favicon_url, og_image_url,
	        input_tokens, output_tokens, llm_cost, created_at, processed_at
	 FROM links WHERE id = $1`, id,
	).Scan(
		&link.ID, &link.URL, &link.Title, &link.Summary, &link.ContentText,
		&link.SourceType, &link.Notes,
		&link.Status, &link.ErrorMessage,
		&link.FaviconURL, &link.OgImageURL,
		&link.InputTokens, &link.OutputTokens, &link.LLMCost,
		&link.CreatedAt, &link.ProcessedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetLink: %w", err)
	}

	// Fetch group_ids from the join table
	link.GroupIDs = []uuid.UUID{}
	groupRows, err := pool.Query(ctx, `SELECT group_id FROM link_groups WHERE link_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("GetLink groups: %w", err)
	}
	defer groupRows.Close()
	for groupRows.Next() {
		var gid uuid.UUID
		if err := groupRows.Scan(&gid); err != nil {
			return nil, fmt.Errorf("GetLink groups scan: %w", err)
		}
		link.GroupIDs = append(link.GroupIDs, gid)
	}

	return link, nil
}

// ListLinks returns links with optional filters and pagination.
// limit=0 means no pagination. Returns (links, totalCount, error).
func ListLinks(ctx context.Context, pool *pgxpool.Pool, groupID, status, search, sort, order string, limit, offset int) ([]*models.Link, int, error) {
	baseFrom := `FROM links l WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if groupID != "" {
		baseFrom += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM link_groups lg WHERE lg.link_id = l.id AND lg.group_id = $%d)", argIdx)
		args = append(args, groupID)
		argIdx++
	}
	if status != "" {
		baseFrom += fmt.Sprintf(" AND l.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if search != "" {
		baseFrom += fmt.Sprintf(" AND (l.title ILIKE $%d OR l.summary ILIKE $%d)", argIdx, argIdx+1)
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
		argIdx += 2
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) " + baseFrom
	if err := pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ListLinks count: %w", err)
	}

	// Sort
	sortCol := "l.created_at"
	switch sort {
	case "title":
		sortCol = "l.title"
	case "status":
		sortCol = "l.status"
	case "url":
		sortCol = "l.url"
	}
	sortDir := "DESC"
	if strings.ToUpper(order) == "ASC" {
		sortDir = "ASC"
	}

	// Fetch data
	selQuery := "SELECT l.id, l.url, l.title, l.summary, l.content_text, l.source_type, " +
		"l.notes, l.status, l.error_message, l.favicon_url, l.og_image_url, " +
		"l.input_tokens, l.output_tokens, l.llm_cost, l.created_at, l.processed_at " + baseFrom +
		fmt.Sprintf(" ORDER BY %s %s", sortCol, sortDir)

	if limit > 0 {
		selQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
	}

	rows, err := pool.Query(ctx, selQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("ListLinks: %w", err)
	}
	defer rows.Close()

	links, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*models.Link, error) {
		l := &models.Link{}
		err := row.Scan(
			&l.ID, &l.URL, &l.Title, &l.Summary, &l.ContentText,
			&l.SourceType, &l.Notes,
			&l.Status, &l.ErrorMessage,
			&l.FaviconURL, &l.OgImageURL,
			&l.InputTokens, &l.OutputTokens, &l.LLMCost,
			&l.CreatedAt, &l.ProcessedAt,
		)
		return l, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("ListLinks scan: %w", err)
	}

	// Batch load group_ids for all fetched links
	if len(links) > 0 {
		linkIndex := make(map[uuid.UUID]int, len(links))
		ids := make([]uuid.UUID, len(links))
		for i, l := range links {
			linkIndex[l.ID] = i
			ids[i] = l.ID
			links[i].GroupIDs = []uuid.UUID{}
		}
		groupRows, err := pool.Query(ctx, `SELECT link_id, group_id FROM link_groups WHERE link_id = ANY($1)`, ids)
		if err != nil {
			return nil, 0, fmt.Errorf("ListLinks groups: %w", err)
		}
		defer groupRows.Close()
		for groupRows.Next() {
			var linkID, groupID uuid.UUID
			if err := groupRows.Scan(&linkID, &groupID); err != nil {
				return nil, 0, fmt.Errorf("ListLinks groups scan: %w", err)
			}
			if idx, ok := linkIndex[linkID]; ok {
				links[idx].GroupIDs = append(links[idx].GroupIDs, groupID)
			}
		}
	}

	return links, total, nil
}

// UpdateLink updates title, source_type, group_ids, and notes for a link.
func UpdateLink(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, upd models.LinkUpdate) (*models.Link, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if upd.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *upd.Title)
		argIdx++
	}
	if upd.SourceType != nil {
		setClauses = append(setClauses, fmt.Sprintf("source_type = $%d", argIdx))
		args = append(args, *upd.SourceType)
		argIdx++
	}
	if upd.Notes != nil {
		setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, *upd.Notes)
		argIdx++
	}

	// Update links table columns (if any)
	if len(setClauses) > 0 {
		args = append(args, id)
		query := fmt.Sprintf(
			`UPDATE links SET %s WHERE id = $%d`,
			strings.Join(setClauses, ", "), argIdx,
		)
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			return nil, fmt.Errorf("UpdateLink: %w", err)
		}
	}

	// Update link_groups if GroupIDs is provided (even if empty array — clears all groups)
	if upd.GroupIDs != nil {
		if _, err := pool.Exec(ctx, `DELETE FROM link_groups WHERE link_id = $1`, id); err != nil {
			return nil, fmt.Errorf("UpdateLink delete groups: %w", err)
		}
		for _, gid := range upd.GroupIDs {
			if _, err := pool.Exec(ctx, `INSERT INTO link_groups (link_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, gid); err != nil {
				return nil, fmt.Errorf("UpdateLink insert group: %w", err)
			}
		}
	}

	// Toggle the special "interesting" group membership via the is_interesting flag.
	if upd.IsInteresting != nil {
		interestingGroup, err := GetOrCreateGroupByName(ctx, pool, "interesting", "#f59e0b", 99)
		if err != nil {
			return nil, fmt.Errorf("UpdateLink interesting group: %w", err)
		}
		if *upd.IsInteresting {
			if _, err := pool.Exec(ctx, `INSERT INTO link_groups (link_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, interestingGroup.ID); err != nil {
				return nil, fmt.Errorf("UpdateLink add interesting: %w", err)
			}
		} else {
			if _, err := pool.Exec(ctx, `DELETE FROM link_groups WHERE link_id = $1 AND group_id = $2`, id, interestingGroup.ID); err != nil {
				return nil, fmt.Errorf("UpdateLink remove interesting: %w", err)
			}
		}
	}

	return GetLink(ctx, pool, id)
}

// DeleteLink deletes a link by ID.
func DeleteLink(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM links WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("DeleteLink: %w", err)
	}
	return nil
}

// SetLinkStatus updates the status and optionally the error_message of a link.
func SetLinkStatus(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, status string, errorMsg *string) error {
	_, err := pool.Exec(ctx,
		`UPDATE links SET status = $1, error_message = $2 WHERE id = $3`,
		status, errorMsg, id,
	)
	if err != nil {
		return fmt.Errorf("SetLinkStatus: %w", err)
	}
	return nil
}

// SetLinkProcessed marks a link as ready with all processed fields filled in.
func SetLinkProcessed(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, title, summary, contentText, sourceType string, faviconURL, ogImageURL *string, inputTokens, outputTokens int, llmCost float64) error {
	now := time.Now()
	// Sanitize content to valid UTF-8 — PDFs and other binary-ish sources can
	// produce invalid byte sequences (including null bytes) that crash
	// PostgreSQL's UTF-8 check.
	safe := strings.ReplaceAll(contentText, "\x00", "")
	safe = strings.ToValidUTF8(safe, "")
	var inP, outP *int
	if inputTokens > 0 || outputTokens > 0 {
		inP = &inputTokens
		outP = &outputTokens
	}
	var costP *float64
	if llmCost > 0 {
		costP = &llmCost
	}
	_, err := pool.Exec(ctx,
		`UPDATE links SET status = $1, title = $2, summary = $3, content_text = $4,
		 source_type = $5, favicon_url = $6, og_image_url = $7, processed_at = $8,
		 input_tokens = $9, output_tokens = $10, llm_cost = $11, error_message = NULL
		 WHERE id = $12`,
		models.StatusReady, title, summary, safe, sourceType, faviconURL, ogImageURL, now, inP, outP, costP, id,
	)
	if err != nil {
		return fmt.Errorf("SetLinkProcessed: %w", err)
	}
	return nil
}

// InsertLinkCost records LLM cost data in the append-only link_costs table.
// This persists independently of link deletion so historical totals remain accurate.
func InsertLinkCost(ctx context.Context, pool *pgxpool.Pool, linkID uuid.UUID, url string, inputTokens, outputTokens int, llmCost float64, modelUsed string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO link_costs (link_id, url, input_tokens, output_tokens, llm_cost, model_used) VALUES ($1, $2, $3, $4, $5, $6)`,
		linkID, url, inputTokens, outputTokens, llmCost, modelUsed,
	)
	if err != nil {
		return fmt.Errorf("InsertLinkCost: %w", err)
	}
	return nil
}

// AddLinkGroup adds a single group to a link (used by the auto-categorization worker).
func AddLinkGroup(ctx context.Context, pool *pgxpool.Pool, linkID, groupID uuid.UUID) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO link_groups (link_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		linkID, groupID,
	)
	if err != nil {
		return fmt.Errorf("AddLinkGroup: %w", err)
	}
	return nil
}
