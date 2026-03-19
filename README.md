# SwarmLens

Swarm operations intelligence with safe execution.

**Stack:** React + Vite · Go API · FastAPI predictor · Docker Stack

---

## What it is

SwarmLens helps engineers diagnose Docker Swarm failures, execute operations safely, and manage incidents end-to-end. The core is a **deterministic diagnostic engine** that produces structured findings — severity, evidence, and a specific recommendation — for every class of Swarm failure.

This is not a Portainer clone. Portainer covers generic container management. SwarmLens covers _why your Swarm is broken_ and _how to fix it safely_.

---

## Quick start

```bash
cp .env.example .env
npm install
npm run dev
```

- Frontend → `http://localhost:5173`
- Backend → `http://localhost:8080`

Runs in `demo` mode with fixture data. No Swarm required.

---

## Connect to a real Swarm

**Unix socket (local manager):**

```bash
# In .env
APP_MODE=dev
DOCKER_HOST=unix:///var/run/docker.sock
```

**TCP + TLS (remote manager):**

```bash
APP_MODE=dev
DOCKER_HOST=tcp://manager-host:2376
DOCKER_TLS_VERIFY=true
DOCKER_CERT_PATH=/path/to/certs
```

---

## Deploy to Swarm

```bash
# Demo — no auth, no writes
docker stack deploy -c deploy/overlays/demo/stack.yml swarmlens

# Production — requires creating secrets first
docker secret create swarmlens_auth_tokens - <<< "admin:admin:your-strong-token"
docker secret create swarmlens_predictor_secret - <<< "your-predictor-secret"
docker stack deploy -c deploy/overlays/prod/stack.yml swarmlens
```

---

## Production automation (VPS)

Use the production env template and rollout scripts:

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

Use `_FILE` secrets in production where possible:
`AUTH_TOKENS_FILE`, `PREDICTOR_SHARED_SECRET_FILE`, `ASSISTANT_API_KEY_FILE`.

Full guide: [docs/PRODUCTION.md](docs/PRODUCTION.md)

---

## Diagnostic plugins

| Plugin                  | Detects                                                 |
| ----------------------- | ------------------------------------------------------- |
| `replica-mismatch`      | Desired vs running vs pending tasks                     |
| `placement-failure`     | Unsatisfied constraints / resource reservation blocking |
| `crash-loop`            | Task restart churn, exit codes, healthcheck instability |
| `image-pull-failure`    | Registry unreachable, bad digest, credential issues     |
| `port-conflict`         | Routing mesh / published port collisions                |
| `secret-config-ref`     | Missing secret/config reference, bad target path        |
| `quorum-risk`           | Manager count vs failure tolerance, Raft health         |
| `update-rollback-state` | Update paused, failure ratio tripped, rollback pending  |
| `node-pressure`         | Reserved vs available CPU/memory causing pending tasks  |

Output format:

```json
{
  "severity": "critical",
  "resource": "payments/worker",
  "scope": "service",
  "message": "Service worker has 0/2 running tasks.",
  "evidence": ["desired replicas: 2", "running tasks: 0", "shortfall: 2"],
  "recommendation": "Check task failure reasons in the Tasks view.",
  "source": "replica-mismatch"
}
```

---

## Modes

| Mode   | Auth         | Writes | Use case                    |
| ------ | ------------ | ------ | --------------------------- |
| `dev`  | Off          | Off    | Local engineering           |
| `demo` | Off          | Off    | Safe showcase with fixtures |
| `prod` | **Required** | Opt-in | Controlled production ops   |

`WRITE_ACTIONS_ENABLED=false` by default in all modes.
`prod` refuses to boot without `AUTH_ENABLED=true`.

---

## Auth

Static token:

```bash
AUTH_ENABLED=true
AUTH_TOKENS=viewer:viewer:token1,operator:operator:token2,admin:admin:token3
```

OIDC:

```bash
AUTH_ENABLED=true
AUTH_PROVIDER=oidc
AUTH_OIDC_ISSUER_URL=https://your-idp/.well-known/openid-configuration
AUTH_OIDC_CLIENT_ID=swarmlens
```

| Role       | Permissions                                |
| ---------- | ------------------------------------------ |
| `viewer`   | Read-only — all views, events, diagnostics |
| `operator` | viewer + write actions (if gate enabled)   |
| `admin`    | operator + policy + force actions          |

---

## Observability

| Endpoint                         | Description               |
| -------------------------------- | ------------------------- |
| `GET /api/v1/healthz`            | Liveness                  |
| `GET /api/v1/readyz`             | Readiness                 |
| `GET /api/v1/metrics`            | JSON telemetry            |
| `GET /api/v1/metrics/prometheus` | Prometheus format         |
| `GET /api/v1/runtime`            | Mode and security posture |

### Grafana dashboard embed

If you have Grafana available, the frontend can render live panel embeds on
Overview and Diagnostics.

```bash
VITE_GRAFANA_URL=https://grafana.example.com
VITE_GRAFANA_DASHBOARD_UID=swarmlens-main
VITE_GRAFANA_ORG_ID=1
VITE_GRAFANA_THEME=dark
VITE_GRAFANA_REFRESH=30s
```

---

## Docker

```bash
npm run docker:up    # builds and starts swarmlens + predictor
npm run docker:down
npm run docker:logs
```

---

## Development

```bash
npm run lint          # ESLint + Prettier
npm run test:go       # Go tests (race detector)
npm run test:web      # Vitest frontend tests
npm run test:predictor # Pytest scorer tests
npm run test:e2e      # Playwright (requires dev server)
npm run ci:backend    # go fmt + go vet + go test -race
```

---

## Directory structure

```
backend/
  cmd/swarmlens/         Entry point
  internal/
    config/              Env loading + validation
    model/               Canonical Swarm types
    docker/              Docker Engine API client + demo fixtures
    state/               Snapshot cache
    intelligence/        Diagnostic engine + 9 plugins
    auth/                Token auth + write gate
    httpapi/             Router, middleware, all handlers
    inventory/           (Phase 1) Live Docker API adapters
    actions/             (Phase 2) Write action implementations
    audit/               Audit log
    incident/            Incident lifecycle
    predictor/           Predictor client with fallback
    stream/              SSE event bus
predictor/
  app/                   FastAPI service (main, models, scorer)
src/
  views/                 One folder per route (11 views)
  components/            layout/ + ui/
  store/                 Zustand stores
  hooks/                 useEventStream, useWriteGate
  lib/                   api.ts, utils.ts
  types/                 index.ts (mirrors internal/model/types.go)
deploy/
  overlays/              dev / demo / prod Swarm stack files
docs/                    PRODUCT_SPEC, ARCHITECTURE, API, SECURITY, IMPLEMENTATION_PROGRAM
```

---

## Comparison

| Capability                       | SwarmLens | Portainer | Swarmpit |
| -------------------------------- | --------- | --------- | -------- |
| Deterministic diagnostics        | ✅        | ❌        | ❌       |
| Evidence-based failure analysis  | ✅        | ❌        | ❌       |
| Incident + runbook workflow      | ✅        | ❌        | ❌       |
| Audit trail with spec snapshots  | ✅        | Partial   | ❌       |
| Four-eyes approval for risky ops | ✅        | ❌        | ❌       |
| Write gate (safe by default)     | ✅        | ❌        | ❌       |
| Real-time SSE event stream       | ✅        | Partial   | Partial  |
| Optional AI assistant            | ✅        | ❌        | ❌       |

---

## Documentation

- [docs/PRODUCT_SPEC.md](docs/PRODUCT_SPEC.md) — entities, views, actions, safety model, API contract
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system topology and data flow
- [docs/API.md](docs/API.md) — endpoint reference
- [docs/SECURITY.md](docs/SECURITY.md) — controls and hardening checklist
- [docs/IMPLEMENTATION_PROGRAM.md](docs/IMPLEMENTATION_PROGRAM.md) — phased roadmap with quality gates
- [CONTRIBUTING.md](CONTRIBUTING.md) — setup, conventions, and plugin development guide
- [CHANGELOG.md](CHANGELOG.md) — version history
