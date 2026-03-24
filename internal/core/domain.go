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
	Date  time.Time
	Tasks []Task
	Goals []string
	Score int
}
