# SwarmLens

SwarmLens is an operations console for Docker Swarm clusters.

It combines deterministic diagnostics, operational telemetry, incident workflows, and AI-assisted triage in one product-oriented UI.

## What you get

- Live or demo inventory of cluster, nodes, stacks, services, tasks, networks, volumes, secrets, and configs.
- Deterministic diagnostics engine with plugin-based findings and evidence.
- Operational telemetry and risk trends (`/api/v1/ops/metrics`) with service risk hot spots.
- Hybrid insights (`/api/v1/ops/insights`): deterministic baseline plus optional OpenAI narrative layer.
- Action orchestrator (`/api/v1/actions/execute`) with structured outcomes and audit IDs.
- Incident lifecycle (`/api/v1/incidents`) and append-only audit log (`/api/v1/audit`).
- Assistant SSE stream (`/api/v1/assistant/chat`) with hypotheses and recommended actions.
- Grafana panel embedding when frontend `VITE_GRAFANA_*` vars are configured.

## Runtime modes

| Mode   | Purpose                                    | Auth           | Mutations            |
| ------ | ------------------------------------------ | -------------- | -------------------- |
| `demo` | Product demo without external dependencies | Off by default | Simulated or dry-run |
| `dev`  | Local development against real Swarm       | Optional       | Policy controlled    |
| `prod` | Production deployment                      | Required       | Policy controlled    |

Important defaults:

- `APP_MODE=demo`
- `WRITE_ACTIONS_ENABLED=false`
- `LIVE_ACTION_POLICY=read_only_dry_run`

## Quick start (demo)

```bash
cp .env.example .env
npm install
npm run dev
```

- Frontend: `http://localhost:5173`
- Backend API: `http://localhost:8080/api/v1`

## Connect to a real Swarm

Set in `.env`:

```bash
APP_MODE=dev
DOCKER_HOST=unix:///var/run/docker.sock
```

For remote manager via TLS:

```bash
APP_MODE=dev
DOCKER_HOST=tcp://manager-host:2376
DOCKER_TLS_VERIFY=true
DOCKER_CERT_PATH=/path/to/certs
```

## Auth and roles

When `AUTH_ENABLED=true`, send `Authorization: Bearer <token>`.

Static token format:

```bash
AUTH_TOKENS=viewer:viewer:token1,operator:operator:token2,admin:admin:token3
```

Role model:

- `viewer`: read endpoints + diagnostics + assistant
- `operator`: viewer + write-gated action endpoints
- `admin`: operator privileges (plus policy-level governance)

## Write safety model

- Global gate: `WRITE_ACTIONS_ENABLED`
- Policy: `LIVE_ACTION_POLICY`
  - `read_only_dry_run`: read-only actions execute, mutations return validated plans
  - `allowlist_live`: allowlist intent only (current mutation adapters still dry-run)
  - `demo_only`: live mutations blocked outside demo
- Every action outcome includes status, mode, and `auditID`.

## Demo mode scenarios

Overview supports forced scenarios with query params:

- `/?scenario=healthy`
- `/?scenario=degraded`
- `/?scenario=incident-burst`
- `/?scenario=recovery`
- `/?scenario=disconnected`

Diagnostics and events also include deterministic synthetic datasets in demo for richer walkthroughs.

## Grafana and charting

SwarmLens renders native ECharts by default.

Set these frontend variables to switch selected sections to Grafana embeds:

```bash
VITE_GRAFANA_URL=https://grafana.example.com
VITE_GRAFANA_DASHBOARD_UID=swarmlens-main
VITE_GRAFANA_ORG_ID=1
VITE_GRAFANA_THEME=dark
VITE_GRAFANA_FROM=now-6h
VITE_GRAFANA_TO=now
VITE_GRAFANA_REFRESH=30s
VITE_GRAFANA_KIOSK=tv
VITE_GRAFANA_DATASOURCE=Prometheus
```

## Local Docker run

```bash
npm run docker:up
npm run docker:logs
npm run docker:down
```

## Production deployment

Use the production template and scripts:

```bash
cp deploy/env/prod.env.example .env
chmod 600 .env
./scripts/preflight-prod.sh
./scripts/deploy.sh
```

Rollback:

```bash
./scripts/rollback.sh
# or
ROLLBACK_TO=<sha-or-tag> ./scripts/rollback.sh
```

Use file-based secrets where possible:

- `AUTH_TOKENS_FILE`
- `PREDICTOR_SHARED_SECRET_FILE`
- `ASSISTANT_API_KEY_FILE`

See [docs/PRODUCTION.md](docs/PRODUCTION.md).

## CI/CD

- CI workflow runs backend format/vet/test, frontend lint/typecheck/test/build, predictor tests, and Docker smoke build.
- Release workflow builds and pushes multi-arch images to GHCR on `v*` tags.

See [docs/CI_CD.md](docs/CI_CD.md).

## Development commands

```bash
npm run lint
npm run lint:ci
npm run typecheck
npm run test:web
npm run test:go
npm run test:predictor
npm run build
```

## Repository layout

```text
backend/
  cmd/swarmlens/            Go entrypoint
  internal/
    audit/                  Audit store
    auth/                   Auth and write gate
    config/                 Environment config and validation
    docker/                 Docker API client + demo fixtures
    httpapi/                Router, middleware, handlers
    incident/               Incident store
    intelligence/           Deterministic diagnostics + plugins
    model/                  Canonical API/domain types
    predictor/              Predictor client with fallback
    state/                  Snapshot + telemetry cache
    stream/                 SSE bus
predictor/
  app/                      FastAPI predictor service
src/
  components/               Layout, charts, UI
  hooks/                    Event stream and helpers
  lib/                      API client, grafana, telemetry, mocks
  store/                    Zustand domain stores
  types/                    Frontend mirrors backend model types
docs/
  API.md
  ARCHITECTURE.md
  CONFIGURATION.md
  CI_CD.md
  DEMO_MODE.md
  PRODUCTION.md
  SECURITY.md
```

## Documentation index

- [docs/API.md](docs/API.md) - REST and SSE contracts
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - system design and data flow
- [docs/CONFIGURATION.md](docs/CONFIGURATION.md) - all environment variables
- [docs/CI_CD.md](docs/CI_CD.md) - CI and release pipelines
- [docs/DEMO_MODE.md](docs/DEMO_MODE.md) - demo datasets and scenarios
- [docs/PRODUCTION.md](docs/PRODUCTION.md) - VPS deployment and rollback
- [docs/SECURITY.md](docs/SECURITY.md) - controls and hardening checklist
- [CONTRIBUTING.md](CONTRIBUTING.md) - contribution workflow
- [CHANGELOG.md](CHANGELOG.md) - notable changes
