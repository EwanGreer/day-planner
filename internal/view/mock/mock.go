package mock

import (
	"github.com/EwanGreer/day-planner/internal/core"
	"github.com/EwanGreer/day-planner/internal/view"
)

var _ view.Presenter = (*Presenter)(nil)

// Presenter is a test double for view.Presenter.
// It records all calls and returns configurable responses.
type Presenter struct {
	// Recorded calls
	TasksShown          []core.Task
	CommitmentsPrompted []core.Task
	PlanConfirmed       *core.DayPlan
	NudgesSent          [][]core.Task
	EODSummariesShown   []*core.DayPlan
	MessagesShown       []string
	ErrorsShown         []error
	Closed              bool

	// Configurable responses
	PromptCommitmentsResult []core.Task
	PromptCommitmentsErr    error
}

func (m *Presenter) ShowTaskList(tasks []core.Task) error {
	m.TasksShown = append(m.TasksShown, tasks...)
	return nil
}

func (m *Presenter) PromptCommitments(tasks []core.Task) ([]core.Task, error) {
	m.CommitmentsPrompted = append(m.CommitmentsPrompted, tasks...)
	if m.PromptCommitmentsResult != nil {
		return m.PromptCommitmentsResult, m.PromptCommitmentsErr
	}
	return tasks, m.PromptCommitmentsErr
}

func (m *Presenter) ShowPlanConfirmation(plan *core.DayPlan, streak *core.Streak) error {
	m.PlanConfirmed = plan
	return nil
}

func (m *Presenter) ShowNudge(commitments []core.Task) error {
	m.NudgesSent = append(m.NudgesSent, commitments)
	return nil
}

func (m *Presenter) ShowEODSummary(plan *core.DayPlan, score int) error {
	m.EODSummariesShown = append(m.EODSummariesShown, plan)
	return nil
}

func (m *Presenter) ShowMessage(msg string) error {
	m.MessagesShown = append(m.MessagesShown, msg)
	return nil
}

func (m *Presenter) ShowError(err error) {
	m.ErrorsShown = append(m.ErrorsShown, err)
}

func (m *Presenter) Close() error {
	m.Closed = true
	return nil
}
