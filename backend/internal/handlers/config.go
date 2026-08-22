package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devkrishna/laterbase/backend/internal/db"
)

// ConfigHandler handles config endpoints.
type ConfigHandler struct {
	pool *pgxpool.Pool
}

// NewConfigHandler creates a new ConfigHandler.
func NewConfigHandler(pool *pgxpool.Pool) *ConfigHandler {
	return &ConfigHandler{pool: pool}
}

// ConfigResponse is the JSON response for config.
type ConfigResponse struct {
	FetcherMode string `json:"fetcher_mode"`
}

// ConfigUpdateRequest is the JSON body for updating config.
type ConfigUpdateRequest struct {
	FetcherMode string `json:"fetcher_mode"`
}

// Get handles GET /api/config
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	mode, err := db.GetFetcherMode(r.Context(), h.pool)
	if err != nil {
		log.Printf("GetFetcherMode error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get config"})
		return
	}

	writeJSON(w, http.StatusOK, ConfigResponse{FetcherMode: mode})
}

// Update handles PUT /api/config
func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.FetcherMode != "simple" && req.FetcherMode != "crawl4ai" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "fetcher_mode must be 'simple' or 'crawl4ai'"})
		return
	}

	if err := db.SetConfig(r.Context(), h.pool, "fetcher_mode", req.FetcherMode); err != nil {
		log.Printf("SetConfig error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update config"})
		return
	}

	writeJSON(w, http.StatusOK, ConfigResponse{FetcherMode: req.FetcherMode})
}
