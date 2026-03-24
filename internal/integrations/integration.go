package integrations

import "github.com/EwanGreer/day-planner/internal/core"

// Integration is the contract all external task-source adapters must satisfy.
type Integration interface {
	// Name returns the human-readable name of the integration.
	Name() string
	// FetchTasks retrieves tasks from the external source.
	FetchTasks() ([]core.Task, error)
	// IsAvailable reports whether the integration is configured and reachable.
	IsAvailable() bool
}
