package core

import "time"

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

// TaskSource identifies where a task originated.
type TaskSource string

const (
	TaskSourceManual      TaskSource = "manual"
	TaskSourceTaskwarrior TaskSource = "taskwarrior"
	TaskSourceJira        TaskSource = "jira"
)

// Task is a unit of work within a day plan.
type Task struct {
	ID       string
	Title    string
	Priority int
	Status   TaskStatus
	Source   TaskSource
}

// DayPlan represents the full plan for a single day.
type DayPlan struct {
	Date      time.Time
	Tasks     []Task
	Goals     []string
	Score     int
	CreatedAt time.Time
}

// CompletionStatus represents the outcome of a task on a given day.
type CompletionStatus string

const (
	CompletionStatusComplete CompletionStatus = "complete"
	CompletionStatusSkipped  CompletionStatus = "skipped"
)

// CompletionRecord tracks the outcome of a task on a given day.
type CompletionRecord struct {
	Date        time.Time
	TaskID      string
	Status      CompletionStatus
	CompletedAt time.Time
}

// Streak tracks the user's consecutive active days.
type Streak struct {
	Current        int
	Longest        int
	LastActiveDate time.Time
}

// Reflection is a free-text end-of-day note.
type Reflection struct {
	Date      time.Time
	Text      string
	CreatedAt time.Time
}

// NudgeWindow defines a focus or DND time window.
type NudgeWindow struct {
	StartTime time.Time
	EndTime   time.Time
	Label     string // "focus" or "dnd"
}
