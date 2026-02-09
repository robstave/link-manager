# Link Manager — Deployment

## Docker Compose Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Docker Compose                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   postgres  │  │    goose    │  │        api          │  │
│  │  (pgvector) │  │ (migrations)│  │       (Go)          │  │
│  │   :5432     │  │  one-shot   │  │      :8080          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│         │                │                    │             │
│         └────────────────┴────────────────────┘             │
│                          │                                  │
│                    volumes:                                 │
│                    - pgdata                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Services

### postgres

```yaml
postgres:
  image: pgvector/pgvector:pg17-bookworm
  environment:
    POSTGRES_USER: linkman
    POSTGRES_PASSWORD: ${DB_PASSWORD}
    POSTGRES_DB: linkman
  volumes:
    - pgdata:/var/lib/postgresql/data
  ports:
    - "5432:5432"
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U linkman"]
    interval: 5s
    timeout: 5s
    retries: 5
```

### goose (migrations)

```yaml
goose:
  build:
    context: .
    dockerfile: Dockerfile.goose
  depends_on:
    postgres:
      condition: service_healthy
  environment:
    GOOSE_DRIVER: postgres
    GOOSE_DBSTRING: postgres://linkman:${DB_PASSWORD}@postgres:5432/linkman?sslmode=disable
  command: ["goose", "-dir", "/migrations", "up"]
  volumes:
    - ./migrations:/migrations:ro
```

### api

```yaml
api:
  build:
    context: .
    dockerfile: Dockerfile
  depends_on:
    goose:
      condition: service_completed_successfully
  environment:
    DATABASE_URL: postgres://linkman:${DB_PASSWORD}@postgres:5432/linkman?sslmode=disable
    JWT_SECRET: ${JWT_SECRET}
    ADMIN_USERNAME: ${ADMIN_USERNAME}
    ADMIN_PASSWORD: ${ADMIN_PASSWORD}
  ports:
    - "8080:8080"
```

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| DB_PASSWORD | Yes | PostgreSQL password |
| JWT_SECRET | Yes | Secret for JWT signing |
| ADMIN_USERNAME | Yes | Initial admin username |
| ADMIN_PASSWORD | Yes | Initial admin password |
| LLM_API_KEY | No | API key for generated notes (V2) |

Example `.env`:
```
DB_PASSWORD=supersecret
JWT_SECRET=anothersecret
ADMIN_USERNAME=admin
ADMIN_PASSWORD=adminpass
```

---

## Startup Flow

1. `postgres` starts and becomes healthy
2. `goose` runs all pending migrations
3. `goose` exits with success
4. `api` starts
5. `api` checks if admin user exists; if not, creates from env vars
6. `api` serves on :8080

---

## Production Considerations

### Reverse Proxy (nginx/Caddy)

```
                    ┌─────────────┐
  HTTPS:443  ───────│   Caddy     │────── HTTP:8080 ────► api
                    │   (TLS)     │
                    └─────────────┘
```

### Backups

```bash
# Backup
docker compose exec postgres pg_dump -U linkman linkman > backup.sql

# Restore
docker compose exec -T postgres psql -U linkman linkman < backup.sql
```

### Volumes

- `pgdata`: Persistent PostgreSQL data
- Mount `./migrations` read-only to goose container

---

## Development

```bash
# Start all services
docker compose up --build

# Run only postgres for local dev
docker compose up postgres

# Connect to DB
docker compose exec postgres psql -U linkman linkman

# Reset everything
docker compose down -v
docker compose up --build
```

---

## Milestones

| # | Milestone | Acceptance Criteria |
|---|-----------|---------------------|
| 1 | Compose runs | `docker compose up --build` starts all services |
| 2 | Migrations | All tables created, admin user seeded |
| 3 | Auth | Login/logout works, JWT validated |
| 4 | CRUD | Projects, categories, links, tags CRUD |
| 5 | UI | Project view with category cards working |
| 6 | Search | Full-text search returns ranked results |
| 7 | Click tracking | Clicks recorded, counts displayed |
| 8 | Cart + Export | Cart toggle, JSON/Obsidian export |
| 9 | Generated notes | LLM integration (stub OK for V1) |
| 10 | Vectors | pgvector semantic search (V2) |
