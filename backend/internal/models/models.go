package models

import (
	"time"

	"github.com/google/uuid"
)

// Link status values
const (
	StatusNew          = "new"
	StatusFetching     = "fetching"
	StatusProcessing   = "processing"
	StatusReady        = "ready"
	StatusInaccessible = "inaccessible"
	StatusPaywalled    = "paywalled"
	StatusFailed       = "failed"
)

// Source type values
const (
	SourceArticle  = "article"
	SourceGitHub   = "github"
	SourceYouTube  = "youtube"
	SourceTwitter  = "twitter"
	SourceDocs     = "docs"
	SourceTools    = "tools"
	SourceOther    = "other"
)

// Link represents a saved link with its metadata.
type Link struct {
	ID           uuid.UUID    `json:"id"`
	URL          string       `json:"url"`
	Title        *string      `json:"title"`
	Summary      *string      `json:"summary"`
	ContentText  *string      `json:"content_text,omitempty"`
	SourceType   *string      `json:"source_type"`
	GroupIDs     []uuid.UUID  `json:"group_ids"`
	Notes        *string      `json:"notes"`
	Status       string       `json:"status"`
	ErrorMessage *string      `json:"error_message,omitempty"`
	FaviconURL   *string      `json:"favicon_url"`
	OgImageURL   *string      `json:"og_image_url"`
	InputTokens  *int         `json:"input_tokens"`
	OutputTokens *int         `json:"output_tokens"`
	LLMCost      *float64     `json:"llm_cost"`
	CreatedAt    time.Time    `json:"created_at"`
	ProcessedAt  *time.Time   `json:"processed_at"`
}

// LinkInput is the payload for creating a new link.
type LinkInput struct {
	URL string `json:"url" validate:"required,url"`
}

// LinkUpdate is the payload for updating an existing link.
type LinkUpdate struct {
	Title         *string     `json:"title,omitempty"`
	SourceType    *string     `json:"source_type,omitempty"`
	GroupIDs      []uuid.UUID `json:"group_ids,omitempty"`
	Notes         *string     `json:"notes,omitempty"`
	IsInteresting *bool       `json:"is_interesting,omitempty"`
}

// Group represents a group/category for organizing links.
type Group struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// GroupWithCount is a group with the number of links in it.
type GroupWithCount struct {
	Group
	LinkCount int `json:"link_count"`
}

// GroupInput is the payload for creating or updating a group.
type GroupInput struct {
	Name      *string `json:"name,omitempty"`
	Color     *string `json:"color,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}

// Config represents a key-value configuration entry.
type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Stats holds aggregate statistics about links.
type Stats struct {
	TotalLinks       int            `json:"total_links"`
	ByStatus         map[string]int `json:"by_status"`
	ByGroup          map[string]int `json:"by_group"`
	BySourceType     map[string]int `json:"by_source_type"`
	ReadyCount       int            `json:"ready_count"`
	ProcessingCount  int            `json:"processing_count"`
	FailedCount      int            `json:"failed_count"`
	TotalLLMCost     float64        `json:"total_llm_cost"`
}
