package view

import "github.com/EwanGreer/day-planner/internal/core"

// Presenter is the contract all UI implementations must satisfy.
type Presenter interface {
	// ShowDayPlan renders the full day plan to the user.
	ShowDayPlan(plan *core.DayPlan) error
	// ShowTasks renders a list of tasks to the user.
	ShowTasks(tasks []core.Task) error
	// ShowMessage displays an informational message to the user.
	ShowMessage(msg string) error
	// Close tears down any resources held by the presenter.
	Close() error
}
