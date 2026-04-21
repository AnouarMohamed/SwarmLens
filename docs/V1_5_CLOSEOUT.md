# SwarmLens v1.5 Closeout

This checklist tracks what is already in place for the v1.5 control-plane milestone and what still needs a final validation pass before we can call the release fully done.

## Implemented foundations

- [x] Postgres-backed durable state for clusters, incidents, audit entries, approvals, action runs, and assistant sessions in production.
- [x] Swarm-first deployment baseline with `backend`, `predictor`, `postgres`, and a migration service/job path.
- [x] Cluster model and cluster-scoped APIs under `/api/v1/clusters/{clusterID}/...` with default-cluster compatibility aliases.
- [x] OIDC login, backend-managed session cookies, `/auth/*` endpoints, and CSRF protection for session-authenticated writes.
- [x] RBAC with `viewer`, `operator`, and `admin` roles plus policy-gated write actions.
- [x] Durable approval and action pipeline with audit linkage and assistant proposals routed through the same flow.
- [x] Persistent assistant sessions, messages, citations, and action proposals.
- [x] Focused multi-cluster frontend shell with approvals, assistant history, cluster switching, and typed direct action routes.
- [x] OpenAPI-backed generated frontend contracts for control-plane and inventory/action surfaces.
- [x] Lazy native chart loading so Grafana-enabled sessions can skip the heavy native chart bundle.

## Confidence work completed

- [x] Backend tests for OIDC claim mapping, role precedence, and username fallback behavior.
- [x] Backend tests for cluster middleware default-cluster selection, requested cluster selection, and missing/disabled cluster handling.
- [x] Backend tests for write-safety guardrails including reason-required and scale approval thresholds.

## Remaining before calling v1.5 fully complete

- [ ] Run a real Postgres-backed, multi-cluster smoke test that survives backend restarts and verifies durable state end to end.
- [ ] Expand Playwright coverage to include login/logout, cluster switching, approvals, assistant cited responses, and incident/operator workflows.
- [ ] Rehearse a production-like OIDC deployment with migration, health-gated rollout, and rollback verification.
- [ ] Decide whether the remaining lazy chart chunk is acceptable or worth one more slimming pass before release.
