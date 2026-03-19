# Contributing

Thanks for contributing to SwarmLens.

## Prerequisites

- Node.js 22+
- Go 1.25+
- Python 3.12+
- Docker + Docker Compose v2

## Local setup

```bash
git clone https://github.com/AnouarMohamed/SwarmLens.git
cd SwarmLens
cp .env.example .env
npm install
npm run dev
```

Default mode is `demo` and does not require a live Swarm cluster.

## Development workflow

1. Create a branch (`feature/*`, `fix/*`, `docs/*`, `chore/*`).
2. Implement changes with tests.
3. Update docs when behavior/contracts/config change.
4. Run local CI parity commands.
5. Open PR to `main`.

## Local validation

```bash
cd backend && go test ./...
cd ../predictor && python -m pytest -q
cd .. && npm run lint:ci
npm run typecheck
npm run test:web -- --passWithNoTests
npm run build
```

Optional full Docker validation:

```bash
docker build -t swarmlens:ci .
docker build -t swarmlens-predictor:ci predictor/
```

## Coding rules

### Backend

- Keep packages under `backend/internal` focused by responsibility.
- Add new env vars to:
  - `.env.example`
  - `backend/internal/config/config.go`
  - `docs/CONFIGURATION.md`
- For action-related changes, preserve `ActionOutcome` structure and audit linkage.

### Frontend

- Keep `src/types/index.ts` aligned with `backend/internal/model/types.go`.
- Use store-driven state (`src/store/*`) for shared domain data.
- Keep operational UI states explicit: `live`, `stale`, `disconnected`.

### Predictor

- Keep `/score` request/response contract backward compatible.
- Keep shared-secret behavior stable (`X-Shared-Secret`).

## Documentation requirements

When behavior changes, update relevant docs in the same PR:

- `README.md`
- `docs/API.md`
- `docs/CONFIGURATION.md`
- `docs/SECURITY.md`
- `docs/PRODUCTION.md`
- `docs/CI_CD.md`
- `docs/DEMO_MODE.md` (if demo flows change)

## Commit message style

Use conventional prefixes:

- `feat:`
- `fix:`
- `docs:`
- `test:`
- `chore:`
- `refactor:`
- `ops:`

## Pull request checklist

- [ ] Tests and checks pass locally
- [ ] API/schema changes reflected in docs
- [ ] Env variable changes documented
- [ ] Security-sensitive changes reviewed
- [ ] No unrelated formatting noise
