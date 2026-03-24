package core

import "time"

// Store defines the full persistence contract for the day planner.
type Store interface {
	// DayPlan
	SaveDayPlan(plan *DayPlan) error
	LoadDayPlan(date time.Time) (*DayPlan, error)

	// CompletionRecord
	SaveCompletion(record CompletionRecord) error
	LoadCompletions(date time.Time) ([]CompletionRecord, error)

	// Streak
	SaveStreak(streak Streak) error
	LoadStreak() (*Streak, error)

	// Reflection
	SaveReflection(reflection Reflection) error
	LoadReflections(date time.Time) ([]Reflection, error)

	// NudgeWindow
	SaveNudgeWindow(window NudgeWindow) error
	LoadNudgeWindows() ([]NudgeWindow, error)

	// Lifecycle
	Close() error
}
