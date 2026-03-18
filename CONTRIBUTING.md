# Contributing

## Setup

```bash
git clone https://github.com/AnouarMohamed/swarmlens
cd swarmlens
cp .env.example .env
npm install
npm run dev          # starts both backend (port 8080) and frontend (port 5173)
```

Runs in `demo` mode — no Swarm required.

## Project layout

```
backend/cmd/swarmlens/   Entry point
backend/internal/        All Go packages (one responsibility per package)
src/views/               One folder per route
src/components/          Shared layout + UI components
src/store/               Zustand stores (one per domain)
predictor/app/           FastAPI predictor service
```

## Branch and commit conventions

- Branch: `feature/<slug>`, `fix/<slug>`, `docs/<slug>`
- Commits: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:`
- Every PR must pass CI (lint + tests + Docker build)

## Backend rules

- All packages live under `backend/internal/`
- No flat Go files in `backend/` root except `go.mod`/`go.sum`
- One package per responsibility — no mega-packages
- Every new env variable must be added to `.env.example` and `internal/config/config.go`
- Every mutating handler must write an audit record via `d.auditLog.Record(...)`
- Demo mode must always work with no external dependencies

## Frontend rules

- `src/types/index.ts` must mirror `backend/internal/model/types.go` — keep in sync manually
- One folder per route under `src/views/`
- Shared components go in `src/components/ui/` or `src/components/layout/`
- State lives in `src/store/` — one store per domain, no prop drilling
- No `any` types — TypeScript strict mode is enforced

## Tests

```bash
npm run test:go         # Go unit tests with race detector
npm run test:web        # Vitest frontend unit tests
npm run test:predictor  # Pytest for predictor scorer
npm run test:e2e        # Playwright e2e (requires dev server running)
```

## Adding a diagnostic plugin

1. Create `backend/internal/intelligence/plugins/<name>.go`
2. Implement the `intelligence.Plugin` interface (`Name() string`, `Analyze(model.Snapshot) []model.Finding`)
3. Register in `backend/internal/intelligence/plugins/register.go`
4. Add fixture data to `backend/internal/docker/demo.go` that triggers the plugin
5. Add the plugin name to `docs/PRODUCT_SPEC.md`
