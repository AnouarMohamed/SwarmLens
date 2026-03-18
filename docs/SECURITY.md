# Security

## Trust model

SwarmLens sits between the browser and the Docker Engine API on a manager node.
It never exposes the Docker socket, TLS credentials, or secret/config values to the browser.

## Controls

| Control | Implementation |
|---|---|
| Authentication | Static token (`Authorization: Bearer`) or OIDC/JWT. Disabled by default in `dev`/`demo`. Required in `prod`. |
| Role-based access | `viewer` read-only · `operator` writes (if gate open) · `admin` policy + force actions |
| Global write gate | `WRITE_ACTIONS_ENABLED=false` by default. Must be explicit opt-in. `prod` rejects writes without auth. |
| Four-eyes approval | High-risk ops (drain node, remove stack) require a second authorized actor in `prod`. |
| Audit trail | Every mutating action records actor, role, action, resource, before/after spec, result, timestamp. Immutable append-only log. |
| Secret values | Never returned by any API endpoint. Only name, ID, created-at, and service references are exposed. |
| Rate limiting | Per-IP token bucket. Configurable via `RATE_LIMIT_*` env vars. |
| Security headers | `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin` |
| CORS | Origin-reflected with explicit method/header allowlist. |
| Non-root container | Runs as UID 1001. Read-only root filesystem recommended. |
| Docker socket | Bind-mounted read-only (`ro`) by default in all stack files. |
| Predictor auth | Shared secret via `X-Shared-Secret` header. Requests without valid secret rejected with 401. |

## Deployment hardening checklist (prod)

- [ ] `AUTH_ENABLED=true` with strong tokens or OIDC
- [ ] `WRITE_ACTIONS_ENABLED=false` unless actively needed
- [ ] Docker socket mounted `:ro`
- [ ] `swarmlens_auth_tokens` stored as a Swarm secret, not an env var
- [ ] TLS termination in front of port 8080 (nginx, Traefik, or Swarm ingress)
- [ ] `RATE_LIMIT_ENABLED=true`
- [ ] Network policy restricts predictor access to swarmlens service only

## Reporting a vulnerability

Open a GitHub Security Advisory or email the maintainer directly.
Do not open a public issue for security findings.
