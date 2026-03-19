# SwarmLens API Reference

Base path: `/api/v1`

## Response envelopes

Most JSON responses use one of these envelopes:

```json
{ "data": { "...": "..." } }
```

```json
{ "data": [{ "...": "..." }], "meta": { "total": 1 } }
```

Error shape:

```json
{ "error": "human readable message", "code": "snake_case_code" }
```

## Authentication

When `AUTH_ENABLED=true`, pass:

```http
Authorization: Bearer <token>
```

Role model:

- `viewer`: read endpoints and diagnostics
- `operator`: write-gated operation endpoints
- `admin`: operator-level access plus policy/governance flows

If auth is disabled, API uses an internal synthetic principal for local/demo convenience.

## Observability and runtime

| Method | Path                  | Auth | Notes                             |
| ------ | --------------------- | ---- | --------------------------------- |
| GET    | `/healthz`            | none | Liveness check                    |
| GET    | `/readyz`             | none | Readiness and dependency checks   |
| GET    | `/metrics`            | none | JSON telemetry                    |
| GET    | `/metrics/prometheus` | none | Prometheus text exposition        |
| GET    | `/runtime`            | none | mode/auth/write/freshness posture |
| GET    | `/openapi.yaml`       | none | OpenAPI spec file                 |

## Operational intelligence

| Method | Path            | Auth   | Description                                        |
| ------ | --------------- | ------ | -------------------------------------------------- |
| GET    | `/ops/metrics`  | viewer | Timeseries metrics and service risk ranking        |
| GET    | `/ops/insights` | viewer | Hybrid insight payload with actions and hypotheses |

### `GET /ops/metrics` shape

```json
{
  "data": {
    "freshness": "live",
    "lastUpdated": "2026-03-19T12:10:00Z",
    "series": [
      {
        "timestamp": "2026-03-19T12:10:00Z",
        "healthyRatio": 0.98,
        "managersOnline": 3,
        "workersOnline": 5,
        "runningTasks": 124,
        "failedTasks": 4,
        "restartCount": 9,
        "critical": 2,
        "warning": 3,
        "riskScore": 0.61
      }
    ],
    "serviceRisk": [
      {
        "service": "payments-worker",
        "score": 0.82,
        "reasons": ["replica drift 0/2", "3 failed tasks"],
        "actionability": "immediate"
      }
    ]
  }
}
```

### `GET /ops/insights` shape

```json
{
  "data": {
    "summary": "Cluster is degraded with 2 critical findings.",
    "risk": {
      "score": 0.68,
      "confidence": 0.74,
      "factors": ["replica drift", "manager reachability"],
      "source": "predictor",
      "updatedAt": "2026-03-19T12:10:00Z"
    },
    "freshness": "stale",
    "hypotheses": [
      {
        "title": "Replica shortfall in payments-worker",
        "why": "desired replicas: 2; running tasks: 0",
        "confidence": 0.85
      }
    ],
    "actions": [
      {
        "title": "Run diagnostics",
        "description": "Recompute findings before remediation.",
        "endpointHint": "/api/v1/diagnostics/run",
        "priority": 1,
        "actionability": "immediate"
      }
    ],
    "generatedAt": "2026-03-19T12:10:00Z",
    "provider": "none",
    "sourceStrategy": "deterministic"
  }
}
```

## Inventory and control-plane resources

| Method | Path                      | Auth                  |
| ------ | ------------------------- | --------------------- |
| GET    | `/swarm`                  | viewer                |
| GET    | `/nodes`                  | viewer                |
| GET    | `/nodes/{id}`             | viewer                |
| POST   | `/nodes/{id}/drain`       | operator + write gate |
| POST   | `/nodes/{id}/activate`    | operator + write gate |
| GET    | `/stacks`                 | viewer                |
| GET    | `/stacks/{name}`          | viewer                |
| POST   | `/stacks/{name}/deploy`   | operator + write gate |
| DELETE | `/stacks/{name}`          | operator + write gate |
| GET    | `/services`               | viewer                |
| GET    | `/services/{id}`          | viewer                |
| POST   | `/services/{id}/scale`    | operator + write gate |
| POST   | `/services/{id}/restart`  | operator + write gate |
| POST   | `/services/{id}/update`   | operator + write gate |
| POST   | `/services/{id}/rollback` | operator + write gate |
| GET    | `/tasks`                  | viewer                |
| GET    | `/tasks/{id}`             | viewer                |
| POST   | `/tasks/{id}/restart`     | operator + write gate |
| GET    | `/networks`               | viewer                |
| GET    | `/volumes`                | viewer                |
| GET    | `/secrets`                | viewer                |
| GET    | `/configs`                | viewer                |

### Query parameters

- `GET /services?stack=<stackName>`
- `GET /tasks?service=<idOrName>&node=<idOrHost>&state=<state>`

## Events

| Method | Path             | Auth   | Notes                      |
| ------ | ---------------- | ------ | -------------------------- |
| GET    | `/events`        | viewer | `?type=<eventType>` filter |
| GET    | `/stream/events` | viewer | SSE stream                 |

SSE events from `/stream/events`:

- `connected`: initial handshake
- `swarm`: each event payload serialized as `SwarmEvent`

Note: this endpoint is protected by auth middleware when `AUTH_ENABLED=true`.

## Diagnostics

| Method | Path                | Auth   |
| ------ | ------------------- | ------ |
| GET    | `/diagnostics`      | viewer |
| POST   | `/diagnostics/run`  | viewer |
| GET    | `/diagnostics/{id}` | viewer |

Query parameters:

- `GET /diagnostics?severity=critical`

## Incidents

| Method | Path                      | Auth   |
| ------ | ------------------------- | ------ |
| GET    | `/incidents`              | viewer |
| POST   | `/incidents`              | viewer |
| GET    | `/incidents/{id}`         | viewer |
| PUT    | `/incidents/{id}`         | viewer |
| POST   | `/incidents/{id}/resolve` | viewer |

Request body for `POST /incidents`:

```json
{
  "title": "Replica drift in payments-worker",
  "description": "0/2 replicas",
  "severity": "high",
  "affectedServices": ["payments-worker"],
  "diagnosticRefs": ["finding-id"]
}
```

## Audit

| Method | Path     | Auth   | Notes                             |
| ------ | -------- | ------ | --------------------------------- |
| GET    | `/audit` | viewer | `limit` and `offset` query params |

## Action orchestrator

| Method | Path               | Auth                                                                     |
| ------ | ------------------ | ------------------------------------------------------------------------ |
| POST   | `/actions/execute` | viewer for read-only actions, operator + write gate for mutating actions |

Request:

```json
{
  "action": "diagnostics.run",
  "resource": "cluster",
  "resourceID": "main",
  "params": {}
}
```

Action response:

```json
{
  "data": {
    "action": "service.scale",
    "resource": "service",
    "resourceID": "payments-worker",
    "status": "dry_run",
    "mode": "dry_run",
    "executed": false,
    "message": "Validated and generated an execution plan.",
    "blockedReason": "action_not_implemented",
    "impact": "No live mutation executed.",
    "plan": ["step 1", "step 2"],
    "auditID": "audit_123",
    "timestamp": "2026-03-19T12:10:00Z"
  }
}
```

Current read-only actions that execute live:

- `diagnostics.run`
- `telemetry.refresh`
- `incident.create`

Other actions are policy-validated and currently return dry-run plans (or demo simulation in demo mode).

## Assistant (SSE)

| Method | Path              | Auth   | Transport           |
| ------ | ----------------- | ------ | ------------------- |
| POST   | `/assistant/chat` | viewer | `text/event-stream` |

Request body:

```json
{ "prompt": "What needs action right now?" }
```

SSE event sequence:

- `context`
- `insight`
- `hypothesis` (0..n)
- `action` (0..n)
- `message`
- `done`

## Predictor service (internal)

SwarmLens backend calls predictor at `PREDICTOR_BASE_URL/score`.

- Method: `POST /score`
- Header: `X-Shared-Secret` when configured
- Response: `{ score, confidence, factors, source }`

If predictor is unavailable, backend falls back to deterministic local scoring.
