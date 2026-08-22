# Link Groupy — Agent Instructions

## Quick start

```bash
cp .env.example .env   # then fill in LLM_API_KEY
docker compose up -d   # starts postgres + api (air hot-reload) + frontend (vite dev)
docker compose --profile crawl4ai up -d  # also start Crawl4AI
```

- Backend: `localhost:8080` (Air hot-reloads on `.go` changes)
- Frontend: `localhost:5173` (Vite HMR, proxies `/api` → `http://api:8080`)
- PostgreSQL: `localhost:5433` (maps to container 5432)

## Architecture

```
WhatsApp → Hermes → POST /api/links → River job queue → fetch → LLM categorize → store
                                                   ↓
                                           Dashboard (React SPA)
```

- **Go backend** (`backend/`): chi router, pgx pool, goose migrations (embedded), River job queue
- **React frontend** (`frontend/`): Vite + Tailwind v4 + TanStack Query + React Router
- **DB migrations**: `backend/internal/db/migrations/` — auto-ran on startup via goose

## Key conventions

- **No tests, no linter, no formatter** — none of these are configured. Don't look for them.
- **`noUnusedLocals` / `noUnusedParameters`** are explicitly `false` in tsconfig — unused vars are fine.
- **TypeScript build**: `tsc -b && vite build` (not just `vite build`).
- **Tailwind v4** — uses `@import "tailwindcss"` (no `tailwind.config.js`). CSS custom properties for Linear-style tokens in `index.css`.
- **`@/*` path alias** is in tsconfig `paths` but **no Vite resolve alias** is configured — importing `@/...` will fail at runtime.
- **Go version**: `go 1.25.7` in `go.mod`. Module path: `github.com/devkrishna/link-groupy/backend`.
- **`writeJSON` helper** lives in `handlers/links.go` — reused across handler files in the same package.
- **`strings` not imported** in `groups.go` — uses custom `stringsJoin` instead.
- **Content storage**: `content_text` column stores cleaned *plain text* (from `fetcher.cleanHTML`), not HTML. `ContentPreview.tsx` renders a persistent toolbar (URL, "Show text"/"Hide text" toggle for content_text, "Open in browser ↗" link) above the iframe. The iframe fills the remaining space — no CSP/X-Frame-Options detection overlay needed, since the "Open in browser" link is always visible.
- **Notes**: Each link has an optional `notes` TEXT column. Update via `PUT /api/links/{id}` with `{"notes": "..."}`. Send `"notes": ""` to clear. The LinkDetail page has an inline editor with Save/Cancel/Clear buttons.
- **CORS**: wide-open in dev (`main.go`). No CORS config for production.

## API

All routes under `/api` and `/health`. See `cmd/server/main.go:73-92` for the full route table.

## Hermes integration (WhatsApp)

Hermes routes `link <url>` → `POST localhost:8080/api/links {"url": "..."}`.
See `PLAN.md` "Capture Flow" for the exact SOUL.md addition.

## Project status

Nearly all of `PLAN.md` is implemented. Rough edges are noted in that doc. The `content/sanitizer.go` file exists but is not wired into the processing pipeline.

## Design Context

Register: **product** · Platform: **web** · Default theme: **light** (warm editorial palette) with dark toggle
- PRODUCT.md — strategic context (users, positioning, design principles)
- DESIGN.md — visual system (colors, typography, components, tokens)
- Live mode configured — run `$impeccable live` for in-browser iteration
