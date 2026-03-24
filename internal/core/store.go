package core

import "time"

// Store defines the persistence contract for day plans.
type Store interface {
	SaveDayPlan(plan DayPlan) error
	LoadDayPlan(date time.Time) (*DayPlan, error)
}
