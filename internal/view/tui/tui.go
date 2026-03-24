package tui

import (
	"fmt"

	"github.com/EwanGreer/day-planner/internal/core"
)

// TUI implements view.Presenter by printing to stdout.
// This is a stub; a full terminal UI will be added in a later issue.
type TUI struct{}

// New returns a new TUI presenter.
func New() *TUI {
	return &TUI{}
}

// ShowDayPlan prints a summary of the day plan to stdout.
func (t *TUI) ShowDayPlan(plan *core.DayPlan) error {
	fmt.Printf("Day Plan for %s — %d task(s), score: %d\n",
		plan.Date.Format("2006-01-02"), len(plan.Tasks), plan.Score)
	return nil
}

// ShowTasks prints each task to stdout.
func (t *TUI) ShowTasks(tasks []core.Task) error {
	for _, task := range tasks {
		fmt.Printf("  [%s] %s (priority: %d, source: %s)\n",
			task.Status, task.Title, task.Priority, task.Source)
	}
	return nil
}

// ShowMessage prints a plain message to stdout.
func (t *TUI) ShowMessage(msg string) error {
	fmt.Println(msg)
	return nil
}

// Close is a no-op for the stdout stub.
func (t *TUI) Close() error {
	return nil
}
