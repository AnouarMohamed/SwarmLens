// Package docker provides the Docker Engine API client factory.
// Supports Unix socket and TCP+TLS connections to a Swarm manager.
package docker

import (
	"context"
	"fmt"
	"os"

	dockerclient "github.com/docker/docker/client"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
)

// Client wraps the Docker client with SwarmLens-specific helpers.
type Client struct {
	docker *dockerclient.Client
	demo   bool
}

// New creates a Docker client from config.
// In demo mode, returns a Client with demo=true; no real connection is made.
func New(cfg config.Config) (*Client, error) {
	if cfg.IsDemo() {
		return &Client{demo: true}, nil
	}

	opts := []dockerclient.Opt{
		dockerclient.WithHost(cfg.DockerHost),
		dockerclient.WithAPIVersionNegotiation(),
	}

	if cfg.DockerTLSVerify && cfg.DockerCertPath != "" {
		cacert := fmt.Sprintf("%s/ca.pem", cfg.DockerCertPath)
		cert := fmt.Sprintf("%s/cert.pem", cfg.DockerCertPath)
		key := fmt.Sprintf("%s/key.pem", cfg.DockerCertPath)

		if _, err := os.Stat(cacert); err != nil {
			return nil, fmt.Errorf("docker TLS: ca.pem not found at %s", cacert)
		}

		opts = append(opts, dockerclient.WithTLSClientConfig(cacert, cert, key))
	}

	dc, err := dockerclient.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}

	// Verify connection
	if _, err := dc.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("docker ping failed: %w", err)
	}

	return &Client{docker: dc}, nil
}

// Docker returns the underlying Docker client for use by inventory/actions packages.
func (c *Client) Docker() *dockerclient.Client {
	return c.docker
}

// IsDemo reports whether this client is in demo mode.
func (c *Client) IsDemo() bool {
	return c.demo
}
