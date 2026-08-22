package handlers

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler handles the health check endpoint.
type HealthHandler struct {
	pool *pgxpool.Pool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// HealthResponse is the JSON response for the health check.
type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// Check handles GET /health
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if err := h.pool.Ping(r.Context()); err != nil {
		dbStatus = "disconnected"
		log.Printf("Health check DB ping failed: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, HealthResponse{
			Status: "unhealthy",
			DB:     dbStatus,
		})
		return
	}

	writeJSON(w, http.StatusOK, HealthResponse{
		Status: "ok",
		DB:     dbStatus,
	})
}
