# CI/CD

SwarmLens uses GitHub Actions for CI validation and release publishing.

## CI workflow

File: `.github/workflows/ci.yml`

Triggers:

- push to `main`
- pull request to `main`
- manual dispatch

Jobs:

1. `backend`

- Setup Go using `backend/go.mod`
- Download modules
- `gofmt` check (`test -z "$(gofmt -l .)"`)
- `go vet ./...`
- `go test -race ./...`

2. `frontend`

- Setup Node 22
- `npm ci`
- `npm run lint:ci`
- `npm run typecheck`
- `npm run test:web`
- `npm run build`

3. `predictor`

- Setup Python 3.12
- Install predictor requirements + pytest/httpx
- `python -m pytest -q`

4. `docker` (depends on backend/frontend/predictor)

- Build main image: `docker build -t swarmlens:ci .`
- Build predictor image: `docker build -t swarmlens-predictor:ci predictor/`
- Run startup smoke check against `/api/v1/healthz`

## Release workflow

File: `.github/workflows/release.yml`

Triggers:

- push tags matching `v*`
- manual dispatch

Behavior:

- Login to GHCR
- Build and push multi-arch images (`linux/amd64`, `linux/arm64`):
  - `ghcr.io/<owner>/swarmlens`
  - `ghcr.io/<owner>/swarmlens-predictor`
- Tag strategies:
  - semantic version (`vX.Y.Z`)
  - major.minor
  - commit SHA

## Local CI parity commands

Run before pushing:

```bash
cd backend && go test ./...
cd ../predictor && python -m pytest -q
cd .. && npm run lint:ci
npm run typecheck
npm run test:web -- --passWithNoTests
npm run build
docker build -t swarmlens:ci .
docker build -t swarmlens-predictor:ci predictor/
```

## Common failure modes

- Go version mismatch between `go.mod` and Docker/runner
- `gofmt` failures from unformatted backend changes
- duplicate Vitest CLI flags (`--passWithNoTests`) in script invocations
- predictor auth contract drift (`POST /score`, `X-Shared-Secret`)
- image build breakage due to Dockerfile stage changes

## Branch and merge expectations

- Keep `main` green.
- Treat CI failures as release blockers.
- Keep docs and env templates updated in the same PR as behavior changes.
