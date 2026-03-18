from pydantic import BaseModel
from typing import Optional


class ServiceSignal(BaseModel):
    name: str
    desired_replicas: int = 0
    running_tasks: int = 0
    failed_tasks: int = 0
    restart_count_max: int = 0
    update_state: str = ""


class NodeSignal(BaseModel):
    hostname: str
    role: str = "worker"
    availability: str = "active"
    state: str = "ready"
    cpu_reservation_pct: float = 0.0
    mem_reservation_pct: float = 0.0


class ScoreRequest(BaseModel):
    managers: int = 1
    workers: int = 0
    services: list[ServiceSignal] = []
    nodes: list[NodeSignal] = []
