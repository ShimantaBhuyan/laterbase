package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devkrishna/laterbase/backend/internal/models"
)

// ListAskCandidates returns lightweight metadata for all ready links,
// used as the candidate pool for ask ranking.
func ListAskCandidates(ctx context.Context, pool *pgxpool.Pool) ([]models.AskCandidate, error) {
	rows, err := pool.Query(ctx,
		`SELECT l.id, l.url, COALESCE(l.title, l.url), COALESCE(l.summary, ''), l.created_at,
		        COALESCE((SELECT ARRAY_AGG(gg.group_id::TEXT) FROM link_groups gg WHERE gg.link_id = l.id), '{}')
		 FROM links l WHERE l.status = $1
		 ORDER BY l.created_at DESC`,
		models.StatusReady,
	)
	if err != nil {
		return nil, fmt.Errorf("ListAskCandidates: %w", err)
	}
	defer rows.Close()

	candidates := []models.AskCandidate{}
	for rows.Next() {
		var c models.AskCandidate
		if err := rows.Scan(&c.ID, &c.URL, &c.Title, &c.Summary, &c.CreatedAt, &c.GroupIDs); err != nil {
			return nil, fmt.Errorf("ListAskCandidates scan: %w", err)
		}
		if c.GroupIDs == nil {
			c.GroupIDs = []string{}
		}
		candidates = append(candidates, c)
	}
	return candidates, nil
}

// GetLinkContentText returns the stored plain-text content of a link.
func GetLinkContentText(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (string, error) {
	var content *string
	err := pool.QueryRow(ctx, `SELECT content_text FROM links WHERE id = $1`, id).Scan(&content)
	if err != nil {
		return "", fmt.Errorf("GetLinkContentText: %w", err)
	}
	if content == nil {
		return "", nil
	}
	return *content, nil
}

// InsertAskCost records LLM usage for an ask query in the append-only cost
// ledger. link_id is NULL since asks span multiple links; the question is
// stored in the url column for traceability.
func InsertAskCost(ctx context.Context, pool *pgxpool.Pool, question string, inputTokens, outputTokens int, llmCost float64, modelUsed string) error {
	url := "ask: " + question
	if len(url) > 500 {
		url = url[:500]
	}
	_, err := pool.Exec(ctx,
		`INSERT INTO link_costs (link_id, url, input_tokens, output_tokens, llm_cost, model_used)
		 VALUES (NULL, $1, $2, $3, $4, $5)`,
		url, inputTokens, outputTokens, llmCost, modelUsed,
	)
	if err != nil {
		return fmt.Errorf("InsertAskCost: %w", err)
	}
	return nil
}
