package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devkrishna/laterbase/backend/internal/ask"
	"github.com/devkrishna/laterbase/backend/internal/config"
	"github.com/devkrishna/laterbase/backend/internal/models"
)

// AskHandler handles natural-language questions over the link archive.
type AskHandler struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

// NewAskHandler creates a new AskHandler.
func NewAskHandler(pool *pgxpool.Pool, cfg *config.Config) *AskHandler {
	return &AskHandler{pool: pool, cfg: cfg}
}

// Ask handles POST /api/ask — streams an SSE response:
//
//	event: citations  data: [{"link_id","reason"},...]   (after ranking)
//	event: delta      data: {"text": "..."}              (answer chunks)
//	event: done       data: {"status","answer","citations","input_tokens","output_tokens"}
//	event: error      data: {"error": "..."}
func (h *AskHandler) Ask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Question        string `json:"question"`
		IncludeFullText bool   `json:"include_full_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	input.Question = strings.TrimSpace(input.Question)
	if input.Question == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "question is required"})
		return
	}
	if h.cfg.LLMAPIKey == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "LLM_API_KEY is not configured"})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming unsupported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	writeEvent := func(event string, data interface{}) {
		b, err := json.Marshal(data)
		if err != nil {
			log.Printf("ask: marshal event %s: %v", event, err)
			return
		}
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		flusher.Flush()
	}

	result, err := ask.Run(r.Context(), h.pool, h.cfg.LLMAPIKey, h.cfg.AskLLMModel,
		input.Question, input.IncludeFullText,
		func(citations []models.AskCitation) {
			writeEvent("citations", citations)
		},
		func(delta string) {
			writeEvent("delta", map[string]string{"text": delta})
		},
	)
	if err != nil {
		if r.Context().Err() != nil {
			return // client went away mid-stream
		}
		log.Printf("Ask error: %v", err)
		writeEvent("error", map[string]string{"error": "couldn't reach the model"})
		return
	}

	status := "ok"
	if result == nil {
		status = "empty_archive"
	} else if len(result.Citations) == 0 {
		status = "no_matches"
	}

	payload := map[string]interface{}{
		"status":  status,
		"answer":  "",
		"model":   h.cfg.AskLLMModel,
		"citations": []models.AskCitation{},
	}
	if result != nil {
		payload["answer"] = result.Answer
		payload["input_tokens"] = result.InputTokens
		payload["output_tokens"] = result.OutputTokens
		if len(result.Citations) > 0 {
			payload["citations"] = result.Citations
		}
	}
	writeEvent("done", payload)
}
