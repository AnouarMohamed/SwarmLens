// Package plugins contains all SwarmLens diagnostic plugins.
// Register() returns the full list to pass to intelligence.New().
package plugins

import "github.com/AnouarMohamed/swarmlens/backend/internal/intelligence"

// Register returns all v1 diagnostic plugins.
func Register() []intelligence.Plugin {
	return []intelligence.Plugin{
		&ReplicaMismatch{},
		&PlacementFailure{},
		&CrashLoop{},
		&ImagePullFailure{},
		&PortConflict{},
		&SecretConfigRef{},
		&QuorumRisk{},
		&UpdateRollbackState{},
		&NodePressure{},
	}
}
