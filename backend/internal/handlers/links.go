package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/devkrishna/laterbase/backend/internal/db"
	"github.com/devkrishna/laterbase/backend/internal/models"
	"github.com/devkrishna/laterbase/backend/internal/workers"
)

// LinkHandler handles link-related HTTP requests.
type LinkHandler struct {
	pool        *pgxpool.Pool
	riverClient *river.Client[pgx.Tx]
}

// NewLinkHandler creates a new LinkHandler.
func NewLinkHandler(pool *pgxpool.Pool, riverClient *river.Client[pgx.Tx]) *LinkHandler {
	return &LinkHandler{
		pool:        pool,
		riverClient: riverClient,
	}
}

// Create handles POST /api/links
func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.LinkInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url is required"})
		return
	}

	link, err := db.CreateLink(r.Context(), h.pool, input.URL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "this link has already been added"})
			return
		}
		log.Printf("CreateLink error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create link"})
		return
	}

	// Enqueue processing job
	_, err = h.riverClient.Insert(r.Context(), &workers.ProcessLinkArgs{
		LinkID: link.ID,
		URL:    link.URL,
	}, nil)
	if err != nil {
		log.Printf("Failed to enqueue process link job: %v", err)
		// Still return 201 — the link was created
	}

	writeJSON(w, http.StatusCreated, link)
}

// List handles GET /api/links
func (h *LinkHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	groupID := q.Get("group_id")
	status := q.Get("status")
	search := q.Get("search")
	sort := q.Get("sort")
	order := q.Get("order")

	limit := 0
	offset := 0
	if l := q.Get("limit"); l != "" {
		if v, err := parseInt(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := q.Get("offset"); o != "" {
		if v, err := parseInt(o); err == nil && v >= 0 {
			offset = v
		}
	}

	links, total, err := db.ListLinks(r.Context(), h.pool, groupID, status, search, sort, order, limit, offset)
	if err != nil {
		log.Printf("ListLinks error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list links"})
		return
	}

	if links == nil {
		links = []*models.Link{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"links": links,
		"total": total,
	})
}

// Get handles GET /api/links/{id}
func (h *LinkHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	link, err := db.GetLink(r.Context(), h.pool, id)
	if err != nil {
		log.Printf("GetLink error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get link"})
		return
	}

	if link == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "link not found"})
		return
	}

	writeJSON(w, http.StatusOK, link)
}

// Update handles PUT /api/links/{id}
func (h *LinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var upd models.LinkUpdate
	if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	link, err := db.UpdateLink(r.Context(), h.pool, id, upd)
	if err != nil {
		log.Printf("UpdateLink error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update link"})
		return
	}

	if link == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "link not found"})
		return
	}

	writeJSON(w, http.StatusOK, link)
}

// Delete handles DELETE /api/links/{id}
func (h *LinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	if err := db.DeleteLink(r.Context(), h.pool, id); err != nil {
		log.Printf("DeleteLink error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete link"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Reprocess handles POST /api/links/{id}/reprocess
func (h *LinkHandler) Reprocess(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	// Get the link to check it exists and get its URL
	link, err := db.GetLink(r.Context(), h.pool, id)
	if err != nil {
		log.Printf("GetLink error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get link"})
		return
	}

	if link == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "link not found"})
		return
	}

	// Reset status to 'new'
	if err := db.SetLinkStatus(r.Context(), h.pool, id, models.StatusNew, nil); err != nil {
		log.Printf("SetLinkStatus error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset link status"})
		return
	}

	// Drop stale auto-assigned groups; manually assigned ones survive re-categorization
	if err := db.ClearAutoGroups(r.Context(), h.pool, []uuid.UUID{id}); err != nil {
		log.Printf("ClearAutoGroups error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to clear link groups"})
		return
	}

	// Enqueue new processing job
	_, err = h.riverClient.Insert(r.Context(), &workers.ProcessLinkArgs{
		LinkID: id,
		URL:    link.URL,
	}, nil)
	if err != nil {
		log.Printf("Failed to enqueue reprocess job: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to enqueue reprocess job"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "reprocess queued"})
}

// ReprocessAll handles POST /api/links/reprocess-all
func (h *LinkHandler) ReprocessAll(w http.ResponseWriter, r *http.Request) {
	// Guard against concurrent bulk runs — only one active reprocess_all job allowed
	var active int
	err := h.pool.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM river_job
		WHERE kind = $1 AND state IN ($2, $3, $4, $5)
	`, "reprocess_all", "available", "scheduled", "running", "retryable").Scan(&active)
	if err != nil {
		log.Printf("ReprocessAll guard query error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check reprocess status"})
		return
	}
	if active > 0 {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "a bulk reprocess is already in progress"})
		return
	}

	job, err := h.riverClient.Insert(r.Context(), &workers.ReprocessAllArgs{}, nil)
	if err != nil {
		log.Printf("Failed to enqueue reprocess-all job: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to enqueue reprocess job"})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"queued": true,
		"job_id": job.Job.ID,
	})
}

// writeJSON is a helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("writeJSON error: %v", err)
		}
	}
}

func parseInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
