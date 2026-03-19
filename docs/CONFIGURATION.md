# Configuration Reference

This document lists all supported environment variables for SwarmLens.

Resolution rules:

- Direct variable value takes precedence.
- If supported, `<NAME>_FILE` is used when direct value is empty.
- Backend defaults are defined in `backend/internal/config/config.go`.

## Backend runtime

| Variable            | Default                       | Description                                   |
| ------------------- | ----------------------------- | --------------------------------------------- |
| `APP_MODE`          | `demo`                        | Runtime mode: `dev`, `demo`, `prod`.          |
| `PORT`              | `8080`                        | Backend HTTP port.                            |
| `DOCKER_HOST`       | `unix:///var/run/docker.sock` | Docker API endpoint for inventory refresh.    |
| `DOCKER_TLS_VERIFY` | `false`                       | Enable Docker TLS verification for TCP hosts. |
| `DOCKER_CERT_PATH`  | empty                         | Docker TLS cert directory.                    |

## Authentication

| Variable                   | Default | Description                              |
| -------------------------- | ------- | ---------------------------------------- |
| `AUTH_ENABLED`             | `false` | Enables bearer token authentication.     |
| `AUTH_TOKENS`              | empty   | Static token list `user:role:token,...`. |
| `AUTH_TOKENS_FILE`         | empty   | File containing `AUTH_TOKENS` value.     |
| `AUTH_PROVIDER`            | empty   | Reserved for external providers.         |
| `AUTH_OIDC_ISSUER_URL`     | empty   | Reserved OIDC config.                    |
| `AUTH_OIDC_CLIENT_ID`      | empty   | Reserved OIDC config.                    |
| `AUTH_OIDC_USERNAME_CLAIM` | empty   | Reserved OIDC config.                    |
| `AUTH_OIDC_ROLE_CLAIM`     | empty   | Reserved OIDC config.                    |
| `AUTH_TRUSTED_PROXY_CIDRS` | empty   | Trusted proxy CIDR list.                 |

## Write safety and action policy

| Variable                  | Default             | Description                                         |
| ------------------------- | ------------------- | --------------------------------------------------- |
| `WRITE_ACTIONS_ENABLED`   | `false`             | Global write gate.                                  |
| `WRITE_APPROVAL_REQUIRED` | `true`              | Policy hint for high-risk approval workflows.       |
| `LIVE_ACTION_POLICY`      | `read_only_dry_run` | `read_only_dry_run`, `allowlist_live`, `demo_only`. |

## Diagnostics and freshness

| Variable                 | Default | Description                                          |
| ------------------------ | ------- | ---------------------------------------------------- |
| `DIAGNOSTICS_SCHEDULE`   | `60`    | Seconds between auto diagnostic refresh triggers.    |
| `SNAPSHOT_STALE_SECONDS` | `45`    | Freshness threshold before data is considered stale. |

## Predictor integration

| Variable                       | Default | Description                                                    |
| ------------------------------ | ------- | -------------------------------------------------------------- |
| `PREDICTOR_BASE_URL`           | empty   | Predictor service base URL (example: `http://predictor:8001`). |
| `PREDICTOR_SHARED_SECRET`      | empty   | Shared secret sent as `X-Shared-Secret`.                       |
| `PREDICTOR_SHARED_SECRET_FILE` | empty   | File containing shared secret value.                           |

## Assistant integration

| Variable                 | Default | Description                                                              |
| ------------------------ | ------- | ------------------------------------------------------------------------ |
| `ASSISTANT_PROVIDER`     | `none`  | `none` or `openai`.                                                      |
| `ASSISTANT_API_BASE_URL` | empty   | OpenAI-compatible API base (example `https://api.openai.com`).           |
| `ASSISTANT_API_KEY`      | empty   | API key for assistant provider.                                          |
| `ASSISTANT_API_KEY_FILE` | empty   | File containing assistant API key.                                       |
| `ASSISTANT_MODEL`        | empty   | Model name (defaults to `gpt-4o-mini` when OpenAI is enabled and empty). |
| `ASSISTANT_RAG_ENABLED`  | `true`  | Feature flag for retrieval-augmented assistant flows.                    |

## Rate limiting

| Variable                    | Default | Description                                     |
| --------------------------- | ------- | ----------------------------------------------- |
| `RATE_LIMIT_ENABLED`        | `true`  | Enables per-client request limiting middleware. |
| `RATE_LIMIT_REQUESTS`       | `300`   | Allowed requests per window.                    |
| `RATE_LIMIT_WINDOW_SECONDS` | `60`    | Window duration in seconds.                     |

## Observability and alerts

| Variable                      | Default             | Description                                            |
| ----------------------------- | ------------------- | ------------------------------------------------------ |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | empty               | OTLP endpoint for traces/metrics export.               |
| `OTEL_SERVICE_NAME`           | `swarmlens-backend` | Service name label for telemetry.                      |
| `OTEL_TRACES_SAMPLE_RATIO`    | empty               | Trace sampling hint (reserved for future OTEL wiring). |
| `SLACK_WEBHOOK_URL`           | empty               | Alerting integration hook (reserved).                  |
| `ALERTMANAGER_WEBHOOK_URL`    | empty               | Alertmanager integration hook (reserved).              |
| `PAGERDUTY_ROUTING_KEY`       | empty               | PagerDuty integration key (reserved).                  |

## Frontend-only variables (Vite)

These are evaluated at frontend build time.

| Variable                     | Default      | Description                                    |
| ---------------------------- | ------------ | ---------------------------------------------- |
| `VITE_API_BASE`              | `/api/v1`    | API base path used by frontend client.         |
| `VITE_GRAFANA_URL`           | empty        | Grafana base URL.                              |
| `VITE_GRAFANA_DASHBOARD_UID` | empty        | Dashboard UID for embed links.                 |
| `VITE_GRAFANA_ORG_ID`        | `1`          | Grafana org ID.                                |
| `VITE_GRAFANA_THEME`         | `dark`       | Embed theme.                                   |
| `VITE_GRAFANA_FROM`          | `now-6h`     | Default time range start.                      |
| `VITE_GRAFANA_TO`            | `now`        | Default time range end.                        |
| `VITE_GRAFANA_REFRESH`       | `30s`        | Refresh interval for embed links.              |
| `VITE_GRAFANA_KIOSK`         | `tv`         | Grafana kiosk mode for embeds.                 |
| `VITE_GRAFANA_DATASOURCE`    | `Prometheus` | Default datasource in generated Explore links. |

## Predictor service variables

Set in predictor container/process:

| Variable                       | Default | Description                            |
| ------------------------------ | ------- | -------------------------------------- |
| `PREDICTOR_SHARED_SECRET`      | empty   | Shared secret for `/score` auth check. |
| `PREDICTOR_SHARED_SECRET_FILE` | empty   | File containing shared secret.         |

## Production template

Use `deploy/env/prod.env.example` as a baseline:

```bash
cp deploy/env/prod.env.example .env
chmod 600 .env
```

Then validate with:

```bash
./scripts/preflight-prod.sh
```
