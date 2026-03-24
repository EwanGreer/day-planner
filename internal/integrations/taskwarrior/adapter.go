package taskwarrior

import (
	"encoding/json"
	"errors"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/EwanGreer/day-planner/internal/core"
	"github.com/EwanGreer/day-planner/internal/integrations"
)

var _ integrations.Integration = (*Adapter)(nil)

// twTask mirrors the JSON fields emitted by `task export`.
type twTask struct {
	UUID        string   `json:"uuid"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Urgency     float64  `json:"urgency"`
	Priority    string   `json:"priority"`
	Project     string   `json:"project"`
	Tags        []string `json:"tags"`
	Due         string   `json:"due"`
}

// Adapter implements integrations.Integration for Taskwarrior.
type Adapter struct {
	cmd Commander
}

// New returns a new Taskwarrior Adapter backed by the real os/exec Commander.
func New() *Adapter {
	return &Adapter{cmd: ExecCommander{}}
}

// NewWithCommander returns an Adapter that uses cmd to run external processes.
// Intended for use in tests to inject a fake Commander.
func NewWithCommander(cmd Commander) *Adapter {
	return &Adapter{cmd: cmd}
}

// Name returns the integration name.
func (a *Adapter) Name() string {
	return "taskwarrior"
}

// IsAvailable reports whether the taskwarrior binary is present in PATH.
func (a *Adapter) IsAvailable() bool {
	_, err := exec.LookPath("task")
	return err == nil
}

// FetchTasks retrieves pending tasks from Taskwarrior sorted by urgency descending.
//
// If the `task` binary is not found (exec.ErrNotFound or "command not found" in
// the error message), FetchTasks returns nil, nil — treating missing Taskwarrior
// as an empty integration rather than an error.
func (a *Adapter) FetchTasks() ([]core.Task, error) {
	out, err := a.cmd.Run("task", "status:pending", "export")
	if err != nil {
		msg := err.Error()
		if errors.Is(err, exec.ErrNotFound) ||
			strings.Contains(msg, "executable file not found") ||
			strings.Contains(msg, "command not found") {
			return nil, nil
		}
		// Exit errors from `task` itself (e.g. no tasks) still produce valid JSON
		// on stdout; unwrap to check whether we still got output.
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return nil, err
		}
		// If stdout is empty after a non-zero exit, propagate the error.
		if len(out) == 0 {
			return nil, err
		}
	}

	var raw []twTask
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}

	tasks := make([]core.Task, 0, len(raw))
	for _, t := range raw {
		tasks = append(tasks, mapTask(t))
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Urgency > tasks[j].Urgency
	})

	return tasks, nil
}

// mapTask converts a Taskwarrior JSON task into a core.Task.
func mapTask(t twTask) core.Task {
	return core.Task{
		ID:       t.UUID,
		Title:    t.Description,
		Status:   core.TaskStatusPending,
		Source:   core.TaskSourceTaskwarrior,
		Urgency:  t.Urgency,
		Priority: mapPriority(t.Priority),
		Project:  t.Project,
		Tags:     t.Tags,
		Due:      parseDue(t.Due),
	}
}

// mapPriority converts a Taskwarrior priority string to an integer.
func mapPriority(p string) int {
	switch p {
	case "H":
		return 3
	case "M":
		return 2
	case "L":
		return 1
	default:
		return 0
	}
}

// taskwarriorTimeFormat is the compact UTC format Taskwarrior uses in its JSON export.
const taskwarriorTimeFormat = "20060102T150405Z"

// parseDue parses a Taskwarrior due date string, trying RFC3339 first, then the
// Taskwarrior compact format. Returns nil when the string is empty or unparseable.
func parseDue(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, taskwarriorTimeFormat} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}
