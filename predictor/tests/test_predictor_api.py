from fastapi.testclient import TestClient

import app.main as main_module
from app.main import app


def test_health_endpoint():
    client = TestClient(app)
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"


def test_score_endpoint_returns_risk_payload():
    client = TestClient(app)
    payload = {
        "managers": 1,
        "workers": 2,
        "services": [
            {
                "name": "payments-worker",
                "desired_replicas": 2,
                "running_tasks": 0,
                "failed_tasks": 1,
                "restart_count_max": 12,
                "update_state": "paused",
            }
        ],
        "nodes": [
            {
                "hostname": "worker-01",
                "role": "worker",
                "availability": "active",
                "state": "ready",
                "cpu_reservation_pct": 0.9,
                "mem_reservation_pct": 0.7,
            }
        ],
    }
    response = client.post("/score", json=payload)
    body = response.json()
    assert response.status_code == 200
    assert "score" in body
    assert "confidence" in body
    assert "factors" in body
    assert body["source"] == "predictor"
    assert body["score"] > 0


def test_score_endpoint_enforces_shared_secret(monkeypatch):
    monkeypatch.setattr(main_module, "SHARED_SECRET", "top-secret")
    client = TestClient(app)
    response = client.post("/score", json={"managers": 1})
    assert response.status_code == 401
    assert response.json()["detail"] == "invalid shared secret"

    allowed = client.post(
        "/score",
        json={"managers": 1},
        headers={"x-shared-secret": "top-secret"},
    )
    assert allowed.status_code == 200


def test_read_secret_uses_file(tmp_path, monkeypatch):
    secret_file = tmp_path / "predictor_secret"
    secret_file.write_text("file-secret\n", encoding="utf-8")
    monkeypatch.delenv("PREDICTOR_SHARED_SECRET", raising=False)
    monkeypatch.setenv("PREDICTOR_SHARED_SECRET_FILE", str(secret_file))

    assert main_module._read_secret("PREDICTOR_SHARED_SECRET") == "file-secret"
