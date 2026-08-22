package workers

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/riverqueue/river"

	"github.com/devkrishna/laterbase/backend/internal/categorizer"
	"github.com/devkrishna/laterbase/backend/internal/config"
	"github.com/devkrishna/laterbase/backend/internal/db"
	"github.com/devkrishna/laterbase/backend/internal/fetcher"
	"github.com/devkrishna/laterbase/backend/internal/models"
)

// Known LLM pricing per model ($ per 1M tokens).
// Update these when models or pricing change.
var llmPricing = map[string]struct{ input, output float64 }{
	"gpt-4o-mini": {0.15, 0.60},
	"gpt-4o":      {2.50, 10.00},
	"gpt-4":       {30.00, 60.00},
	"gpt-3.5-turbo": {0.50, 1.50},
}

func computeCost(model string, inputTokens, outputTokens int) float64 {
	p, ok := llmPricing[model]
	if !ok {
		// Fallback to gpt-4o-mini pricing for unknown models
		p = llmPricing["gpt-4o-mini"]
	}
	cost := (float64(inputTokens) * p.input / 1_000_000) + (float64(outputTokens) * p.output / 1_000_000)
	return cost
}

// ProcessLinkArgs are the River job arguments for processing a link.
type ProcessLinkArgs struct {
	LinkID uuid.UUID `json:"link_id"`
	URL    string    `json:"url"`
}

// Kind returns the River job kind string.
func (ProcessLinkArgs) Kind() string { return "process_link" }

// ProcessLinkWorker processes links by fetching, categorizing, and storing results.
type ProcessLinkWorker struct {
	river.WorkerDefaults[ProcessLinkArgs]
	pool    *pgxpool.Pool
	cfg     *config.Config
}

// NewProcessLinkWorker creates a new ProcessLinkWorker.
func NewProcessLinkWorker(pool *pgxpool.Pool, cfg *config.Config) *ProcessLinkWorker {
	return &ProcessLinkWorker{
		pool: pool,
		cfg:  cfg,
	}
}

// Work executes the link processing pipeline.
func (w *ProcessLinkWorker) Work(ctx context.Context, job *river.Job[ProcessLinkArgs]) error {
	args := job.Args
	log.Printf("Processing link %s: %s", args.LinkID, args.URL)

	// Step 1: Set status = 'fetching'
	if err := db.SetLinkStatus(ctx, w.pool, args.LinkID, models.StatusFetching, nil); err != nil {
		return err
	}

	// Step 2: Determine fetcher
	mode, err := db.GetFetcherMode(ctx, w.pool)
	if err != nil {
		mode = "simple"
	}

	var f fetcher.Fetcher
	if mode == "crawl4ai" && w.cfg.Crawl4AIURL != "" {
		f = fetcher.NewCrawl4AIFetcher(w.cfg.Crawl4AIURL)
	} else {
		f = fetcher.NewSimpleFetcher()
	}

	// Step 3: Fetch the URL
	result := f.Fetch(ctx, args.URL)

	// Handle fetch errors
	if result.Error != nil {
		errMsg := result.Error.Error()
		switch {
		case errMsg == "inaccessible":
			_ = db.SetLinkStatus(ctx, w.pool, args.LinkID, models.StatusInaccessible, &errMsg)
			return nil // Not a retriable error
		case result.IsPaywalled:
			_ = db.SetLinkStatus(ctx, w.pool, args.LinkID, models.StatusPaywalled, &errMsg)
			return nil // Not a retriable error
		default:
			_ = db.SetLinkStatus(ctx, w.pool, args.LinkID, models.StatusFailed, &errMsg)
			return nil // Not a retriable error
		}
	}

	// Check paywall from FetchResult even if no error
	if result.IsPaywalled {
		errMsg := "content is behind a paywall"
		_ = db.SetLinkStatus(ctx, w.pool, args.LinkID, models.StatusPaywalled, &errMsg)
		return nil
	}

	// Step 4: Set status = 'processing'
	if err := db.SetLinkStatus(ctx, w.pool, args.LinkID, models.StatusProcessing, nil); err != nil {
		return err
	}

	// Step 5: Call Categorize
	catResult, err := categorizer.Categorize(ctx, w.cfg.LLMAPIKey, w.cfg.LLMModel, args.URL, result.ContentText)
	if err != nil {
		log.Printf("Categorize error for %s: %v", args.LinkID, err)
		// Fallback categorization
		catResult, _ = categorizer.Categorize(ctx, "", "", args.URL, result.ContentText)
	}

	// Step 6: Find the suggested group
	group, err := db.GetGroupByName(ctx, w.pool, catResult.SuggestedGroup)
	if err != nil {
		log.Printf("GetGroupByName error: %v", err)
	}
	var groupID *uuid.UUID
	if group != nil {
		groupID = &group.ID
	}

	// Determine the final title
	title := result.Title
	if catResult.Title != "" {
		title = catResult.Title
	}

	// Determine favicon and OG image
	var faviconURL *string
	if result.FaviconURL != "" {
		faviconURL = &result.FaviconURL
	}
	var ogImageURL *string
	if result.OgImageURL != "" {
		ogImageURL = &result.OgImageURL
	}

	// Step 7: Store results (no group_id in SetLinkProcessed anymore)
	llmCost := computeCost(catResult.ModelUsed, catResult.InputTokens, catResult.OutputTokens)
	if err := db.SetLinkProcessed(ctx, w.pool, args.LinkID, title, catResult.Summary, result.ContentText, catResult.SourceType, faviconURL, ogImageURL, catResult.InputTokens, catResult.OutputTokens, llmCost); err != nil {
		return err
	}

	// Step 7b: Record cost in append-only table (survives link deletion)
	if catResult.InputTokens > 0 || catResult.OutputTokens > 0 {
		if err := db.InsertLinkCost(ctx, w.pool, args.LinkID, args.URL, catResult.InputTokens, catResult.OutputTokens, llmCost, catResult.ModelUsed); err != nil {
			log.Printf("Failed to insert link cost for %s: %v", args.LinkID, err)
		}
	}

	// Step 8: Assign the auto-categorized group (only one — the most relevant)
	if groupID != nil {
		if err := db.AddLinkGroup(ctx, w.pool, args.LinkID, *groupID); err != nil {
			log.Printf("Failed to assign group for link %s: %v", args.LinkID, err)
			// Non-fatal — link is already processed
		}
	}

	log.Printf("Successfully processed link %s: %s", args.LinkID, title)
	return nil
}
