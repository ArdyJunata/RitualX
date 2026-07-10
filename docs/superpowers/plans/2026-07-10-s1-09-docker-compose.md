# S1-09 Docker Compose (All Services) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Single `docker-compose.yml` at repo root that starts postgres, backend (Go Fiber), and frontend (Next.js) with correct networking, env injection, and health checks.

**Architecture:** Root-level `docker-compose.yml` orchestrates 3 services + 1 one-shot migrate service. Backend and frontend each get their own `Dockerfile`. Migrations run as a one-shot `migrate` init container that exits before backend starts. Dev-only DB compose (`backend/docker-compose.dev.yml`) stays untouched.

**Tech Stack:** Docker Compose v2, Go 1.25 multi-stage build, Node 20 Alpine, golang-migrate CLI, postgres:16-alpine

## Global Constraints

- Docker Compose file version: Compose Spec (no `version:` key)
- Backend binary built with `CGO_ENABLED=0 GOOS=linux`
- Frontend runs `next start` on port 3000 (production/standalone mode)
- Backend runs on port 8080
- Postgres exposed on 5432 for dev convenience
- All secrets via env vars; `.env` file at repo root (gitignored)
- `migrate` binary fetched in backend builder stage — no extra image needed
- Network name: `ritualx-net`
- Volume name: `postgres_data`

---

## File Map

| Action | File |
|--------|------|
| Create | `docker-compose.yml` (repo root) |
| Create | `backend/Dockerfile` |
| Create | `frontend/Dockerfile` |
| Create | `.env.example` (repo root) |
| Modify | `frontend/next.config.mjs` — add `output: "standalone"` |
| Modify | `.gitignore` (repo root) — add `.env` |

---

## Task 1: Backend Dockerfile

**Files:**
- Create: `backend/Dockerfile`

**Interfaces:**
- Produces: Docker image exposing port `8080`
- Entry point: `/app/server`
- Env reads: `APP_PORT`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `LOG_LEVEL`
- Also ships: `/usr/local/bin/migrate` + `/app/migrations/` (used by migrate service)

- [ ] **Step 1: Create `backend/Dockerfile`**

```dockerfile
# ── builder ──────────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache curl ca-certificates && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz \
      | tar xvz -C /usr/local/bin migrate

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# ── runtime ──────────────────────────────────────────────────────────────────
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["/app/server"]
```

> Note: `wget` is included because the backend healthcheck in compose uses `wget`.

- [ ] **Step 2: Verify Dockerfile syntax locally**

```bash
docker build -t ritualx-backend ./backend --no-cache --progress=plain
```

Expected: Build completes with no errors. Image `ritualx-backend` listed in `docker images`.

---

## Task 2: Frontend Dockerfile + Next.js Standalone

**Files:**
- Modify: `frontend/next.config.mjs`
- Create: `frontend/Dockerfile`

**Interfaces:**
- Produces: Docker image exposing port `3000`
- Entry point: `node server.js` (Next.js standalone output)
- Env reads at runtime: `NEXT_PUBLIC_API_URL`

- [ ] **Step 1: Enable standalone output in `frontend/next.config.mjs`**

Replace entire file content with:

```js
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
};

export default nextConfig;
```

- [ ] **Step 2: Create `frontend/Dockerfile`**

```dockerfile
# ── deps ─────────────────────────────────────────────────────────────────────
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# ── builder ──────────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

# ── runtime ──────────────────────────────────────────────────────────────────
FROM node:20-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000
CMD ["node", "server.js"]
```

- [ ] **Step 3: Verify build**

```bash
docker build -t ritualx-frontend ./frontend --no-cache --progress=plain
```

Expected: Build completes. Image `ritualx-frontend` listed in `docker images`.

---

## Task 3: Root `.env.example` and `.gitignore`

**Files:**
- Create: `.env.example` (repo root)
- Modify: `.gitignore` (repo root)

- [ ] **Step 1: Create `.env.example` at repo root**

```dotenv
# Postgres
DB_USER=ritualx
DB_PASSWORD=ritualx_dev
DB_NAME=ritualx

# Backend
APP_PORT=8080
JWT_SECRET=change-me-in-production
LOG_LEVEL=info

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080
```

- [ ] **Step 2: Add `.env` to repo root `.gitignore`**

Append to the end of `.gitignore` at repo root:

```
# local secrets
.env
```

- [ ] **Step 3: Create `.env` from example (developer does once)**

```bash
cp .env.example .env
```

---

## Task 4: Root `docker-compose.yml`

**Files:**
- Create: `docker-compose.yml` (repo root)

**Interfaces:**
- Consumes: `./backend` build context → backend + migrate services
- Consumes: `./frontend` build context → frontend service
- Consumes: `.env` at repo root
- Produces: Stack accessible at `http://localhost:3000` (frontend), `http://localhost:8080` (backend)

**Dependency chain:** `postgres` (healthy) → `migrate` (completed) → `backend` (healthy) → `frontend`

- [ ] **Step 1: Create `docker-compose.yml` at repo root**

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: ritualx-postgres
    environment:
      POSTGRES_USER: ${DB_USER:-ritualx}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-ritualx_dev}
      POSTGRES_DB: ${DB_NAME:-ritualx}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - ritualx-net
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-ritualx} -d ${DB_NAME:-ritualx}"]
      interval: 5s
      timeout: 5s
      retries: 10

  migrate:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: ritualx-migrate
    entrypoint: >
      sh -c "migrate -path /app/migrations -database
      'postgres://${DB_USER:-ritualx}:${DB_PASSWORD:-ritualx_dev}@postgres:5432/${DB_NAME:-ritualx}?sslmode=disable'
      up"
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - ritualx-net
    restart: "no"

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: ritualx-backend
    environment:
      APP_PORT: ${APP_PORT:-8080}
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${DB_USER:-ritualx}
      DB_PASSWORD: ${DB_PASSWORD:-ritualx_dev}
      DB_NAME: ${DB_NAME:-ritualx}
      JWT_SECRET: ${JWT_SECRET:-change-me-in-production}
      LOG_LEVEL: ${LOG_LEVEL:-info}
    ports:
      - "8080:8080"
    depends_on:
      migrate:
        condition: service_completed_successfully
    networks:
      - ritualx-net
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://localhost:8080/api/v1/health || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: ritualx-frontend
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080}
    ports:
      - "3000:3000"
    depends_on:
      backend:
        condition: service_healthy
    networks:
      - ritualx-net
    restart: unless-stopped

networks:
  ritualx-net:
    driver: bridge

volumes:
  postgres_data:
```

- [ ] **Step 2: Validate compose file**

```bash
docker compose config
```

Expected: Prints resolved config with no errors, no warnings.

---

## Task 5: End-to-End Smoke Test

**Files:** none

- [ ] **Step 1: Bring up full stack**

```bash
docker compose up --build -d
```

Expected: 4 containers started. `ritualx-migrate` exits with code 0.

- [ ] **Step 2: Watch logs until stable**

```bash
docker compose logs -f --tail=50
```

Expected: Backend logs `"database connected"` and `"starting server"`. No error lines.

- [ ] **Step 3: Check backend health endpoint**

```bash
curl http://localhost:8080/api/v1/health
```

Expected: HTTP 200 with JSON body like `{"status":"ok"}`.

- [ ] **Step 4: Check frontend responds**

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000
```

Expected: `200`

- [ ] **Step 5: Tear down**

```bash
docker compose down
```

Expected: All containers stopped and removed. `postgres_data` volume persists (verify with `docker volume ls`).

---

## Self-Review

- [x] Postgres healthcheck gates migrate; migrate `service_completed_successfully` gates backend; backend healthcheck gates frontend — correct dependency chain
- [x] `DB_HOST=postgres` (compose service name), not `localhost`
- [x] `next start` requires `output: "standalone"` → covered Task 2 Step 1
- [x] `migrate` CLI bundled inside backend image — no separate image pull
- [x] `wget` installed in backend runtime image for healthcheck
- [x] `.env` gitignored, `.env.example` committed
- [x] All steps have exact commands and code — no placeholders
- [x] `backend/docker-compose.dev.yml` untouched
