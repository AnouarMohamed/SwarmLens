# Demo Mode

SwarmLens demo mode is designed to feel like a real operator workflow, not a blank mock.

## Goals

- Keep product behavior interactive without requiring a live Swarm cluster.
- Preserve realistic operational narratives (healthy, degraded, disconnected, incident pressure, recovery).
- Exercise the same frontend action and incident paths used in live mode.

## Backend demo behavior

When `APP_MODE=demo`:

- Docker client runs in demo mode and returns fixture snapshot/event data.
- Router seeds cache with demo snapshot and events.
- Freshness/risk/diagnostics APIs still run through normal handler logic.

Relevant files:

- `backend/internal/docker/demo.go`
- `backend/internal/httpapi/router.go`

## Frontend demo behavior

Frontend augments demo UX with deterministic scenario presets and synthetic telemetry.

Overview supports scenario query params:

- `/?scenario=healthy`
- `/?scenario=degraded`
- `/?scenario=incident-burst`
- `/?scenario=recovery`
- `/?scenario=disconnected`

Diagnostics can fallback to synthetic findings when live findings are empty in demo mode.

Relevant files:

- `src/views/overview/OverviewView.tsx`
- `src/views/diagnostics/DiagnosticsView.tsx`
- `src/lib/mockData.ts`
- `src/lib/telemetry.ts`

## Synthetic datasets

`src/lib/mockData.ts` seeds:

- rich findings (critical/high/medium/low/info)
- realistic event stream samples for services/nodes/network/audit

`src/lib/telemetry.ts` generates deterministic trend series for chart modules.

## Demo action behavior

Actions still call the real action orchestrator endpoint.

- read-only actions execute normally (`diagnostics.run`, `telemetry.refresh`, `incident.create`)
- mutating actions are simulated or dry-run per policy
- every action returns structured `ActionOutcome` and audit reference

This keeps operator interaction semantics close to production while remaining safe.

## How to run demo

```bash
cp .env.example .env
# keep APP_MODE=demo
npm install
npm run dev
```

## Suggested demo walkthrough

1. Open Overview in `healthy` scenario.
2. Switch to `degraded` and inspect findings/events/telemetry deltas.
3. Switch to `incident-burst` and open Incidents workflow.
4. Switch to `disconnected` and verify stale/disconnected affordances.
5. Open Diagnostics and generate incidents from findings.
6. Open Assistant and request triage summary and next actions.
