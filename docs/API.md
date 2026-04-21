# SwarmLens API Reference

Base path: `/api/v1`

Canonical contract: [docs/openapi.yaml](openapi.yaml)

This document is the human-oriented overview. The OpenAPI file above is the source of truth for the v1.5 control-plane contract and is used to generate frontend types.

## Contract shape

Most JSON responses use one of these envelopes:

```json
{ "data": { "...": "..." } }
```

```json
{ "data": [{ "...": "..." }], "meta": { "total": 1 } }
```

Errors use:

```json
{ "error": "human readable message", "code": "snake_case_code" }
```

## Authentication

Primary production auth is OIDC session login with an HttpOnly cookie.

Supported modes:

- Session cookie auth for browser flows.
- Static bearer token auth for dev and break-glass use.
- Synthetic local admin identity when auth is disabled.

Role model:

- `viewer`: read-only access.
- `operator`: diagnostics, incidents, restart, bounded scale, and other write-gated ops.
- `admin`: approvals, cluster management, and rollback-capable flows.

## Primary routing model

v1.5 is cluster-scoped. The primary route family is:

```text
/api/v1/clusters/{clusterID}/...
```

Legacy single-cluster aliases still exist for compatibility, but they are intentionally omitted from the canonical contract.

## Core endpoints

### Observability

| Method | Path            | Auth | Notes                    |
| ------ | --------------- | ---- | ------------------------ |
| GET    | `/healthz`      | none | Liveness probe           |
| GET    | `/readyz`       | none | Dependency readiness     |
| GET    | `/runtime`      | none | Runtime posture          |
| GET    | `/openapi.yaml` | none | Canonical OpenAPI spec   |

### Auth

| Method | Path           | Auth | Notes                    |
| ------ | -------------- | ---- | ------------------------ |
| GET    | `/auth/me`     | none | Session or bearer lookup |
| POST   | `/auth/logout` | none | Ends current session     |

### Clusters

| Method | Path                    | Auth  |
| ------ | ----------------------- | ----- |
| GET    | `/clusters`             | admin |
| POST   | `/clusters`             | admin |
| GET    | `/clusters/{clusterID}` | admin |
| PUT    | `/clusters/{clusterID}` | admin |

### Inventory and workload reads

| Method | Path                                        | Auth   |
| ------ | ------------------------------------------- | ------ |
| GET    | `/clusters/{clusterID}/swarm`               | viewer |
| GET    | `/clusters/{clusterID}/nodes`               | viewer |
| GET    | `/clusters/{clusterID}/nodes/{id}`          | viewer |
| GET    | `/clusters/{clusterID}/stacks`              | viewer |
| GET    | `/clusters/{clusterID}/stacks/{name}`       | viewer |
| GET    | `/clusters/{clusterID}/services`            | viewer |
| GET    | `/clusters/{clusterID}/services/{id}`       | viewer |
| GET    | `/clusters/{clusterID}/tasks`               | viewer |
| GET    | `/clusters/{clusterID}/tasks/{id}`          | viewer |
| GET    | `/clusters/{clusterID}/networks`            | viewer |
| GET    | `/clusters/{clusterID}/volumes`             | viewer |
| GET    | `/clusters/{clusterID}/secrets`             | viewer |
| GET    | `/clusters/{clusterID}/configs`             | viewer |
| GET    | `/clusters/{clusterID}/events`              | viewer |
| GET    | `/clusters/{clusterID}/stream/events`       | viewer |

Useful query parameters:

- `GET /clusters/{clusterID}/services?stack=<stackName>`
- `GET /clusters/{clusterID}/tasks?service=<idOrName>&node=<idOrHost>&state=<state>`
- `GET /clusters/{clusterID}/events?type=<eventType>`

### Cluster posture and intelligence

| Method | Path                                   | Auth   |
| ------ | -------------------------------------- | ------ |
| GET    | `/clusters/{clusterID}/diagnostics`    | viewer |
| GET    | `/clusters/{clusterID}/ops/metrics`    | viewer |
| GET    | `/clusters/{clusterID}/ops/insights`   | viewer |
| GET    | `/clusters/{clusterID}/audit`          | viewer |

### Incidents

| Method | Path                                           | Auth     |
| ------ | ---------------------------------------------- | -------- |
| GET    | `/clusters/{clusterID}/incidents`              | viewer   |
| POST   | `/clusters/{clusterID}/incidents`              | operator |
| GET    | `/clusters/{clusterID}/incidents/{id}`         | viewer   |
| PUT    | `/clusters/{clusterID}/incidents/{id}`         | operator |
| POST   | `/clusters/{clusterID}/incidents/{id}/resolve` | operator |

### Actions and approvals

| Method | Path                                               | Auth     |
| ------ | -------------------------------------------------- | -------- |
| POST   | `/clusters/{clusterID}/nodes/{id}/drain`           | operator |
| POST   | `/clusters/{clusterID}/nodes/{id}/activate`        | operator |
| POST   | `/clusters/{clusterID}/stacks/{name}/deploy`       | operator |
| DELETE | `/clusters/{clusterID}/stacks/{name}`              | operator |
| POST   | `/clusters/{clusterID}/services/{id}/scale`        | operator |
| POST   | `/clusters/{clusterID}/services/{id}/restart`      | operator |
| POST   | `/clusters/{clusterID}/services/{id}/update`       | operator |
| POST   | `/clusters/{clusterID}/services/{id}/rollback`     | operator |
| POST   | `/clusters/{clusterID}/tasks/{id}/restart`         | operator |
| GET    | `/clusters/{clusterID}/actions`                    | viewer   |
| POST   | `/clusters/{clusterID}/actions/execute`            | operator |
| GET    | `/clusters/{clusterID}/approvals`                  | admin    |
| POST   | `/clusters/{clusterID}/approvals/{id}/approve`     | admin    |
| POST   | `/clusters/{clusterID}/approvals/{id}/reject`      | admin    |

Notes:

- Every live action requires a human reason.
- `stack.deploy` and `stack.remove` are contract-covered but still dry-run only in this milestone.
- `service.rollback` and large scale changes enter approval first.
- Assistant-suggested actions use the same action pipeline.

### Assistant

| Method | Path                                              | Auth     | Notes                 |
| ------ | ------------------------------------------------- | -------- | --------------------- |
| GET    | `/clusters/{clusterID}/assistant/sessions`        | viewer   | Durable session list  |
| POST   | `/clusters/{clusterID}/assistant/sessions`        | operator | Create session        |
| GET    | `/clusters/{clusterID}/assistant/sessions/{id}`   | viewer   | Session with messages |
| POST   | `/clusters/{clusterID}/assistant/chat`            | operator | `text/event-stream`   |

SSE events currently include:

- `session`
- `context`
- `insight`
- `hypothesis`
- `action`
- `citation`
- `action_proposal`
- `message`
- `done`

## Contract generation workflow

Frontend control-plane types are generated from `docs/openapi.yaml`.

The generated contract now covers:

- auth and session identity
- cluster management
- cluster posture
- inventory and workload read models
- typed stack mutation request bodies
- typed node, service, and task action request bodies
- diagnostics, incidents, audit
- actions, approvals, assistant sessions

Regenerate them with:

```bash
npm run contracts:generate
```
