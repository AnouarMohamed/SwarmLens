package docker

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// Snapshot fetches the current cluster inventory and recent events.
func (c *Client) Snapshot(ctx context.Context) (model.Snapshot, []model.SwarmEvent, error) {
	if c.demo {
		return DemoSnapshot(), DemoEvents(), nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	nodes, err := c.docker.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return model.Snapshot{}, nil, fmt.Errorf("list nodes: %w", err)
	}
	services, err := c.docker.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return model.Snapshot{}, nil, fmt.Errorf("list services: %w", err)
	}
	tasks, err := c.docker.TaskList(ctx, types.TaskListOptions{})
	if err != nil {
		return model.Snapshot{}, nil, fmt.Errorf("list tasks: %w", err)
	}
	networks, err := c.docker.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return model.Snapshot{}, nil, fmt.Errorf("list networks: %w", err)
	}
	volumesResp, err := c.docker.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return model.Snapshot{}, nil, fmt.Errorf("list volumes: %w", err)
	}
	secrets, err := c.docker.SecretList(ctx, types.SecretListOptions{})
	if err != nil {
		return model.Snapshot{}, nil, fmt.Errorf("list secrets: %w", err)
	}
	configs, err := c.docker.ConfigList(ctx, types.ConfigListOptions{})
	if err != nil {
		return model.Snapshot{}, nil, fmt.Errorf("list configs: %w", err)
	}

	snap := model.Snapshot{
		Nodes:    make([]model.Node, 0, len(nodes)),
		Services: make([]model.Service, 0, len(services)),
		Tasks:    make([]model.Task, 0, len(tasks)),
		Networks: make([]model.Network, 0, len(networks)),
		Volumes:  make([]model.Volume, 0, len(volumesResp.Volumes)),
		Secrets:  make([]model.Secret, 0, len(secrets)),
		Configs:  make([]model.Config, 0, len(configs)),
	}

	nodeHostnameByID := make(map[string]string, len(nodes))
	for _, n := range nodes {
		nodeModel := toModelNode(n)
		snap.Nodes = append(snap.Nodes, nodeModel)
		nodeHostnameByID[n.ID] = nodeModel.Hostname
		if nodeModel.Role == "manager" {
			snap.Managers++
		} else {
			snap.Workers++
		}
	}

	taskStatsByService := make(map[string]struct {
		running  int
		failed   int
		restarts int
	})
	for _, t := range tasks {
		modelTask := toModelTask(t, nodeHostnameByID)
		snap.Tasks = append(snap.Tasks, modelTask)
		stats := taskStatsByService[t.ServiceID]
		if modelTask.CurrentState == "running" {
			stats.running++
		}
		if modelTask.CurrentState == "failed" || modelTask.CurrentState == "rejected" {
			stats.failed++
		}
		stats.restarts += modelTask.RestartCount
		taskStatsByService[t.ServiceID] = stats
	}

	serviceRefsBySecret := map[string][]string{}
	serviceRefsByConfig := map[string][]string{}
	serviceRefsByNetwork := map[string]map[string]struct{}{}

	for _, s := range services {
		serviceModel := toModelService(s, taskStatsByService[s.ID])
		snap.Services = append(snap.Services, serviceModel)
		for _, secret := range serviceModel.SecretRefs {
			serviceRefsBySecret[secret] = append(serviceRefsBySecret[secret], serviceModel.Name)
		}
		for _, cfg := range serviceModel.ConfigRefs {
			serviceRefsByConfig[cfg] = append(serviceRefsByConfig[cfg], serviceModel.Name)
		}
		for _, netRef := range serviceModel.NetworkRefs {
			if _, ok := serviceRefsByNetwork[netRef]; !ok {
				serviceRefsByNetwork[netRef] = map[string]struct{}{}
			}
			serviceRefsByNetwork[netRef][serviceModel.ID] = struct{}{}
		}
	}

	for _, n := range networks {
		serviceCount := 0
		if refs, ok := serviceRefsByNetwork[n.ID]; ok {
			serviceCount = len(refs)
		}
		if refs, ok := serviceRefsByNetwork[n.Name]; ok && len(refs) > serviceCount {
			serviceCount = len(refs)
		}
		snap.Networks = append(snap.Networks, toModelNetwork(n, serviceCount))
	}

	for _, v := range volumesResp.Volumes {
		if v == nil {
			continue
		}
		snap.Volumes = append(snap.Volumes, model.Volume{
			Name:       v.Name,
			Driver:     v.Driver,
			Scope:      v.Scope,
			Mountpoint: v.Mountpoint,
			Labels:     v.Labels,
		})
	}

	for _, s := range secrets {
		snap.Secrets = append(snap.Secrets, model.Secret{
			ID:          s.ID,
			Name:        s.Spec.Name,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
			ServiceRefs: dedupe(serviceRefsBySecret[s.Spec.Name]),
		})
	}
	for _, cfg := range configs {
		snap.Configs = append(snap.Configs, model.Config{
			ID:          cfg.ID,
			Name:        cfg.Spec.Name,
			CreatedAt:   cfg.CreatedAt,
			UpdatedAt:   cfg.UpdatedAt,
			ServiceRefs: dedupe(serviceRefsByConfig[cfg.Spec.Name]),
		})
	}

	eventsList := c.recentEvents(ctx, 45*time.Minute)
	return snap, eventsList, nil
}

func toModelNode(n swarm.Node) model.Node {
	role := string(n.Spec.Role)
	m := model.Node{
		ID:            n.ID,
		Hostname:      n.Description.Hostname,
		Role:          role,
		Availability:  string(n.Spec.Availability),
		State:         string(n.Status.State),
		Labels:        n.Spec.Labels,
		CPUTotal:      n.Description.Resources.NanoCPUs,
		MemTotal:      n.Description.Resources.MemoryBytes,
		EngineVersion: n.Description.Engine.EngineVersion,
		Addr:          n.Status.Addr,
	}
	if n.ManagerStatus != nil {
		m.ManagerStatus = &model.ManagerStatus{
			Leader:       n.ManagerStatus.Leader,
			Reachability: string(n.ManagerStatus.Reachability),
		}
	}
	return m
}

func toModelTask(t swarm.Task, nodeHostnameByID map[string]string) model.Task {
	state := strings.ToLower(string(t.Status.State))
	desired := strings.ToLower(string(t.DesiredState))
	exitCode := 0
	if t.Status.ContainerStatus != nil {
		exitCode = t.Status.ContainerStatus.ExitCode
	}
	image := ""
	if t.Spec.ContainerSpec != nil {
		image = t.Spec.ContainerSpec.Image
	}
	errMsg := t.Status.Err
	if errMsg == "" {
		errMsg = t.Status.Message
	}

	return model.Task{
		ID:           t.ID,
		ServiceID:    t.ServiceID,
		ServiceName:  t.Annotations.Name,
		NodeID:       t.NodeID,
		NodeHostname: nodeHostnameByID[t.NodeID],
		DesiredState: desired,
		CurrentState: state,
		ExitCode:     exitCode,
		Error:        errMsg,
		Image:        image,
		RestartCount: extractRestartCount(t.Status.Message),
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

func toModelService(s swarm.Service, stats struct {
	running  int
	failed   int
	restarts int
}) model.Service {
	mode := "replicated"
	desired := 0
	if s.Spec.Mode.Global != nil {
		mode = "global"
	}
	if s.Spec.Mode.Replicated != nil && s.Spec.Mode.Replicated.Replicas != nil {
		desired = int(*s.Spec.Mode.Replicated.Replicas)
	}
	if s.ServiceStatus != nil && s.ServiceStatus.DesiredTasks > 0 {
		desired = int(s.ServiceStatus.DesiredTasks)
	}

	running := stats.running
	if s.ServiceStatus != nil && s.ServiceStatus.RunningTasks > 0 {
		running = int(s.ServiceStatus.RunningTasks)
	}

	updateState := ""
	if s.UpdateStatus != nil {
		updateState = string(s.UpdateStatus.State)
	}
	image := ""
	var secretRefs []string
	var configRefs []string
	var constraints []string
	var preferences []string
	var networkRefs []string
	if s.Spec.TaskTemplate.ContainerSpec != nil {
		image = s.Spec.TaskTemplate.ContainerSpec.Image
		for _, sec := range s.Spec.TaskTemplate.ContainerSpec.Secrets {
			if sec != nil && sec.SecretName != "" {
				secretRefs = append(secretRefs, sec.SecretName)
			}
		}
		for _, cfg := range s.Spec.TaskTemplate.ContainerSpec.Configs {
			if cfg != nil && cfg.ConfigName != "" {
				configRefs = append(configRefs, cfg.ConfigName)
			}
		}
	}
	if s.Spec.TaskTemplate.Placement != nil {
		constraints = append(constraints, s.Spec.TaskTemplate.Placement.Constraints...)
		for _, pref := range s.Spec.TaskTemplate.Placement.Preferences {
			if pref.Spread != nil && pref.Spread.SpreadDescriptor != "" {
				preferences = append(preferences, pref.Spread.SpreadDescriptor)
			}
		}
	}
	for _, nw := range s.Spec.TaskTemplate.Networks {
		if nw.Target != "" {
			networkRefs = append(networkRefs, nw.Target)
		}
	}
	for _, nw := range s.Spec.Networks {
		if nw.Target != "" {
			networkRefs = append(networkRefs, nw.Target)
		}
	}

	ports := make([]model.PublishedPort, 0)
	if s.Spec.EndpointSpec != nil {
		for _, p := range s.Spec.EndpointSpec.Ports {
			ports = append(ports, model.PublishedPort{
				PublishedPort: p.PublishedPort,
				TargetPort:    p.TargetPort,
				Protocol:      string(p.Protocol),
			})
		}
	}

	updateParallel := uint64(0)
	updateDelay := ""
	updateFailure := ""
	if s.Spec.UpdateConfig != nil {
		updateParallel = s.Spec.UpdateConfig.Parallelism
		updateDelay = s.Spec.UpdateConfig.Delay.String()
		updateFailure = string(s.Spec.UpdateConfig.FailureAction)
	}
	rollbackParallel := uint64(0)
	rollbackDelay := ""
	if s.Spec.RollbackConfig != nil {
		rollbackParallel = s.Spec.RollbackConfig.Parallelism
		rollbackDelay = s.Spec.RollbackConfig.Delay.String()
	}

	return model.Service{
		ID:                  s.ID,
		Name:                s.Spec.Name,
		Stack:               s.Spec.Labels["com.docker.stack.namespace"],
		Image:               image,
		Mode:                mode,
		DesiredReplicas:     desired,
		RunningTasks:        running,
		FailedTasks:         stats.failed,
		UpdateState:         updateState,
		UpdateParallelism:   updateParallel,
		UpdateDelay:         updateDelay,
		UpdateFailureAction: updateFailure,
		RollbackParallelism: rollbackParallel,
		RollbackDelay:       rollbackDelay,
		Constraints:         dedupe(constraints),
		Preferences:         dedupe(preferences),
		PublishedPorts:      ports,
		SecretRefs:          dedupe(secretRefs),
		ConfigRefs:          dedupe(configRefs),
		NetworkRefs:         dedupe(networkRefs),
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
	}
}

func toModelNetwork(n network.Summary, serviceCount int) model.Network {
	subnet := ""
	if len(n.IPAM.Config) > 0 {
		subnet = n.IPAM.Config[0].Subnet
	}
	return model.Network{
		ID:           n.ID,
		Name:         n.Name,
		Driver:       n.Driver,
		Scope:        n.Scope,
		Subnet:       subnet,
		Attachable:   n.Attachable,
		Ingress:      n.Ingress,
		ServiceCount: serviceCount,
	}
}

func (c *Client) recentEvents(ctx context.Context, window time.Duration) []model.SwarmEvent {
	if c.demo {
		return DemoEvents()
	}
	since := strconv.FormatInt(time.Now().Add(-window).Unix(), 10)
	until := strconv.FormatInt(time.Now().Unix(), 10)
	f := filters.NewArgs()
	f.Add("scope", "swarm")
	opts := events.ListOptions{Since: since, Until: until, Filters: f}

	eventCh, errCh := c.docker.Events(ctx, opts)
	collected := make([]model.SwarmEvent, 0, 64)
	for {
		select {
		case evt, ok := <-eventCh:
			if !ok {
				return sortEvents(collected)
			}
			collected = append(collected, model.SwarmEvent{
				Type:      string(evt.Type),
				Action:    string(evt.Action),
				Actor:     eventActor(evt),
				Message:   eventMessage(evt),
				Timestamp: eventTimestamp(evt),
			})
		case err, ok := <-errCh:
			if !ok || err == nil {
				return sortEvents(collected)
			}
			return sortEvents(collected)
		case <-ctx.Done():
			return sortEvents(collected)
		}
	}
}

func eventActor(evt events.Message) string {
	if name := evt.Actor.Attributes["name"]; name != "" {
		return name
	}
	if svc := evt.Actor.Attributes["com.docker.swarm.service.name"]; svc != "" {
		return svc
	}
	if node := evt.Actor.Attributes["com.docker.swarm.node.id"]; node != "" {
		return node
	}
	if evt.Actor.ID != "" {
		return evt.Actor.ID
	}
	return "cluster"
}

func eventMessage(evt events.Message) string {
	if msg := evt.Actor.Attributes["message"]; msg != "" {
		return msg
	}
	if status := evt.Status; status != "" {
		return status
	}
	return string(evt.Action)
}

func eventTimestamp(evt events.Message) time.Time {
	if evt.TimeNano > 0 {
		return time.Unix(0, evt.TimeNano).UTC()
	}
	if evt.Time > 0 {
		return time.Unix(evt.Time, 0).UTC()
	}
	return time.Now().UTC()
}

func sortEvents(eventsList []model.SwarmEvent) []model.SwarmEvent {
	sort.Slice(eventsList, func(i, j int) bool {
		return eventsList[i].Timestamp.After(eventsList[j].Timestamp)
	})
	if len(eventsList) > 200 {
		return eventsList[:200]
	}
	return eventsList
}

func dedupe(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func extractRestartCount(msg string) int {
	// Docker task status does not expose restart count directly in swarm APIs.
	// This parser keeps best-effort parity with demo mode signals.
	if msg == "" {
		return 0
	}
	idx := strings.Index(strings.ToLower(msg), "restart")
	if idx < 0 {
		return 0
	}
	fields := strings.Fields(msg[idx:])
	for _, f := range fields {
		n, err := strconv.Atoi(strings.Trim(f, "(),.:"))
		if err == nil && n >= 0 {
			return n
		}
	}
	return 0
}
