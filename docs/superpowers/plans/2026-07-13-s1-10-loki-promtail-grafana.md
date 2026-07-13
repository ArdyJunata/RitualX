# S1-10 Loki + Promtail + Grafana Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Loki + Promtail + Grafana to the dev Docker Compose stack so backend JSON logs are collected and visualized with zero backend code changes.

**Architecture:** Promtail scrapes Docker container stdout via the Docker socket, ships logs to Loki (filesystem storage), Grafana auto-provisions Loki as datasource and loads a pre-built dashboard. All services join the existing `ritualx-net` network.

**Tech Stack:** Loki 3.0.0, Promtail 3.0.0, Grafana 10.4.3, Docker Compose, YAML configs.

## Global Constraints

- No backend Go code changes
- Windows host (Docker Desktop WSL2) — Docker socket path: `/var/run/docker.sock`
- All config files use YAML, no TOML
- Grafana on port `3001` (not 3000 — conflicts with frontend)
- Loki on port `3100`
- Branch: `feat/s1-10-loki-promtail-grafana`
- No commits until explicitly instructed

---

### Task 1: Create feature branch

**Files:** none

- [ ] **Step 1: Create branch**

```powershell
git checkout -b feat/s1-10-loki-promtail-grafana
```

Expected: `Switched to a new branch 'feat/s1-10-loki-promtail-grafana'`

---

### Task 2: Loki config

**Files:**
- Create: `infra/loki/config.yml`

- [ ] **Step 1: Create directories**

```powershell
New-Item -ItemType Directory -Force -Path "infra/loki"
```

- [ ] **Step 2: Create `infra/loki/config.yml`**

```yaml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096
  log_level: warn

common:
  instance_addr: 127.0.0.1
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

query_range:
  results_cache:
    cache:
      embedded_cache:
        enabled: true
        max_size_mb: 100

schema_config:
  configs:
    - from: 2020-10-24
      store: tsdb
      object_store: filesystem
      schema: v13
      index:
        prefix: index_
        period: 24h

ruler:
  alertmanager_url: http://localhost:9093

compactor:
  working_directory: /loki/compactor
  retention_enabled: true
  delete_request_store: filesystem

limits_config:
  retention_period: 168h
```

- [ ] **Step 3: Verify file created**

```powershell
Get-Content infra/loki/config.yml | head -5
```

Expected: first 5 lines of the config

---

### Task 3: Promtail config

**Files:**
- Create: `infra/promtail/config.yml`

- [ ] **Step 1: Create directory**

```powershell
New-Item -ItemType Directory -Force -Path "infra/promtail"
```

- [ ] **Step 2: Create `infra/promtail/config.yml`**

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
        filters:
          - name: label
            values: ["com.docker.compose.project=ritualx"]
    relabel_configs:
      - source_labels: [__meta_docker_container_name]
        regex: /(.*)
        target_label: container
      - source_labels: [__meta_docker_compose_service]
        target_label: service
      - source_labels: [container]
        regex: ritualx-(loki|promtail|grafana)
        action: drop
    pipeline_stages:
      - json:
          expressions:
            level: level
            trace_id: trace_id
            msg: msg
      - labels:
          level:
          trace_id:
```

- [ ] **Step 3: Verify file created**

```powershell
Get-Content infra/promtail/config.yml | Select-Object -First 5
```

Expected: first 5 lines of the config

---

### Task 4: Grafana provisioning — datasource

**Files:**
- Create: `infra/grafana/provisioning/datasources/loki.yml`

- [ ] **Step 1: Create directories**

```powershell
New-Item -ItemType Directory -Force -Path "infra/grafana/provisioning/datasources"
New-Item -ItemType Directory -Force -Path "infra/grafana/provisioning/dashboards"
New-Item -ItemType Directory -Force -Path "infra/grafana/dashboards"
```

- [ ] **Step 2: Create `infra/grafana/provisioning/datasources/loki.yml`**

```yaml
apiVersion: 1

datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    isDefault: true
    editable: false
    jsonData:
      maxLines: 1000
```

- [ ] **Step 3: Verify**

```powershell
Get-Content infra/grafana/provisioning/datasources/loki.yml
```

Expected: full file content as written

---

### Task 5: Grafana provisioning — dashboard provider

**Files:**
- Create: `infra/grafana/provisioning/dashboards/dashboard.yml`

- [ ] **Step 1: Create `infra/grafana/provisioning/dashboards/dashboard.yml`**

```yaml
apiVersion: 1

providers:
  - name: RitualX
    orgId: 1
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
      foldersFromFilesStructure: false
```

---

### Task 6: Grafana dashboard JSON

**Files:**
- Create: `infra/grafana/dashboards/ritualx.json`

- [ ] **Step 1: Create `infra/grafana/dashboards/ritualx.json`**

```json
{
  "id": null,
  "uid": "ritualx-logs",
  "title": "RitualX — Logs",
  "tags": ["ritualx", "logs"],
  "timezone": "browser",
  "schemaVersion": 39,
  "version": 1,
  "refresh": "10s",
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "templating": {
    "list": [
      {
        "name": "service",
        "type": "query",
        "label": "Service",
        "datasource": { "type": "loki", "uid": "loki" },
        "query": "label_values(service)",
        "refresh": 2,
        "includeAll": true,
        "allValue": ".*",
        "multi": true,
        "current": {}
      },
      {
        "name": "level",
        "type": "custom",
        "label": "Level",
        "query": "debug,info,warn,error",
        "includeAll": true,
        "allValue": ".*",
        "multi": true,
        "current": {}
      }
    ]
  },
  "panels": [
    {
      "id": 1,
      "type": "logs",
      "title": "Application Logs",
      "gridPos": { "h": 24, "w": 24, "x": 0, "y": 0 },
      "datasource": { "type": "loki", "uid": "loki" },
      "targets": [
        {
          "expr": "{service=~\"$service\", level=~\"$level\"}",
          "refId": "A",
          "legendFormat": "{{container}}"
        }
      ],
      "options": {
        "dedupStrategy": "none",
        "enableLogDetails": true,
        "prettifyLogMessage": true,
        "showTime": true,
        "sortOrder": "Descending",
        "wrapLogMessage": true
      }
    }
  ]
}
```

---

### Task 7: Update docker-compose.yml

**Files:**
- Modify: `docker-compose.yml`

- [ ] **Step 1: Add loki, promtail, grafana services**

Open `docker-compose.yml`. After the `frontend` service block (before `networks:`), add:

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
    restart: unless-stopped

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
    restart: unless-stopped

  grafana:
    image: grafana/grafana:10.4.3
    container_name: ritualx-grafana
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD:-ritualx_dev}
      GF_USERS_ALLOW_SIGN_UP: "false"
      GF_AUTH_ANONYMOUS_ENABLED: "false"
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
    restart: unless-stopped
```

- [ ] **Step 2: Add volumes `loki_data` and `grafana_data`**

In the `volumes:` section at the bottom, add:

```yaml
  loki_data:
  grafana_data:
```

- [ ] **Step 3: Verify docker-compose syntax**

```powershell
docker compose config --quiet
```

Expected: no errors (silent exit 0)

---

### Task 8: Update .env.example

**Files:**
- Modify: `.env.example`

- [ ] **Step 1: Append GRAFANA_PASSWORD to `.env.example`**

Open `.env.example`, add at the bottom:

```
GRAFANA_PASSWORD=ritualx_dev
```

---

### Task 9: Smoke test — start observability stack

- [ ] **Step 1: Pull images**

```powershell
docker compose pull loki promtail grafana
```

Expected: all 3 images pulled without error

- [ ] **Step 2: Start loki + grafana only (promtail needs backend logs)**

```powershell
docker compose up -d loki grafana
```

Expected: both containers start healthy

- [ ] **Step 3: Verify Loki is ready**

```powershell
Start-Sleep -Seconds 5
Invoke-RestMethod -Uri "http://localhost:3100/ready"
```

Expected: response body `ready`

- [ ] **Step 4: Verify Grafana UI**

Open browser → `http://localhost:3001`
Login: `admin` / `ritualx_dev`
Navigate to Connections → Data Sources → confirm `Loki` datasource exists.
Navigate to Dashboards → confirm `RitualX — Logs` dashboard exists.

- [ ] **Step 5: Start promtail + full stack**

```powershell
docker compose up -d
```

Expected: all services running

- [ ] **Step 6: Verify logs flow**

In Grafana → Explore → select Loki datasource → run query:
```
{service="backend"}
```
Expected: backend JSON log lines visible

---

### Task 10: Update CHECKPOINT.md

**Files:**
- Modify: `.gemini/CHECKPOINT.md`

- [ ] **Step 1: Mark S1-10 as done**

Change:
```
| S1-10 | Loki + Promtail + Grafana | ⬜ Pending | — |
```
To:
```
| S1-10 | Loki + Promtail + Grafana | ✅ Done | — |
```

- [ ] **Step 2: Update "What To Do Next" section**

Replace S1-10 priority entry with Sprint 2 planning note.
