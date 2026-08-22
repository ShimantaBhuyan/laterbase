package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devkrishna/laterbase/backend/internal/db"
	"github.com/devkrishna/laterbase/backend/internal/models"
)

// GroupHandler handles group-related HTTP requests.
type GroupHandler struct {
	pool *pgxpool.Pool
}

// NewGroupHandler creates a new GroupHandler.
func NewGroupHandler(pool *pgxpool.Pool) *GroupHandler {
	return &GroupHandler{pool: pool}
}

// List handles GET /api/groups
func (h *GroupHandler) List(w http.ResponseWriter, r *http.Request) {
	groups, err := db.ListGroupsWithCounts(r.Context(), h.pool)
	if err != nil {
		log.Printf("ListGroupsWithCounts error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list groups"})
		return
	}

	if groups == nil {
		groups = []*models.GroupWithCount{}
	}

	writeJSON(w, http.StatusOK, groups)
}

// Create handles POST /api/groups
func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.GroupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Name == nil || *input.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	color := "#5e6ad2"
	if input.Color != nil {
		color = *input.Color
	}

	sortOrder := 0
	if input.SortOrder != nil {
		sortOrder = *input.SortOrder
	}

	group, err := db.CreateGroup(r.Context(), h.pool, *input.Name, color, sortOrder)
	if err != nil {
		log.Printf("CreateGroup error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create group"})
		return
	}

	writeJSON(w, http.StatusCreated, group)
}

// Update handles PUT /api/groups/{id}
func (h *GroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var input models.GroupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	group, err := db.UpdateGroup(r.Context(), h.pool, id, input)
	if err != nil {
		log.Printf("UpdateGroup error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update group"})
		return
	}

	if group == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "group not found"})
		return
	}

	writeJSON(w, http.StatusOK, group)
}

// Delete handles DELETE /api/groups/{id}
func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	if err := db.DeleteGroup(r.Context(), h.pool, id); err != nil {
		log.Printf("DeleteGroup error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete group"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
