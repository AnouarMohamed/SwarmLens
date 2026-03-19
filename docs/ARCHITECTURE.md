# Architecture

SwarmLens is a three-service architecture with optional external integrations.

## High-level components

1. Frontend (`src/`)

- React + Vite + Tailwind
- Zustand stores for cluster, diagnostics, incidents, and ops intelligence
- Native ECharts visualizations with optional Grafana embed mode

2. Backend (`backend/`)

- Go HTTP API (`cmd/swarmlens`)
- Reads Docker Swarm inventory through Docker Engine API
- Runs deterministic diagnostics plugin engine
- Aggregates ops metrics, risk, and insights
- Exposes action orchestrator and assistant SSE

3. Predictor (`predictor/`)

- FastAPI service (`/score`)
- Computes risk score and confidence from snapshot-derived signals
- Authenticated with shared secret header

## Backend package map

- `internal/config`: env loading and validation
- `internal/docker`: live inventory adapters + demo fixtures
- `internal/state`: snapshot cache, freshness, ops history points
- `internal/intelligence`: diagnostics engine and plugins
- `internal/httpapi`: router, middleware, handlers
- `internal/predictor`: predictor client with deterministic fallback
- `internal/auth`: principal extraction and role checks
- `internal/audit`: in-memory append-only audit store
- `internal/incident`: incident lifecycle store
- `internal/stream`: SSE event bus
- `internal/model`: shared domain/API types

## Request flow

1. Browser calls `/api/v1/*`
2. Middleware chain applies:
   - recovery
   - structured logging
   - CORS
   - security headers
   - rate limiting
3. Auth middleware validates bearer token if enabled
4. Handlers fetch/refresh snapshot cache as needed
5. Diagnostics, risk, and insights are computed
6. Response is returned in JSON envelope (or SSE for stream endpoints)

## Freshness model

Cache status is propagated as:

- `live`
- `stale`
- `disconnected`

Transitions are based on snapshot refresh success/failure and `SNAPSHOT_STALE_SECONDS`.

## Action model

All operations funnel through `POST /api/v1/actions/execute` and resource-specific action endpoints.

Outcomes are structured:

- `status`: `success`, `failed`, `dry_run`, `blocked`
- `mode`: `live`, `dry_run`, `demo`, `blocked`
- `executed`: boolean
- `plan`: dry-run steps when mutation is not executed
- `auditID`: audit reference for traceability

## Assistant and insights

- Deterministic insights are always available (`/ops/insights`).
- Optional OpenAI narrative augmentation is used when configured.
- `/assistant/chat` streams context, insight object, hypotheses, actions, and final message via SSE.

## Observability and charting

- Backend emits runtime and posture telemetry (`/metrics`, `/ops/metrics`).
- Frontend charts can render either:
  - native ECharts from SwarmLens APIs
  - Grafana embeds from configured dashboard UID

## Demo mode design

In `APP_MODE=demo`:

- backend seeds deterministic snapshots/events
- frontend can force scenario profiles via query param (`healthy`, `degraded`, `incident-burst`, `recovery`, `disconnected`)
- actions execute through same orchestrator path but remain safe (simulated or dry-run)

## Deployment topology

Supported deployment paths:

- Local compose (`docker-compose.yml`)
- Docker Swarm overlays (`deploy/overlays/dev|demo|prod/stack.yml`)
- VPS automation scripts (`scripts/preflight-prod.sh`, `deploy.sh`, `rollback.sh`)
