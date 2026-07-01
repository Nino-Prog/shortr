# Shortr

A high-performance URL shortener and link analytics service built with Go.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=flat&logo=docker)

## Features

- **URL shortening** — generate short codes or use your own custom alias
- **Link analytics** — track clicks with country and city breakdown via GeoIP
- **JWT authentication** — secure register/login flow
- **Link expiration** — set an optional expiry date on any link
- **Rate limiting** — in-memory token bucket on the redirect endpoint (no Redis needed)
- **Dark UI** — clean vanilla HTML/CSS/JS dashboard, no frontend framework
- **Docker ready** — multi-stage build, single binary, ships with docker-compose

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.23 |
| Router | [chi](https://github.com/go-chi/chi) |
| Database | PostgreSQL 16 via [pgx/v5](https://github.com/jackc/pgx) |
| Auth | JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt)) |
| GeoIP | [ip-api.com](http://ip-api.com) (free, no key needed) |
| Frontend | Vanilla HTML / CSS / JS |
| Container | Docker (multi-stage alpine build) |
| CI | GitHub Actions |

## Getting Started

### Option A — Docker (recommended)

```bash
git clone https://github.com/Nino-Prog/shortr.git
cd shortr
docker compose up
```

App runs at `http://localhost:8080`. Postgres is started automatically.

### Option B — Local

**Prerequisites:** Go 1.23+, PostgreSQL 16

```bash
git clone https://github.com/Nino-Prog/shortr.git
cd shortr

# Install dependencies
go mod tidy

# Set up environment
cp .env.example .env
# Edit .env with your DATABASE_URL and JWT_SECRET

# Run migrations
psql $DATABASE_URL -f migrations/001_init.sql

# Start the server
go run ./cmd/server
```

App runs at `http://localhost:8080`.

## API Reference

### Auth

```
POST /auth/register   { "email": "...", "password": "..." }
POST /auth/login      { "email": "...", "password": "..." }
```

Both return `{ "token": "<jwt>", "user": { ... } }`.

### Links (requires `Authorization: Bearer <token>`)

```
POST   /api/shorten          { "url": "https://...", "code": "optional", "expires_at": "2025-01-01T00:00:00Z" }
GET    /api/links            → list all your links
DELETE /api/links/:code      → delete a link
GET    /api/analytics/:code  → click analytics for a link
```

### Redirect

```
GET /:code   → 302 redirect + async click recording
```

## Project Structure

```
shortr/
├── cmd/server/main.go          # Entry point, router setup
├── internal/
│   ├── handler/
│   │   ├── auth.go             # Register, login, JWT middleware
│   │   ├── links.go            # Shorten, redirect, list, delete, analytics
│   │   └── ratelimit.go        # In-memory token bucket rate limiter
│   ├── store/store.go          # All database queries (pgx/v5)
│   ├── model/model.go          # Domain types: User, Link, Click, Analytics
│   └── geo/geo.go              # GeoIP via ip-api.com
├── migrations/001_init.sql     # DB schema (users, links, clicks)
├── web/
│   ├── templates/              # index.html, dashboard.html
│   └── static/                 # CSS + vanilla JS
├── Dockerfile                  # Multi-stage alpine build
├── docker-compose.yml          # App + Postgres
└── .github/workflows/ci.yml    # go test + docker build on push to main
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | — (required) |
| `JWT_SECRET` | Secret key for signing JWTs | `dev-secret-change-me` |
| `PORT` | HTTP port | `8080` |

## CI/CD

GitHub Actions runs on every push to `main`:
1. Spins up a Postgres service container
2. Runs migrations
3. Runs `go test ./...`
4. Builds the Docker image

## License

MIT
