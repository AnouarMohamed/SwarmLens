"""
Deterministic risk scorer.
Produces a 0–1 score and confidence from Swarm signals.
Higher score = higher incident risk.
"""
from .models import ScoreRequest


def score_snapshot(req: ScoreRequest) -> dict:
    score = 0.0
    factors = []

    # Manager quorum risk
    if req.managers == 1:
        score += 0.35
        factors.append("single manager — no fault tolerance")
    elif req.managers == 2:
        score += 0.15
        factors.append("2 managers — quorum fragile")

    # Replica mismatch
    for svc in req.services:
        if svc.desired_replicas > 0 and svc.running_tasks == 0:
            score += 0.20
            factors.append(f"{svc.name}: 0/{svc.desired_replicas} replicas running")
            break
        elif svc.desired_replicas > 0 and svc.running_tasks < svc.desired_replicas:
            score += 0.10
            factors.append(f"{svc.name}: partial replicas")

    # Crash loops
    for svc in req.services:
        if svc.restart_count_max >= 10:
            score += 0.15
            factors.append(f"{svc.name}: crash loop ({svc.restart_count_max} restarts)")
            break

    # Update paused
    for svc in req.services:
        if svc.update_state in ("paused", "rollback_started", "rollback_paused"):
            score += 0.10
            factors.append(f"{svc.name}: update {svc.update_state}")

    # Node pressure
    pressure_count = sum(
        1 for n in req.nodes
        if n.cpu_reservation_pct >= 0.85 or n.mem_reservation_pct >= 0.85
    )
    if pressure_count > 0:
        score += min(0.10 * pressure_count, 0.20)
        factors.append(f"{pressure_count} node(s) under resource pressure")

    # Drained nodes
    drained = sum(1 for n in req.nodes if n.availability == "drain")
    if drained > 0:
        score += min(0.05 * drained, 0.15)
        factors.append(f"{drained} node(s) draining")

    score = min(round(score, 3), 1.0)
    confidence = 0.85 if len(factors) >= 2 else 0.60

    return {
        "score": score,
        "confidence": confidence,
        "factors": factors,
        "source": "predictor",
    }
