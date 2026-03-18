package docker

import (
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// DemoSnapshot returns a realistic Swarm snapshot for demo mode.
// It is pre-loaded with findings triggers across all 9 diagnostic plugins.
func DemoSnapshot() model.Snapshot {
	return model.Snapshot{
		Managers: 2,
		Workers:  3,
		Nodes: []model.Node{
			{
				ID: "node-mgr-01", Hostname: "manager-01", Role: "manager",
				Availability: "active", State: "ready",
				CPUTotal: 4_000_000_000, CPUReserved: 1_200_000_000,
				MemTotal: 8 * 1024 * 1024 * 1024, MemReserved: 3 * 1024 * 1024 * 1024,
				Labels:        map[string]string{"zone": "us-east-1"},
				ManagerStatus: &model.ManagerStatus{Leader: true, Reachability: "reachable"},
				EngineVersion: "27.0.0", Addr: "10.0.0.1:2377",
			},
			{
				ID: "node-mgr-02", Hostname: "manager-02", Role: "manager",
				Availability: "active", State: "ready",
				CPUTotal: 4_000_000_000, CPUReserved: 800_000_000,
				MemTotal: 8 * 1024 * 1024 * 1024, MemReserved: 2 * 1024 * 1024 * 1024,
				Labels:        map[string]string{"zone": "us-east-1"},
				ManagerStatus: &model.ManagerStatus{Leader: false, Reachability: "reachable"},
				EngineVersion: "27.0.0", Addr: "10.0.0.2:2377",
			},
			{
				// triggers: node-pressure (CPU 87%, Mem 87%)
				ID: "node-wkr-01", Hostname: "worker-01", Role: "worker",
				Availability: "active", State: "ready",
				CPUTotal: 8_000_000_000, CPUReserved: 7_000_000_000,
				MemTotal: 16 * 1024 * 1024 * 1024, MemReserved: 14 * 1024 * 1024 * 1024,
				Labels: map[string]string{"zone": "us-west-2"},
			},
			{
				// triggers: drained node visibility
				ID: "node-wkr-02", Hostname: "worker-02", Role: "worker",
				Availability: "drain", State: "ready",
				Labels: map[string]string{"zone": "us-east-1"},
			},
			{
				ID: "node-wkr-03", Hostname: "worker-03", Role: "worker",
				Availability: "active", State: "ready",
				CPUTotal: 4_000_000_000, CPUReserved: 500_000_000,
				MemTotal: 8 * 1024 * 1024 * 1024, MemReserved: 1 * 1024 * 1024 * 1024,
				Labels: map[string]string{"zone": "us-east-1"},
			},
		},
		Services: []model.Service{
			{
				// healthy
				ID: "svc-api-01", Name: "api", Stack: "payments",
				Image: "acme/payments-api:v2.1.0", Mode: "replicated",
				DesiredReplicas: 3, RunningTasks: 3, UpdateState: "completed",
				PublishedPorts: []model.PublishedPort{{PublishedPort: 8080, TargetPort: 8080, Protocol: "tcp"}},
				NetworkRefs:    []string{"payments_default"},
				SecretRefs:     []string{"db_password"},
				CreatedAt:      time.Now().Add(-48 * time.Hour),
				UpdatedAt:      time.Now().Add(-2 * time.Hour),
			},
			{
				// triggers: replica-mismatch (0/2) + crash-loop
				ID: "svc-worker-01", Name: "worker", Stack: "payments",
				Image: "acme/payments-worker:v2.1.0", Mode: "replicated",
				DesiredReplicas: 2, RunningTasks: 0, FailedTasks: 8, UpdateState: "completed",
				NetworkRefs: []string{"payments_default"},
				SecretRefs:  []string{"db_password", "redis_url"},
				CreatedAt:   time.Now().Add(-48 * time.Hour),
				UpdatedAt:   time.Now().Add(-30 * time.Minute),
			},
			{
				// triggers: secret-config-ref (smtp_password missing) + image-pull-failure
				ID: "svc-notifier-01", Name: "notifier", Stack: "payments",
				Image: "acme/notifier:latest", Mode: "replicated",
				DesiredReplicas: 1, RunningTasks: 0, UpdateState: "completed",
				SecretRefs: []string{"smtp_password"},
				CreatedAt:  time.Now().Add(-24 * time.Hour),
				UpdatedAt:  time.Now().Add(-1 * time.Hour),
			},
			{
				// triggers: update-rollback-state (paused)
				ID: "svc-frontend-01", Name: "frontend", Stack: "web",
				Image: "acme/frontend:v3.0.0", Mode: "replicated",
				DesiredReplicas: 2, RunningTasks: 1, UpdateState: "paused",
				PublishedPorts: []model.PublishedPort{{PublishedPort: 443, TargetPort: 3000, Protocol: "tcp"}},
				CreatedAt:      time.Now().Add(-72 * time.Hour),
				UpdatedAt:      time.Now().Add(-10 * time.Minute),
			},
			{
				// triggers: placement-failure (gpu label missing)
				ID: "svc-ml-01", Name: "ml-inference", Stack: "ml",
				Image: "acme/ml:v1.0.0", Mode: "replicated",
				DesiredReplicas: 2, RunningTasks: 0, UpdateState: "completed",
				Constraints: []string{"node.labels.gpu==true"},
				CreatedAt:   time.Now().Add(-1 * time.Hour),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			},
		},
		Tasks: []model.Task{
			// payments/api healthy tasks
			{ID: "task-01", ServiceID: "svc-api-01", ServiceName: "payments_api", NodeID: "node-mgr-01", NodeHostname: "manager-01", DesiredState: "running", CurrentState: "running"},
			{ID: "task-02", ServiceID: "svc-api-01", ServiceName: "payments_api", NodeID: "node-wkr-01", NodeHostname: "worker-01", DesiredState: "running", CurrentState: "running"},
			{ID: "task-03", ServiceID: "svc-api-01", ServiceName: "payments_api", NodeID: "node-wkr-03", NodeHostname: "worker-03", DesiredState: "running", CurrentState: "running"},
			// payments/worker crash-looping
			{ID: "task-04", ServiceID: "svc-worker-01", ServiceName: "payments_worker", NodeID: "node-wkr-03", NodeHostname: "worker-03", DesiredState: "running", CurrentState: "failed", ExitCode: 1, Error: "task: non-zero exit (1)", RestartCount: 14},
			{ID: "task-05", ServiceID: "svc-worker-01", ServiceName: "payments_worker", NodeID: "node-mgr-01", NodeHostname: "manager-01", DesiredState: "running", CurrentState: "failed", ExitCode: 1, Error: "task: non-zero exit (1)", RestartCount: 12},
			// payments/notifier image pull failure
			{ID: "task-06", ServiceID: "svc-notifier-01", ServiceName: "payments_notifier", NodeID: "node-wkr-01", NodeHostname: "worker-01", DesiredState: "running", CurrentState: "failed", Image: "acme/notifier:latest", Error: "No such image: acme/notifier:latest — pull access denied, repository does not exist or may require authentication", RestartCount: 3},
		},
		Networks: []model.Network{
			{ID: "net-01", Name: "payments_default", Driver: "overlay", Scope: "swarm", Subnet: "10.10.0.0/24"},
			{ID: "net-02", Name: "ingress", Driver: "overlay", Scope: "swarm", Ingress: true},
		},
		Volumes: []model.Volume{
			{Name: "postgres_data", Driver: "local", Scope: "local"},
		},
		Secrets: []model.Secret{
			{ID: "secret-01", Name: "db_password", CreatedAt: time.Now().Add(-72 * time.Hour), ServiceRefs: []string{"payments_api", "payments_worker"}},
			{ID: "secret-02", Name: "redis_url", CreatedAt: time.Now().Add(-72 * time.Hour), ServiceRefs: []string{"payments_worker"}},
			// smtp_password intentionally missing — triggers secret-config-ref plugin
		},
		Configs: []model.Config{
			{ID: "config-01", Name: "nginx_config", CreatedAt: time.Now().Add(-48 * time.Hour)},
		},
	}
}

// DemoEvents returns a set of realistic recent Swarm events for demo mode.
func DemoEvents() []model.SwarmEvent {
	base := time.Now().UTC()
	return []model.SwarmEvent{
		{Type: "service", Action: "update", Actor: "payments_worker", Message: "service update failed: task failure ratio exceeded", Timestamp: base.Add(-2 * time.Minute)},
		{Type: "task", Action: "failed", Actor: "payments_worker.1", Message: "task: non-zero exit (1)", Timestamp: base.Add(-90 * time.Second)},
		{Type: "service", Action: "update", Actor: "web_frontend", Message: "service update paused", Timestamp: base.Add(-5 * time.Minute)},
		{Type: "node", Action: "update", Actor: "worker-02", Message: "node availability changed to drain", Timestamp: base.Add(-10 * time.Minute)},
		{Type: "service", Action: "create", Actor: "ml_ml-inference", Message: "service created", Timestamp: base.Add(-15 * time.Minute)},
		{Type: "task", Action: "failed", Actor: "payments_notifier.1", Message: "pull access denied, repository does not exist", Timestamp: base.Add(-20 * time.Minute)},
		{Type: "node", Action: "join", Actor: "worker-03", Message: "node joined swarm as worker", Timestamp: base.Add(-2 * time.Hour)},
	}
}
