package view

import "github.com/EwanGreer/day-planner/internal/core"

// Presenter is the contract all UI implementations must satisfy.
// It is defined here (internal/view) so it can reference core types
// without creating an import cycle.
type Presenter interface {
	// ShowTaskList displays the prioritised list of tasks.
	ShowTaskList(tasks []core.Task) error
	// PromptCommitments asks the user to select 2-3 tasks to commit to.
	// Returns the selected tasks.
	PromptCommitments(tasks []core.Task) ([]core.Task, error)
	// ShowPlanConfirmation displays the confirmed plan and current streak.
	ShowPlanConfirmation(plan *core.DayPlan, streak *core.Streak) error
	// ShowNudge displays a mid-day reminder of the user's commitments.
	ShowNudge(commitments []core.Task) error
	// ShowEODSummary displays the end-of-day review summary.
	ShowEODSummary(plan *core.DayPlan, score int) error
	// ShowMessage displays a plain informational message.
	ShowMessage(msg string) error
	// ShowError displays an error to the user.
	ShowError(err error)
	// Close releases any resources held by the presenter.
	Close() error
}
