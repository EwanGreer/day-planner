package core

import "time"

// PlannerService defines the core day-planning operations.
type PlannerService interface {
	StartDay(date time.Time) (*DayPlan, error)
	GetCurrentPlan() (*DayPlan, error)
	CompleteTask(taskID string) error
	ScoreDay(plan *DayPlan) (int, error)
}
