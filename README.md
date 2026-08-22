# laterbase

A lightweight, self-hosted read-later link collector. Share a link from WhatsApp (via a Hermes agent), paste it into the dashboard, and laterbase fetches the page, categorizes it with an LLM, generates a summary, and files it into a group — all in a clean, minimal dashboard.

```
WhatsApp → Hermes → POST /api/links → River job queue → fetch → LLM categorize → store
                                                   ↓
                                           Dashboard (React SPA)
```

## Features

- **Capture anywhere** — send `link <url>` to a Hermes WhatsApp agent, or add links via the dashboard's Add Link dialog
- **Automatic processing** — each link is fetched, cleaned, categorized (`article`, `github`, `youtube`, `docs`, `tools`, ...), summarized (gpt-4o-mini), and assigned to a group by a River background job
- **Honest statuses** — inaccessible, paywalled, or failed pages are clearly labeled; content is never fabricated
- **Groups** — organize links into color-coded groups (articles, github, docs, videos, tools, uncategorized) or create your own
- **Content preview** — read saved page text inline, with an "Open in browser" fallback for sites that block embedding
- **Notes** — attach free-form notes to any link
- **Pluggable fetcher** — simple HTTP + readability extraction by default, with an optional Crawl4AI service for JS-heavy pages

## Tech Stack

- **Backend**: Go (chi router, pgx pool, goose migrations, River job queue)
- **Frontend**: React + Vite, Tailwind v4, TanStack Query, React Router
- **Database**: PostgreSQL 17
- **LLM**: gpt-4o-mini via OpenAI API
- **Optional**: Crawl4AI (Docker profile) for JavaScript-heavy sites

## Quick Start

Requires Docker (Docker Compose v2).

```bash
cp .env.example .env   # then fill in LLM_API_KEY
docker compose up -d   # starts postgres + api (air hot-reload) + frontend (vite dev)
```

Optionally start Crawl4AI:

```bash
docker compose --profile crawl4ai up -d
```

| Service | URL |
|---------|-----|
| Dashboard | http://localhost:5173 |
| API | http://localhost:8080 |
| PostgreSQL | localhost:5433 |
| Crawl4AI (optional) | http://localhost:11235 |

Database migrations run automatically on API startup via goose. The backend hot-reloads on `.go` changes (Air) and the frontend uses Vite HMR.

## Configuration

Environment variables (see `.env.example`):

| Variable | Required | Description |
|----------|----------|-------------|
| `LLM_API_KEY` | Yes | OpenAI API key for categorization + summarization |
| `CRAWL4AI_URL` | No | Crawl4AI endpoint; empty = simple HTTP fetcher |

The fetcher mode can also be toggled at runtime:

```bash
curl -X PUT localhost:8080/api/config \
  -H 'Content-Type: application/json' \
  -d '{"fetcher_mode": "crawl4ai"}'
```

## Hermes Integration

Add the following to your Hermes agent's `SOUL.md` to capture links from WhatsApp:

```
### `link <url>` or `save <url>`
- POST the URL to the laterbase backend at http://localhost:8080/api/links
- Body: {"url": "<the url>"}
- Acknowledge with: "Saved. I'll fetch and summarize it shortly."
- If the API is unreachable, respond: "laterbase isn't running. Start it first."
```

## API

All routes are under `/api` (plus `GET /health`). Highlights:

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/links` | Ingest a link — body: `{"url": "..."}` |
| `GET` | `/api/links` | List links — supports `?group_id=`, `?status=`, `?search=` |
| `GET` | `/api/links/:id` | Single link with full extracted text |
| `PUT` | `/api/links/:id` | Update group, title, source type, or notes |
| `DELETE` | `/api/links/:id` | Delete a link |
| `POST` | `/api/links/:id/reprocess` | Re-run fetch + categorize + summarize |
| `GET` / `POST` / `PUT` / `DELETE` | `/api/groups[...]` | Manage groups |
| `GET` | `/api/stats` | Counts by status, group, and source type |
| `GET` / `PUT` | `/api/config` | Read/update fetcher mode |
| `GET` | `/health` | DB ping + queue status |

See `cmd/server/main.go` for the full route table.

## Project Structure

```
laterbase/
├── docker-compose.yml
├── .env.example
├── PLAN.md                  # Original build plan & design decisions
├── backend/
│   ├── cmd/server/main.go   # Route table & wiring
│   └── internal/
│       ├── config/          # Env config
│       ├── db/              # pgx pool, queries, embedded goose migrations
│       ├── handlers/        # HTTP handlers (chi)
│       ├── fetcher/         # Simple HTTP + Crawl4AI implementations
│       ├── categorizer/     # gpt-4o-mini call
│       ├── workers/         # River ProcessLink job
│       └── content/         # HTML sanitizer
└── frontend/
    └── src/
        ├── pages/           # Dashboard, LinkDetail, Groups
        ├── components/      # LinkCard, ContentPreview, etc.
        ├── hooks/           # TanStack Query hooks
        └── api/             # Axios client
```

## Documentation

- [PLAN.md](PLAN.md) — full architecture, data model, and design rationale
- [PRODUCT.md](PRODUCT.md) — product context and positioning
- [DESIGN.md](DESIGN.md) — visual system and tokens
