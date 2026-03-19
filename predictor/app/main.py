"""SwarmLens predictor service — optional risk scoring for Swarm clusters."""
from fastapi import FastAPI, Header, HTTPException, status
from fastapi.responses import JSONResponse
import os

from .models import ScoreRequest
from .scorer import score_snapshot

app = FastAPI(title="SwarmLens Predictor", version="0.1.0")


def _read_secret(name: str) -> str:
    direct = os.environ.get(name, "").strip()
    if direct:
        return direct
    path = os.environ.get(f"{name}_FILE", "").strip()
    if not path:
        return ""
    try:
        with open(path, "r", encoding="utf-8") as fh:
            return fh.read().strip()
    except OSError:
        return ""


SHARED_SECRET = _read_secret("PREDICTOR_SHARED_SECRET")


def _check_secret(x_shared_secret: str | None) -> None:
    if SHARED_SECRET and x_shared_secret != SHARED_SECRET:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="invalid shared secret")


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/score")
def score(
    request: ScoreRequest,
    x_shared_secret: str | None = Header(default=None),
):
    _check_secret(x_shared_secret)
    result = score_snapshot(request)
    return JSONResponse(content=result)
