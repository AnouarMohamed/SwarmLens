package auth

import "errors"

var (
	ErrWritesDisabled = errors.New("write actions are disabled (WRITE_ACTIONS_ENABLED=false)")
)

// Gate enforces the global write gate.
type Gate struct {
	enabled bool
}

// NewGate creates a Gate from the WRITE_ACTIONS_ENABLED setting.
func NewGate(enabled bool) *Gate {
	return &Gate{enabled: enabled}
}

// Check returns an error if writes are not enabled.
func (g *Gate) Check() error {
	if !g.enabled {
		return ErrWritesDisabled
	}
	return nil
}

// Enabled reports whether writes are allowed.
func (g *Gate) Enabled() bool { return g.enabled }
