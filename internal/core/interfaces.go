package core

// TaskProvider is the interface integration adapters must satisfy to provide tasks.
// Defined here so core can depend on it without importing integration packages.
//
// Note: the name TaskSource is already taken as a string type in domain.go
// (it identifies the origin of a task, e.g. "taskwarrior"). TaskProvider is
// used here to avoid a redeclaration conflict.
type TaskProvider interface {
	// Name returns the human-readable name of the provider.
	Name() string
	// FetchTasks retrieves tasks from the external source.
	FetchTasks() ([]Task, error)
	// IsAvailable reports whether the provider is configured and reachable.
	IsAvailable() bool
}

// PlanPresenter is the subset of view.Presenter used by PlanningService.
// Defined here so core can depend on it without importing the view package.
type PlanPresenter interface {
	// ShowTaskList displays the prioritised list of tasks.
	ShowTaskList(tasks []Task) error
	// PromptCommitments asks the user to select tasks to commit to.
	PromptCommitments(tasks []Task) ([]Task, error)
	// ShowPlanConfirmation displays the confirmed plan and current streak.
	ShowPlanConfirmation(plan *DayPlan, streak *Streak) error
	// ShowMessage displays a plain informational message.
	ShowMessage(msg string) error
	// ShowError displays an error to the user.
	ShowError(err error)
}
