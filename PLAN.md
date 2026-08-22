# laterbase — Read-Later Link Collector & Organizer

> **For Hermes:** Use the "Large Project Implementation from Plan" workflow (development-workflows skill) to implement this plan via parallel subagent delegation. **Do NOT start building without explicit user confirmation.**

**Goal:** A lightweight, self-hosted link-collection app. Links shared via WhatsApp to a Hermes agent (`link <url>`) are fetched, categorized by an LLM, summarized, and displayed in a minimal dark dashboard. Runs always-on via Orbstack.

**Architecture:** Go REST API (chi + pgx) ingests links → River job queue fetches page content + calls LLM → React SPA dashboard with Linear-inspired dark UI. PostgreSQL 17 in Orbstack.

**Tech Stack:** Go (chi, pgx, goose, river), React (Vite, TanStack Query, Tailwind v4), PostgreSQL 17, Docker Compose (Orbstack-managed), gpt-4o-mini, Crawl4AI (optional, toggleable).

---

## Feasibility Assessment

**Verdict: Highly feasible.** Every component is proven and already available on this machine. ~200 MB RAM idle. ~20 source files. Single-session build.

| Component | Feasibility | Risk |
|-----------|-------------|------|
| WhatsApp → Hermes → API ingestion | Trivial. Hermes SOUL.md routes `link <url>` messages (no `/` prefix — gateway-reserved). Same pattern as `adre today`. | None |
| Page content extraction | Medium. Simple HTTP + readability works for 80% of pages. Crawl4AI is available as toggle for JS-heavy sites. | Low — Crawl4AI Docker image already downloaded |
| LLM categorization + summary | Trivial. gpt-4o-mini via existing OpenAI key. ~2s per link. | None |
| Dashboard UI | Low. Linear-inspired dark dashboard, established React patterns. | None |
| Orbstack always-on | Low. Docker Compose project, Orbstack manages container lifecycle. | None |

---

## Capture Flow

```
iPhone Chrome tabs → Share → Copy link → Paste into WhatsApp self-chat
                                                    │
                                          "link https://example.com"
                                                    │
                                          Hermes Gateway receives
                                          (no / prefix, so not intercepted)
                                                    │
                                          Hermes agent routes per SOUL.md:
                                          matches "link <url>" pattern
                                                    │
                                          POST /api/links {"url": "..."}
                                          (localhost:8080)
                                                    │
                                          River worker picks it up:
                                          1. Fetch page content
                                          2. LLM categorize + summarize
                                          3. Store results
                                                    │
                                          Dashboard at localhost:5173
                                          shows the new link
```

### Hermes SOUL.md Addition

```
### `link <url>` or `save <url>`
- POST the URL to the laterbase backend at http://localhost:8080/api/links
- Body: {"url": "<the url>"}
- Acknowledge with: "Saved. I'll fetch and summarize it shortly."
- If the API is unreachable (Orbstack not running), respond: "laterbase isn't running. Start it from Orbstack first."
```

### Toggle Crawl4AI Mode

```
### `link mode crawl4ai` or `link mode simple`
- Toggle the fetcher mode by calling PUT /api/config with {"fetcher_mode": "crawl4ai"|"simple"}
- Acknowledge with current mode
```

---

## Data Model

### `links` table
```sql
CREATE TABLE links (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url           TEXT NOT NULL UNIQUE,
    title         TEXT,
    summary       TEXT,                    -- LLM-generated 2-3 sentence summary
    content_text  TEXT,                    -- cleaned page text (for in-app preview)
    source_type   TEXT,                    -- 'article','github','youtube','twitter','docs','tools','other'
    group_id      UUID REFERENCES groups(id) ON DELETE SET NULL,
    status        TEXT NOT NULL DEFAULT 'new',
    -- status values: 'new','fetching','processing','ready','inaccessible','paywalled','failed'
    error_message TEXT,                    -- set when status is 'inaccessible','paywalled', or 'failed'
    favicon_url   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at  TIMESTAMPTZ
);
```

### `groups` table
```sql
CREATE TABLE groups (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    color      TEXT NOT NULL DEFAULT '#5e6ad2',  -- Linear indigo
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### `config` table (single-row app config)
```sql
CREATE TABLE config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
-- Initial rows: ('fetcher_mode', 'simple'), ('fetcher_mode', 'crawl4ai' is alternative)
```

### Default groups (seeded on first migration)
- `articles` — blog posts, news, essays (#5e6ad2 indigo)
- `github` — repos, issues, PRs (#10b981 emerald)
- `docs` — documentation, API specs (#f59e0b amber)
- `videos` — YouTube, Vimeo (#ef4444 red)
- `tools` — SaaS, utilities (#8b5cf6 violet)
- `uncategorized` — fallback (#6b7280 gray)

---

## Honesty About Inaccessible Content

The system must NEVER fabricate or claim content for pages it cannot access. Status handling:

| Scenario | Status | Behavior |
|----------|--------|----------|
| Page fetched successfully | `ready` | Display title, summary, content preview |
| HTTP 403/401 (bot-blocked) | `inaccessible` | Show URL + domain + "This site blocks automated access" |
| Paywall detected (NYT, Medium, etc.) | `paywalled` | Show URL + domain + "Behind a paywall — open in browser to read" |
| HTTP 4xx/5xx, DNS failure, timeout | `failed` | Show URL + domain + specific error message |
| Still processing | `fetching` / `processing` | Show spinner + "Fetching..." / "Summarizing..." |

The LLM is NEVER called for `inaccessible`, `paywalled`, or `failed` links. The title comes from the URL or the HTTP response headers only. The summary field remains empty. The dashboard shows the link card with the appropriate status badge and no fabricated content.

---

## River Job: `ProcessLink`

```
ProcessLink(url, linkID):
  1. Mark status = 'fetching'
  2. Determine fetcher: check config.fetcher_mode
     - 'crawl4ai': POST to Crawl4AI /crawl endpoint (Docker service)
     - 'simple': net/http GET + readability extraction
  3. If fetch fails:
     - 403/401 → status = 'inaccessible', store error_message, RETURN
     - Paywall detected → status = 'paywalled', store error_message, RETURN
     - Other error → status = 'failed', store error_message, RETURN
  4. Extract: title, cleaned markdown, favicon
  5. Mark status = 'processing'
  6. Call gpt-4o-mini with prompt (truncate content to 3000 words):
     "Given this URL and extracted page content, return ONLY valid JSON:
      {
        "source_type": "article|github|youtube|twitter|docs|tools|other",
        "suggested_group": "one of: articles, github, docs, videos, tools, uncategorized",
        "summary": "2-3 sentence factual summary of what this page contains",
        "title": "cleaned title if the extracted one is missing or clearly wrong, else null"
      }
      If the content is inaccessible or empty, set source_type to 'other',
      suggested_group to 'uncategorized', and summary to an empty string."
  7. Store results, assign group (match existing group name or use 'uncategorized')
  8. Mark status = 'ready', set processed_at = now()
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/links` | Ingest new link. Body: `{"url": "..."}`. Enqueues River job. |
| `GET` | `/api/links` | List all links. Query: `?group_id=`, `?status=`, `?search=`, `?sort=`, `?order=` |
| `GET` | `/api/links/:id` | Single link with full content_text |
| `PUT` | `/api/links/:id` | Update group_id, title, or source_type |
| `DELETE` | `/api/links/:id` | Remove a link |
| `POST` | `/api/links/:id/reprocess` | Re-run fetch + categorize + summarize |
| `GET` | `/api/groups` | List all groups with link counts |
| `POST` | `/api/groups` | Create a group |
| `PUT` | `/api/groups/:id` | Rename, change color |
| `DELETE` | `/api/groups/:id` | Delete group (links fall to uncategorized) |
| `GET` | `/api/stats` | Total links, by status, by group, by source_type |
| `GET` | `/api/config` | Get current config (fetcher_mode) |
| `PUT` | `/api/config` | Update config. Body: `{"fetcher_mode": "crawl4ai"}` or `"simple"` |
| `GET` | `/health` | DB ping + River status |

---

## Dashboard Design — Linear-Inspired Dark UI

Reference: `popular-web-designs` skill → `templates/linear.app.md`

### Visual Identity
- **Canvas**: Near-black `#08090a` background — content emerges from darkness
- **Text**: `#f7f8f8` primary, `#8a8f98` secondary, `#62666d` muted
- **Surface elevation**: `rgba(255,255,255,0.02)` cards, `0.04` hover, `0.05` elevated
- **Borders**: `1px solid rgba(255,255,255,0.08)` — whisper-thin
- **Accent**: `#5e6ad2` (Linear indigo) for primary actions, group badges tinted per group color
- **Font**: Inter Variable with `"cv01", "ss03"` OpenType features, weight 510 as signature emphasis

### Pages

**1. Dashboard (`/`)** — Main Link Feed
- Top bar: "laterbase" + total count + "Add Link" button (indigo CTA)
- Search input: ghost-style, `rgba(255,255,255,0.02)` bg
- Group filter chips: horizontal scroll, pill-shaped (9999px), count badges
- Status filter: dropdown or inline pills
- Link cards (see below)
- Empty state: subtle illustration + "Share a link via WhatsApp — send `link <url>`"

**2. Link Detail (`/links/:id`)**
- Back arrow → dashboard
- Title bar: full title + "Open in Browser →" (external) + regroup dropdown
- Status badge: colored dot + label (ready = green, inaccessible = amber, paywalled = red)
- Summary block: rendered markdown
- Content preview: scrollable, sanitized HTML on dark surface
- Metadata row: domain, source type, group, date
- Actions: delete, reprocess, copy URL

**3. Groups (`/groups`)**
- List of groups with color swatches, names, link counts
- Inline rename, color picker (preset swatches)
- Delete with confirmation modal
- "New Group" button

### Link Card Component
```
┌──────────────────────────────────────────────────────┐
│  ◆ favicon   domain.com                   2 hours ago │
│                                                       │
│  Article Title — 20px weight 590, #f7f8f8            │
│                                                       │
│  Two-line summary truncated with ellipsis...          │
│  16px weight 400, #8a8f98                             │
│                                                       │
│  [articles]  [article]                                │
│   pill badge  source type badge                       │
└──────────────────────────────────────────────────────┘
```
Card background: `rgba(255,255,255,0.02)`, border: `1px solid rgba(255,255,255,0.08)`, radius: 8px.

Status variants:
- `ready`: no special indicator (normal card)
- `inaccessible`: amber dot + "Blocked" label in the summary area instead of summary text
- `paywalled`: red dot + "Paywalled" label
- `failed`: red dot + error snippet

---

## Project Structure

```
laterbase/
├── docker-compose.yml
├── .env.example
├── .gitignore
├── PLAN.md                        ← This file
├── backend/
│   ├── Dockerfile
│   ├── .air.toml
│   ├── go.mod
│   ├── cmd/server/main.go
│   └── internal/
│       ├── config/config.go
│       ├── db/
│       │   ├── db.go              # //go:embed migrations/*.sql
│       │   ├── links.go
│       │   ├── groups.go
│       │   └── migrations/
│       │       ├── 001_links.sql
│       │       ├── 002_groups.sql
│       │       └── 003_config.sql
│       ├── handlers/
│       │   ├── links.go
│       │   ├── groups.go
│       │   ├── stats.go
│       │   └── config.go
│       ├── models/models.go
│       ├── fetcher/
│       │   ├── fetcher.go         # interface + simple HTTP impl
│       │   └── crawl4ai.go        # Crawl4AI implementation
│       ├── categorizer/
│       │   └── categorizer.go     # gpt-4o-mini call
│       ├── workers/
│       │   └── process_link.go    # River workflow
│       └── content/
│           └── sanitizer.go       # HTML sanitization
└── frontend/
    ├── Dockerfile.dev
    ├── vite.config.ts
    ├── package.json
    ├── tsconfig.json
    ├── index.html
    └── src/
        ├── main.tsx
        ├── App.tsx
        ├── index.css              # Tailwind v4 + Inter Variable + Linear tokens
        ├── api/client.ts          # Axios instance
        ├── types.ts
        ├── hooks/
        │   ├── useLinks.ts
        │   ├── useGroups.ts
        │   ├── useStats.ts
        │   └── useConfig.ts
        ├── pages/
        │   ├── Dashboard.tsx
        │   ├── LinkDetail.tsx
        │   └── Groups.tsx
        └── components/
            ├── Layout.tsx
            ├── LinkCard.tsx
            ├── AddLinkDialog.tsx
            ├── GroupChips.tsx
            ├── ContentPreview.tsx
            ├── RegroupDropdown.tsx
            ├── StatusBadge.tsx
            ├── ColorPicker.tsx
            └── EmptyState.tsx
```

---

## Docker Compose (Orbstack-Managed)

```yaml
services:
  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_DB: laterbase
      POSTGRES_USER: laterbase
      POSTGRES_PASSWORD: laterbase_dev
    ports:
      - "5433:5432"          # avoid conflict with other PG instances
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U laterbase -d laterbase"]
      interval: 5s
      retries: 5

  api:
    build: ./backend
    ports:
      - "8080:8080"
    volumes:
      - ./backend:/app
    environment:
      DATABASE_URL: postgres://laterbase:laterbase_dev@postgres:5432/laterbase?sslmode=disable
      PORT: "8080"
      LLM_API_KEY: ${LLM_API_KEY}
      LLM_MODEL: gpt-4o-mini
      CRAWL4AI_URL: ""        # set when Crawl4AI is toggled on
    depends_on:
      postgres:
        condition: service_healthy

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.dev
    ports:
      - "5173:5173"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    environment:
      VITE_API_URL: ""        # uses Vite proxy
    depends_on:
      - api

  # Crawl4AI — started on-demand, not always running
  crawl4ai:
    image: unclecode/crawl4ai:latest
    ports:
      - "11235:11235"
    profiles:
      - crawl4ai             # only starts with: docker compose --profile crawl4ai up
    environment:
      CRAWL4AI_API_TOKEN: "" # no auth needed for local use

volumes:
  pgdata:
```

**Orbstack workflow**: The user starts/stops containers from Orbstack's GUI. `docker compose up -d` once to create the project — after that, Orbstack's UI handles start/stop/pause. The `crawl4ai` service uses a Docker Compose profile so it's NOT started by default — only when the user explicitly enables it (via dashboard or gateway command).

---

## Implementation Phases

### Phase 0: Scaffolding
- `docker-compose.yml` with all 4 services (crawl4ai profiled)
- Go module init, Dockerfile with air hot-reload
- Vite React scaffold with Tailwind v4, TanStack Query, React Router

### Phase 1: Data Layer
- Migration 001: links table (with status enum values)
- Migration 002: groups table + 6 default groups seed
- Migration 003: config table + seed `('fetcher_mode', 'simple')`
- Config package, DB connection + goose auto-migrate

### Phase 2: Core API
- Link CRUD handlers (POST/GET/GET:id/PUT/DELETE/reprocess)
- Group CRUD handlers
- Stats handler (counts by status, group, source_type)
- Config GET/PUT handler
- Health endpoint
- Route wiring in main.go

### Phase 3: Link Processing
- Fetcher interface + simple HTTP implementation (detect paywalls via DOM markers)
- Crawl4AI fetcher implementation (behind the same interface)
- GPT-4o-mini categorizer (structured JSON output)
- River ProcessLink worker (fetch → categorize → store, with honest status handling)
- River integration in main.go

### Phase 4: Frontend — Linear Dark UI
- Design token setup: Linear color palette, Inter Variable, CSS custom properties
- API client + TypeScript types
- TanStack Query hooks (links, groups, stats, config)
- Dashboard page: link list with search, group chips, status filter
- LinkCard component: favicon, title, summary, badges, status variants
- LinkDetail page: summary, content preview, actions
- Groups page: list, rename, color pick, delete
- AddLinkDialog (paste URL modal)
- EmptyState, StatusBadge, RegroupDropdown, ContentPreview components
- Layout: top nav with stats, responsive sidebar

### Phase 5: Hermes Integration
- Add `link <url>` routing to SOUL.md
- Add `link mode crawl4ai|simple` routing to SOUL.md
- Test end-to-end: WhatsApp → Hermes → API → River → Dashboard

### Phase 6: Polish
- Content preview: sanitize and render HTML
- Crawl4AI toggle: dashboard setting + gateway command
- Paywall/inaccessible detection improvements
- Responsive tweaks (desktop-first, tablet fallback)

---

## Resource Budget (Orbstack)

| Service | RAM | Notes |
|---------|-----|-------|
| PostgreSQL 17 Alpine | ~60 MB | Orbstack optimized |
| Go API + River | ~40 MB | Single binary |
| React Vite dev server | ~100 MB | HMR active |
| Crawl4AI (optional profile) | ~200 MB | Only when enabled |
| **Total (simple mode)** | **~200 MB** | Default always-on |
| **Total (with Crawl4AI)** | **~400 MB** | On-demand |

---

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Sites block bot UA | Fallback to storing URL-only with `inaccessible` status. User opens in browser |
| Paywall false positives | Simple keyword check in DOM (`paywall`, `subscribe`, `metered`). User can reprocess or open in browser |
| LLM hallucinates summary from thin context | System prompt explicitly says "set summary to empty string if content is inaccessible" |
| Crawl4AI not running when toggled | API detects connection failure, auto-falls back to simple HTTP, logs warning |
| Page content too large for LLM | Truncate to first 3000 words (~12K chars) before sending |
| Orbstack not running | Hermes responds "laterbase isn't running. Start it from Orbstack first." |

---

## Design Decision: Linear vs Notion

Two design systems are available:

| | Linear (dark) | Notion (light/warm) |
|---|---|---|
| Best for | Dashboards, data-dense UIs, tooling | Reading, writing, content-first |
| Eye strain | Low (dark canvas `#08090a`) | Moderate (white canvas) |
| Card system | Translucent surfaces, whisper borders | Warm white cards, layered shadows |
| Tag/badge system | Pill badges, indigo accent | Tinted blue pills |
| Font | Inter Variable, compressed tracking | NotionInter, warm spacing |

**Recommendation: Linear dark.** laterbase is a dashboard/tool, not a reading app. The dark canvas is better for scanning many links, and Linear's translucent surface system is purpose-built for data-dense UIs. The Notion system could work as a light-mode toggle later.

---

## Open Questions (Answered)

1. ✅ Crawl4AI or simple HTTP? → Simple by default, Crawl4AI toggleable via dashboard or `link mode crawl4ai`
2. ✅ Always-on? → Yes, Orbstack manages container lifecycle
3. ✅ LLM model? → gpt-4o-mini
4. ✅ Design system? → Linear dark (popular-web-designs template)
5. ✅ Command prefix? → No `/` — use bare `link <url>` to bypass gateway interception
6. ✅ Inaccessible content? → Never fabricate. Mark as `inaccessible`, `paywalled`, or `failed` with honest messaging
