# URL Shortener — Go + Next.js

A production-grade URL shortener with JWT authentication, real-time analytics, QR codes, and an interactive 3D globe visualization.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22+, Fiber v2, PostgreSQL, Redis, sqlc |
| Frontend | Next.js 16, TypeScript, Tailwind CSS 4, TanStack Query, react-globe.gl |
| Infra | Docker Compose, Nginx, golang-migrate |

## Features

- **Auth** — Register/login with JWT (access 15 min + refresh 7 d, httpOnly cookie)
- **Shorten & Redirect** — Base62 slugs, custom aliases, link expiry, Redis caching with singleflight
- **Analytics** — Country, device, browser, referrer tracking; 3D globe visualization of click distribution
- **Rate Limiting** — Sliding window per IP via Redis
- **QR Codes** — Per-link PNG generation with configurable size
- **Dashboard** — Aggregate stats, per-link charts, globe view, link management

## Quick Start (Development)

```bash
git clone https://github.com/YOUR_USER/short-url-generator.git
cd short-url-generator

# Copy env
cp backend/.env.example backend/.env
# Edit backend/.env with your secrets (JWT_ACCESS_SECRET, JWT_REFRESH_SECRET, etc.)

# Start all services
docker compose -f docker-compose.dev.yml up -d

# Run migrations (first time only)
docker compose -f docker-compose.dev.yml run --rm migrate

# App is live at http://localhost
```

## Production Deployment

```bash
# Copy env and set production values
cp backend/.env.example backend/.env
# ── Change these for production ──
# ENV=production
# JWT_ACCESS_SECRET=<strong-random-secret>
# JWT_REFRESH_SECRET=<strong-random-secret>
# DB_PASSWORD=<strong-db-password>
# REDIS_PASSWORD=<strong-redis-password>
# ALLOWED_ORIGINS=https://yourdomain.com

# Build and start all services
docker compose -f docker-compose.prod.yml up -d --build

# Run migrations
docker compose -f docker-compose.prod.yml run --rm migrate

# App is live at http://localhost (port 80)
```

To tear down:

```bash
# Dev
docker compose -f docker-compose.dev.yml down

# Prod (preserves data volumes)
docker compose -f docker-compose.prod.yml down

# Prod (wipe everything including data)
docker compose -f docker-compose.prod.yml down -v
```

## Database Migrations

Migrations use [golang-migrate](https://github.com/golang-migrate/migrate). All migration files live in `backend/db/migrations/`.

```bash
# ── With Docker Compose (recommended) ──

# Apply all pending migrations (up)
docker compose -f docker-compose.dev.yml run --rm migrate

# Rollback the last migration (down)
docker compose -f docker-compose.dev.yml run --rm migrate -- sh -c \
  'migrate -path=/migrations/ -database "postgres://$${DB_USER}:$${DB_PASSWORD}@db:$${DB_PORT}/$${DB_NAME}?sslmode=disable" down 1'

# Rollback all migrations (full reset)
docker compose -f docker-compose.dev.yml run --rm migrate -- sh -c \
  'migrate -path=/migrations/ -database "postgres://$${DB_USER}:$${DB_PASSWORD}@db:$${DB_PORT}/$${DB_NAME}?sslmode=disable" down -all'



```

Current migrations:

| # | Name | Description |
|---|------|-------------|
| 1 | `create_users` | Users table (id, email, password, created_at) |
| 2 | `create_urls` | URLs table (slug, original, custom, expires_at, user_id FK) |
| 3 | `create_clicks` | Clicks table (url_id FK, ip_hash, country, city, device, browser, referrer, clicked_at) |

## API Endpoints

### Auth (Public)
```
POST /api/auth/register    Register
POST /api/auth/login       Login → access token + refresh cookie
POST /api/auth/refresh     Refresh access token
POST /api/auth/logout      Clear refresh cookie
```

### Public
```
GET /:slug                  Redirect to original URL
GET /api/links/:slug/qr     QR code PNG
```

### Authenticated
```
POST   /api/links                Create short URL
GET    /api/links                List user's links
GET    /api/links/stats/aggregate  Aggregate stats across all links
GET    /api/links/:slug          Link detail
PATCH  /api/links/:slug          Update alias or expiry
DELETE /api/links/:slug          Delete link
GET    /api/links/:slug/stats    Per-link click statistics
```

## Project Structure

```
├── backend/
│   ├── cmd/server/main.go          # Entry point, wiring
│   ├── internal/
│   │   ├── analytics/worker.go     # Async click ingestion worker pool
│   │   ├── cache/redis.go           # Redis cache layer
│   │   ├── config/config.go         # Viper config
│   │   ├── middleware/              # JWT auth, rate limiter, logger
│   │   ├── modules/
│   │   │   ├── auth/               # Register, login, refresh, logout
│   │   │   ├── links/              # CRUD, stats, aggregate stats, QR
│   │   │   └── redirect/          # Public redirect + analytics enqueue
│   │   └── repository/             # sqlc-generated typed queries
│   ├── pkg/                         # slug gen, JWT, response, logger, validator
│   ├── db/migrations/               # golang-migrate SQL files
│   └── sqlc.yaml
├── frontend/
│   ├── src/
│   │   ├── app/                    # Next.js App Router pages
│   │   ├── components/
│   │   │   ├── links/              # LinkTable, CreateLinkForm, DashboardGlobe, LinkStatCharts
│   │   │   ├── landing/            # Hero, Features, Navbar, etc.
│   │   │   └── ui/                 # Loading, GlobeView
│   │   ├── hooks/                   # useAuth, useLinks, useLinkStats, useAggregateStats
│   │   ├── lib/                     # api.ts (axios), countries.ts, validators.ts
│   │   ├── store/                   # Zustand stores
│   │   └── types/                   # TypeScript interfaces
│   └── public/                      # Globe textures (earth-dark.jpg, earth-topology.png)
├── nginx/
│   ├── dev.conf                     # Dev reverse proxy config
│   └── prod.conf                    # Prod reverse proxy (with gzip, caching)
├── docker-compose.dev.yml
├── docker-compose.prod.yml
├── .gitignore
├── .dockerignore
└── README.md
```

## Running Tests

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && bun run lint
```

## Environment Variables

See `backend/.env.example` for the full list. Key variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | API port | `8080` |
| `ENV` | Environment | `development` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `JWT_ACCESS_SECRET` | Access token secret | — |
| `JWT_REFRESH_SECRET` | Refresh token secret | — |
| `BASE_URL` | Public URL for short links | `http://localhost:8080` |
| `RATE_LIMIT_REDIRECT` | Max redirects per minute per IP | `60` |
| `RATE_LIMIT_CREATE` | Max link creates per minute per IP | `10` |
| `GEOIP_DB_PATH` | MaxMind GeoLite2 path | `./GeoLite2-City.mmdb` |
| `ALLOWED_ORIGINS` | CORS origins | `http://localhost:3000` |

## License

MIT