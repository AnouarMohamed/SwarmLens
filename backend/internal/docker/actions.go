package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

func (c *Client) ServiceScale(ctx context.Context, serviceID string, replicas uint64) error {
	service, err := c.inspectService(ctx, serviceID)
	if err != nil {
		return err
	}
	spec := service.Spec
	if spec.Mode.Replicated == nil {
		spec.Mode.Replicated = &swarm.ReplicatedService{}
	}
	spec.Mode.Replicated.Replicas = &replicas
	_, err = c.docker.ServiceUpdate(ctx, service.ID, service.Version, spec, types.ServiceUpdateOptions{})
	if err != nil {
		return fmt.Errorf("scale service: %w", err)
	}
	return nil
}

func (c *Client) ServiceRestart(ctx context.Context, serviceID string) error {
	service, err := c.inspectService(ctx, serviceID)
	if err != nil {
		return err
	}
	spec := service.Spec
	spec.TaskTemplate.ForceUpdate++
	_, err = c.docker.ServiceUpdate(ctx, service.ID, service.Version, spec, types.ServiceUpdateOptions{})
	if err != nil {
		return fmt.Errorf("restart service: %w", err)
	}
	return nil
}

func (c *Client) ServiceRollback(ctx context.Context, serviceID string) error {
	service, err := c.inspectService(ctx, serviceID)
	if err != nil {
		return err
	}
	_, err = c.docker.ServiceUpdate(ctx, service.ID, service.Version, service.Spec, types.ServiceUpdateOptions{
		Rollback: "previous",
	})
	if err != nil {
		return fmt.Errorf("rollback service: %w", err)
	}
	return nil
}

func (c *Client) inspectService(ctx context.Context, serviceID string) (swarm.Service, error) {
	if c.demo {
		return swarm.Service{}, fmt.Errorf("service actions unavailable in demo client")
	}
	service, _, err := c.docker.ServiceInspectWithRaw(ctx, serviceID, types.ServiceInspectOptions{})
	if err == nil {
		return service, nil
	}

	services, listErr := c.docker.ServiceList(ctx, types.ServiceListOptions{})
	if listErr != nil {
		return swarm.Service{}, fmt.Errorf("inspect service: %w", err)
	}
	for _, candidate := range services {
		if candidate.ID == serviceID || candidate.Spec.Name == serviceID {
			return candidate, nil
		}
	}
	return swarm.Service{}, fmt.Errorf("service %q not found", serviceID)
}
