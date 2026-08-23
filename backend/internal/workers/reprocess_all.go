package workers

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/devkrishna/laterbase/backend/internal/db"
	"github.com/devkrishna/laterbase/backend/internal/models"
)

// reprocessChunkSize is how many links are fetched, reset, and enqueued
// per transaction while walking the links table.
const reprocessChunkSize = 500

// ReprocessAllArgs are the River job arguments for bulk reprocessing.
type ReprocessAllArgs struct{}

// Kind returns the River job kind string.
func (ReprocessAllArgs) Kind() string { return "reprocess_all" }

// ReprocessAllWorker walks all links in chunks, resets each chunk's status,
// and enqueues a ProcessLinkArgs job per link. Concurrency and pacing are
// handled by the River queue itself (MaxWorkers caps parallel process_link jobs).
type ReprocessAllWorker struct {
	river.WorkerDefaults[ReprocessAllArgs]
	pool *pgxpool.Pool
}

// NewReprocessAllWorker creates a new ReprocessAllWorker.
func NewReprocessAllWorker(pool *pgxpool.Pool) *ReprocessAllWorker {
	return &ReprocessAllWorker{pool: pool}
}

// Work executes the bulk reprocess pipeline in chunks.
func (w *ReprocessAllWorker) Work(ctx context.Context, job *river.Job[ReprocessAllArgs]) error {
	client := river.ClientFromContext[pgx.Tx](ctx)

	var total int
	lastID := uuid.Nil

	for {
		ids, urls, err := w.nextChunk(ctx, lastID)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			break
		}
		lastID = ids[len(ids)-1]

		if err := w.enqueueChunk(ctx, client, ids, urls); err != nil {
			return err
		}

		total += len(ids)
		log.Printf("Reprocess-all: %d links queued so far", total)
		if len(ids) < reprocessChunkSize {
			break
		}
	}

	log.Printf("Reprocess-all: done, queued %d links for processing", total)
	return nil
}

// nextChunk fetches the next page of link IDs/URLs using keyset pagination.
func (w *ReprocessAllWorker) nextChunk(ctx context.Context, afterID uuid.UUID) ([]uuid.UUID, []string, error) {
	rows, err := w.pool.Query(ctx, `
		SELECT id, url FROM links
		WHERE id > $1
		ORDER BY id ASC
		LIMIT $2
	`, afterID, reprocessChunkSize)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	var urls []string
	for rows.Next() {
		var id uuid.UUID
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		urls = append(urls, url)
	}
	return ids, urls, rows.Err()
}

// enqueueChunk atomically resets the chunk's statuses and enqueues one
// process_link job per link in a single transaction, so a rollback never
// leaves links stuck in 'new' without jobs.
func (w *ReprocessAllWorker) enqueueChunk(ctx context.Context, client *river.Client[pgx.Tx], ids []uuid.UUID, urls []string) error {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE links SET status = $1, error_message = NULL
		WHERE id = ANY($2)
	`, models.StatusNew, ids); err != nil {
		return err
	}

	// Drop stale auto-assigned groups; manually assigned ones survive re-categorization
	if err := db.ClearAutoGroupsTx(ctx, tx, ids); err != nil {
		return err
	}

	params := make([]river.InsertManyParams, len(ids))
	for i := range ids {
		params[i] = river.InsertManyParams{
			Args: &ProcessLinkArgs{LinkID: ids[i], URL: urls[i]},
		}
	}
	if _, err := client.InsertManyTx(ctx, tx, params); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
