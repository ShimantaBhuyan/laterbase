package handlers

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devkrishna/laterbase/backend/internal/models"
)

// StatsHandler handles the stats endpoint.
type StatsHandler struct {
	pool *pgxpool.Pool
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(pool *pgxpool.Pool) *StatsHandler {
	return &StatsHandler{pool: pool}
}

// Get handles GET /api/stats
func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	stats := models.Stats{
		ByStatus:     make(map[string]int),
		ByGroup:      make(map[string]int),
		BySourceType: make(map[string]int),
	}

	// Total links
	err := h.pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM links`).Scan(&stats.TotalLinks)
	if err != nil {
		log.Printf("Stats total error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get stats"})
		return
	}

	// By status
	statusRows, err := h.pool.Query(r.Context(),
		`SELECT status, COUNT(*) FROM links GROUP BY status`)
	if err != nil {
		log.Printf("Stats by status error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get stats"})
		return
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			continue
		}
		stats.ByStatus[status] = count
	}

	// By group (via link_groups join table)
	groupRows, err := h.pool.Query(r.Context(),
		`SELECT g.name, COUNT(lg.link_id)
		 FROM groups g
		 LEFT JOIN link_groups lg ON lg.group_id = g.id
		 GROUP BY g.name`)
	if err != nil {
		log.Printf("Stats by group error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get stats"})
		return
	}
	defer groupRows.Close()
	for groupRows.Next() {
		var name string
		var count int
		if err := groupRows.Scan(&name, &count); err != nil {
			continue
		}
		stats.ByGroup[name] = count
	}

	// By source_type
	sourceRows, err := h.pool.Query(r.Context(),
		`SELECT COALESCE(source_type, 'unknown'), COUNT(*) FROM links GROUP BY source_type`)
	if err != nil {
		log.Printf("Stats by source_type error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get stats"})
		return
	}
	defer sourceRows.Close()
	for sourceRows.Next() {
		var sourceType string
		var count int
		if err := sourceRows.Scan(&sourceType, &count); err != nil {
			continue
		}
		stats.BySourceType[sourceType] = count
	}

	// Compute convenience counts
	stats.ReadyCount = stats.ByStatus[models.StatusReady]
	stats.ProcessingCount = stats.ByStatus[models.StatusFetching] + stats.ByStatus[models.StatusProcessing]
	stats.FailedCount = stats.ByStatus[models.StatusFailed] + stats.ByStatus[models.StatusInaccessible] + stats.ByStatus[models.StatusPaywalled]

	// Total LLM cost (from append-only link_costs so deletion doesn't affect totals)
	err = h.pool.QueryRow(r.Context(), `SELECT COALESCE(SUM(llm_cost), 0) FROM link_costs`).Scan(&stats.TotalLLMCost)
	if err != nil {
		log.Printf("Stats llm_cost error: %v", err)
		// Non-fatal — cost is just a display nicety
		stats.TotalLLMCost = 0
	}

	writeJSON(w, http.StatusOK, stats)
}
