package taskwarrior

import "github.com/EwanGreer/day-planner/internal/core"

// Adapter implements integrations.Integration for Taskwarrior.
type Adapter struct{}

// New returns a new Taskwarrior Adapter.
func New() *Adapter {
	return &Adapter{}
}

// Name returns the integration name.
func (a *Adapter) Name() string {
	return "taskwarrior"
}

// IsAvailable reports whether the taskwarrior binary is present and callable.
// Stub: always returns false until the real check is implemented.
func (a *Adapter) IsAvailable() bool {
	return false
}

// FetchTasks retrieves pending tasks from Taskwarrior.
// Stub: returns an empty slice until the real implementation is added.
func (a *Adapter) FetchTasks() ([]core.Task, error) {
	return []core.Task{}, nil
}
