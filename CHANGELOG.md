# Changelog

All notable changes to SwarmLens are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## [Unreleased]

### Added
- Initial project scaffold — Phase 0 complete
- `backend/internal/config` — env loading with prod safety validation
- `backend/internal/model` — canonical Swarm types (Node, Service, Task, Finding, Incident, AuditEntry, Snapshot)
- `backend/internal/docker` — Docker Engine API client factory (socket + TCP/TLS) and demo fixture snapshot
- `backend/internal/state` — thread-safe snapshot cache
- `backend/internal/intelligence` — deterministic diagnostic engine with 9 plugins:
  replica-mismatch, placement-failure, crash-loop, image-pull-failure, port-conflict,
  secret-config-ref, quorum-risk, update-rollback-state, node-pressure
- `backend/internal/auth` — static token extraction and role checking, write gate
- `backend/internal/stream` — in-process SSE event bus
- `backend/internal/audit` — append-only in-memory audit log
- `backend/internal/incident` — incident lifecycle (create, update, resolve)
- `backend/internal/predictor` — optional predictor client with local deterministic fallback
- `backend/internal/httpapi` — full HTTP router, middleware stack, all handlers
- `src/` — React + Vite frontend with 11 views, Zustand stores, typed API client
- `predictor/` — FastAPI risk scoring service with deterministic scorer
- `deploy/overlays/` — dev, demo, prod Swarm stack files
- `Dockerfile` — multi-stage Go + Vite + Alpine final image
- `docker-compose.yml` — local dev compose
- `.github/workflows/` — CI (lint + test + build) and release (GHCR push on tag)
- `docs/` — PRODUCT_SPEC, ARCHITECTURE, FEATURES (placeholder), API, SECURITY, IMPLEMENTATION_PROGRAM
