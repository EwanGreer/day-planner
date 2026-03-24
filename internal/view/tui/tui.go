package tui

import (
	"fmt"

	"github.com/EwanGreer/day-planner/internal/core"
	"github.com/EwanGreer/day-planner/internal/view"
)

var _ view.Presenter = (*TUI)(nil)

// TUI implements view.Presenter by printing to stdout.
// This is a stub; a full terminal UI will be added in a later issue.
type TUI struct{}

// New returns a new TUI presenter.
func New() *TUI {
	return &TUI{}
}

// ShowTaskList prints a numbered list of tasks to stdout.
func (t *TUI) ShowTaskList(tasks []core.Task) error {
	fmt.Println("Tasks:")
	for i, task := range tasks {
		fmt.Printf("  %d. [%s] %s (priority: %d, source: %s)\n",
			i+1, task.Status, task.Title, task.Priority, task.Source)
	}
	return nil
}

// PromptCommitments prints tasks and returns the first min(3, len(tasks)) as a stub.
// A real interactive prompt will be added in issue #5.
func (t *TUI) PromptCommitments(tasks []core.Task) ([]core.Task, error) {
	fmt.Println("Select your commitments for today:")
	for i, task := range tasks {
		fmt.Printf("  %d. [%s] %s (priority: %d)\n",
			i+1, task.Status, task.Title, task.Priority)
	}
	limit := 3
	if len(tasks) < limit {
		limit = len(tasks)
	}
	return tasks[:limit], nil
}

// ShowPlanConfirmation prints the confirmed day plan and streak info.
func (t *TUI) ShowPlanConfirmation(plan *core.DayPlan, streak *core.Streak) error {
	if plan == nil {
		return fmt.Errorf("ShowPlanConfirmation: nil plan")
	}
	fmt.Printf("Plan confirmed for %s\n", plan.Date.Format("2006-01-02"))
	fmt.Println("Committed tasks:")
	for _, task := range plan.Tasks {
		fmt.Printf("  - [%s] %s\n", task.Status, task.Title)
	}
	if streak != nil {
		fmt.Printf("Current streak: %d day(s) (longest: %d)\n", streak.Current, streak.Longest)
	}
	return nil
}

// ShowNudge prints a mid-day reminder of the user's commitments.
func (t *TUI) ShowNudge(commitments []core.Task) error {
	fmt.Println("Reminder: your commitments for today:")
	for _, task := range commitments {
		fmt.Printf("  - [%s] %s\n", task.Status, task.Title)
	}
	return nil
}

// ShowEODSummary prints the end-of-day review summary.
func (t *TUI) ShowEODSummary(plan *core.DayPlan, score int) error {
	if plan == nil {
		return fmt.Errorf("ShowEODSummary: nil plan")
	}
	fmt.Printf("End of day summary for %s\n", plan.Date.Format("2006-01-02"))
	fmt.Println("Tasks:")
	for _, task := range plan.Tasks {
		fmt.Printf("  - [%s] %s\n", task.Status, task.Title)
	}
	fmt.Printf("Score: %d\n", score)
	return nil
}

// ShowMessage prints a plain message to stdout.
func (t *TUI) ShowMessage(msg string) error {
	fmt.Println(msg)
	return nil
}

// ShowError prints an error to stdout.
func (t *TUI) ShowError(err error) {
	fmt.Printf("Error: %v\n", err)
}

// Close is a no-op for the stdout stub.
func (t *TUI) Close() error {
	return nil
}
