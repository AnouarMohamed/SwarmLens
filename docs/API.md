# API Reference

Base path: `/api/v1`

All responses use the envelope:
```json
{ "data": <payload>, "meta": { "total": <int> } }
```
Errors:
```json
{ "error": "<message>", "code": "<snake_case_code>" }
```

## Auth

Pass `Authorization: Bearer <token>` on every request when `AUTH_ENABLED=true`.

## Observability

| Method | Path | Description |
|---|---|---|
| GET | `/healthz` | Liveness — always 200 if process is running |
| GET | `/readyz` | Readiness — 200 ok / 503 degraded |
| GET | `/metrics` | JSON API telemetry |
| GET | `/metrics/prometheus` | Prometheus exposition format |
| GET | `/runtime` | Mode, auth, write gate posture |
| GET | `/openapi.yaml` | OpenAPI contract |

## Swarm

| Method | Path | Auth |
|---|---|---|
| GET | `/swarm` | viewer |

## Nodes

| Method | Path | Auth |
|---|---|---|
| GET | `/nodes` | viewer |
| GET | `/nodes/:id` | viewer |
| POST | `/nodes/:id/drain` | operator + write gate |
| POST | `/nodes/:id/activate` | operator + write gate |

## Stacks

| Method | Path | Auth |
|---|---|---|
| GET | `/stacks` | viewer |
| GET | `/stacks/:name` | viewer |
| POST | `/stacks/:name/deploy` | operator + write gate |
| DELETE | `/stacks/:name` | admin + write gate + approval (prod) |

## Services

| Method | Path | Auth | Body |
|---|---|---|---|
| GET | `/services` | viewer | — |
| GET | `/services/:id` | viewer | — |
| POST | `/services/:id/scale` | operator | `{"replicas": N}` |
| POST | `/services/:id/restart` | operator | — |
| POST | `/services/:id/update` | operator | `{"image": "...", "env": {...}}` |
| POST | `/services/:id/rollback` | operator | — |

## Tasks

| Method | Path | Query params |
|---|---|---|
| GET | `/tasks` | `service`, `node`, `state` |
| GET | `/tasks/:id` | — |

## Networks / Volumes / Secrets / Configs

| Method | Path |
|---|---|
| GET | `/networks` |
| GET | `/volumes` |
| GET | `/secrets` |
| GET | `/configs` |

## Events

| Method | Path | Description |
|---|---|---|
| GET | `/events` | Recent events (query: `?type=service`) |
| GET | `/stream/events` | SSE live stream |

## Diagnostics

| Method | Path | Auth |
|---|---|---|
| GET | `/diagnostics` | viewer (query: `?severity=critical`) |
| POST | `/diagnostics/run` | viewer |
| GET | `/diagnostics/:id` | viewer |

## Incidents

| Method | Path | Auth | Body |
|---|---|---|---|
| GET | `/incidents` | viewer | — |
| POST | `/incidents` | operator | `{"title","description","severity","affectedServices","diagnosticRefs"}` |
| GET | `/incidents/:id` | viewer | — |
| PUT | `/incidents/:id` | operator | `{"status","note"}` |
| POST | `/incidents/:id/resolve` | operator | — |

## Audit

| Method | Path | Query params |
|---|---|---|
| GET | `/audit` | `limit` (default 50), `offset` (default 0) |

## Assistant

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/assistant/chat` | viewer | SSE streaming response. Requires `ASSISTANT_PROVIDER` to be set. |
