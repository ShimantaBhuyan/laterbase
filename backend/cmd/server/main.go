package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/devkrishna/laterbase/backend/internal/config"
	"github.com/devkrishna/laterbase/backend/internal/db"
	"github.com/devkrishna/laterbase/backend/internal/handlers"
	"github.com/devkrishna/laterbase/backend/internal/workers"
)

func main() {
	// Load config
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Create River client
	workersSet := river.NewWorkers()
	processLinkWorker := workers.NewProcessLinkWorker(pool, cfg)
	river.AddWorker(workersSet, processLinkWorker)
	river.AddWorker(workersSet, workers.NewReprocessAllWorker(pool))

	riverClient, err := river.NewClient[pgx.Tx](riverpgxv5.New(pool), &river.Config{
		Workers: workersSet,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 25},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create River client: %v", err)
	}

	// Start River
	if err := riverClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start River: %v", err)
	}

	// Create handlers
	healthHandler := handlers.NewHealthHandler(pool)
	linkHandler := handlers.NewLinkHandler(pool, riverClient)
	groupHandler := handlers.NewGroupHandler(pool)
	statsHandler := handlers.NewStatsHandler(pool)
	configHandler := handlers.NewConfigHandler(pool)
	askHandler := handlers.NewAskHandler(pool, cfg)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(corsMiddleware)
	// Per-request timeout, except SSE endpoints (/api/ask) which stream long-running answers
	r.Use(timeoutUnless("/api/ask", 30*time.Second))

	// Routes
	r.Get("/health", healthHandler.Check)

	r.Route("/api", func(r chi.Router) {
		r.Post("/links", linkHandler.Create)
		r.Post("/links/batch", linkHandler.CreateBatch)
		r.Post("/links/reprocess-all", linkHandler.ReprocessAll)
		r.Get("/links", linkHandler.List)
		r.Get("/links/{id}", linkHandler.Get)
		r.Put("/links/{id}", linkHandler.Update)
		r.Delete("/links/{id}", linkHandler.Delete)
		r.Post("/links/{id}/reprocess", linkHandler.Reprocess)

		r.Get("/groups", groupHandler.List)
		r.Post("/groups", groupHandler.Create)
		r.Put("/groups/{id}", groupHandler.Update)
		r.Delete("/groups/{id}", groupHandler.Delete)

		r.Get("/stats", statsHandler.Get)

		r.Get("/config", configHandler.Get)
		r.Put("/config", configHandler.Update)

		r.Post("/ask", askHandler.Ask)
	})

	// Start HTTP server
	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0, // SSE (/api/ask) streams can run long; per-route timeouts handle abuse
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	// Stop River first
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := riverClient.Stop(shutdownCtx); err != nil {
		log.Printf("River stop error: %v", err)
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Close DB pool
	pool.Close()
	log.Println("Server stopped")
}

// corsMiddleware allows all origins in development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// timeoutUnless applies a request timeout except for the given path,
// which streams long-running responses.
func timeoutUnless(path string, d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == path {
				next.ServeHTTP(w, r)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
