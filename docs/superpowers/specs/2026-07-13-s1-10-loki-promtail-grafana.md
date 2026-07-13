# S1-10 — Loki + Promtail + Grafana Design

> **Date:** 2026-07-13

## Overview

Add log aggregation and visualization to RitualX dev stack. Backend already emits structured JSON logs to stdout via `slog`. Promtail collects container logs and ships to Loki. Grafana queries Loki for visualization and alerting.

## Goals

- Aggregate all container logs in Loki
- Ship backend structured JSON logs (with `trace_id`, `level`, `msg`) via Promtail
- Visualize logs in Grafana with a pre-provisioned datasource and dashboard
- Zero backend code changes required

## Non-Goals

- Production-grade log retention / replication (HA Loki)
- Metrics (Prometheus/Grafana Mimir) — separate future sprint
- Log-based alerting rules (future sprint)

## Approach

**Single docker-compose addition** — add `loki`, `promtail`, `grafana` services to the root `docker-compose.yml`. Configs stored in `infra/loki/`, `infra/promtail/`, `infra/grafana/`.

## Design Details

### Directory Layout

```
infra/
  loki/
    config.yml           # Loki server config (filesystem storage)
  promtail/
    config.yml           # Promtail scrape config (Docker logs)
  grafana/
    provisioning/
      datasources/
        loki.yml         # Auto-provision Loki datasource
      dashboards/
        dashboard.yml    # Dashboard provider config
    dashboards/
      ritualx.json       # Pre-built RitualX logs dashboard
```

### Loki Config

- Storage: filesystem (`/loki/chunks`, `/loki/index`) via Docker volume `loki_data`
- Retention: 7 days (`retention_period: 168h`)
- HTTP listen: `:3100`
- No auth (dev only)

### Promtail Config

- Scrapes Docker container logs via `docker_sd_configs` (discovery via Docker socket)
- Relabels: `container_name` → label `container`, `service_name` → label `service`
- Drops logs from `loki`, `promtail`, `grafana` containers to avoid noise
- Pipeline stage: `json` parser → extract `level`, `trace_id`, `msg` as labels

### Grafana Config

- Version: `grafana/grafana:10.4.3`
- Admin creds: `admin` / `ritualx_dev` (env vars `GF_SECURITY_ADMIN_PASSWORD`)
- Auto-provisioned datasource: Loki at `http://loki:3100`
- Pre-built dashboard: log stream panel + level filter variable

### docker-compose.yml Changes

Add to existing `docker-compose.yml`:

```yaml
loki:
  image: grafana/loki:3.0.0
  container_name: ritualx-loki
  volumes:
    - ./infra/loki/config.yml:/etc/loki/config.yml
    - loki_data:/loki
  command: -config.file=/etc/loki/config.yml
  ports:
    - "3100:3100"
  networks:
    - ritualx-net

promtail:
  image: grafana/promtail:3.0.0
  container_name: ritualx-promtail
  volumes:
    - ./infra/promtail/config.yml:/etc/promtail/config.yml
    - /var/lib/docker/containers:/var/lib/docker/containers:ro
    - /var/run/docker.sock:/var/run/docker.sock
  command: -config.file=/etc/promtail/config.yml
  depends_on:
    - loki
  networks:
    - ritualx-net

grafana:
  image: grafana/grafana:10.4.3
  container_name: ritualx-grafana
  environment:
    GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD:-ritualx_dev}
    GF_USERS_ALLOW_SIGN_UP: "false"
  volumes:
    - ./infra/grafana/provisioning:/etc/grafana/provisioning
    - ./infra/grafana/dashboards:/var/lib/grafana/dashboards
    - grafana_data:/var/lib/grafana
  ports:
    - "3001:3000"
  depends_on:
    - loki
  networks:
    - ritualx-net
```

Add volumes: `loki_data`, `grafana_data`

### Ports

| Service  | Port  | Purpose            |
|----------|-------|--------------------|
| Loki     | 3100  | Push/query API     |
| Grafana  | 3001  | Web UI (host)      |

### Env Vars

Add to `.env.example`:
```
GRAFANA_PASSWORD=ritualx_dev
```

## Error Cases

- Docker socket not available → Promtail fails to discover containers; fallback: use `positions.yaml` + static file path
- Loki unreachable → Promtail retries with backoff (built-in)
- Grafana first-start provisioning error → logs in `ritualx-grafana` container

## Dependencies

- S1-09 (Docker Compose) ✅ done — this extends it
- Docker socket accessible on host (Linux/Mac native; Windows via Docker Desktop WSL2)

## Open Questions

- None — scope is clear
