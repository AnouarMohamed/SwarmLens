# Changelog

All notable changes to this project are documented in this file.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- Industrial dark redesign for shell, sidebar, top header, and Overview experience.
- Rich Overview modules: KPI narrative, findings, events, posture, guidance, and quick actions.
- Native ECharts telemetry modules plus optional Grafana embed mode.
- Demo scenario presets: `healthy`, `degraded`, `incident-burst`, `recovery`, `disconnected`.
- Operational intelligence endpoints:
  - `GET /api/v1/ops/metrics`
  - `GET /api/v1/ops/insights`
- Action orchestrator endpoint: `POST /api/v1/actions/execute` with structured `ActionOutcome`.
- Assistant streaming endpoint: `POST /api/v1/assistant/chat` with SSE event stream.
- Predictor contract alignment to `POST /score`.
- Production deployment scripts:
  - `scripts/preflight-prod.sh`
  - `scripts/deploy.sh`
  - `scripts/rollback.sh`

### Changed

- Backend and predictor now support file-based secrets (`*_FILE`) for production-safe secret injection.
- CI pipeline gates backend, frontend, predictor, and Docker smoke builds.
- Main Docker build uses Go 1.25 to match backend toolchain requirements.
- Frontend test script updated to avoid duplicate `--passWithNoTests` conflicts.

### Fixed

- Predictor Docker user creation is now non-interactive and deterministic.
- Production overlay now wires predictor base URL and secret-file envs correctly.
- Documentation set updated and expanded for API, architecture, config, demo mode, CI/CD, security, and production ops.
