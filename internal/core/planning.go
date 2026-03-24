package core

import "time"

// PlanningService orchestrates the morning planning session.
type PlanningService struct {
	store     Store
	providers []TaskProvider
	presenter PlanPresenter
}

// NewPlanningService creates a PlanningService wired with the given store,
// task providers, and presenter.
func NewPlanningService(store Store, providers []TaskProvider, presenter PlanPresenter) *PlanningService {
	return &PlanningService{
		store:     store,
		providers: providers,
		presenter: presenter,
	}
}

// StartMorning runs the morning planning session for the given date.
// It returns the saved DayPlan on success. If a plan already exists for
// today it reports the fact and returns the existing plan without error.
func (s *PlanningService) StartMorning(date time.Time) (*DayPlan, error) {
	// 1. Check for an existing plan — bail out early if one already exists.
	existing, err := s.store.LoadDayPlan(date)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if err := s.presenter.ShowMessage("A plan for today already exists."); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// 2. Fetch tasks from all available providers.
	var allTasks []Task
	for _, p := range s.providers {
		if !p.IsAvailable() {
			continue
		}
		tasks, err := p.FetchTasks()
		if err != nil {
			s.presenter.ShowError(err)
			continue
		}
		allTasks = append(allTasks, tasks...)
	}

	// 3. Present the task list.
	if err := s.presenter.ShowTaskList(allTasks); err != nil {
		return nil, err
	}

	// 4. Prompt the user to select their commitments.
	committed, err := s.presenter.PromptCommitments(allTasks)
	if err != nil {
		return nil, err
	}

	// 5. Build and persist the day plan.
	plan := DayPlan{
		Date:      date,
		Tasks:     committed,
		Goals:     []string{},
		CreatedAt: time.Now(),
	}
	if err := s.store.SaveDayPlan(&plan); err != nil {
		return nil, err
	}

	// 6. Load the current streak; a missing streak is treated as zero.
	streak, err := s.store.LoadStreak()
	if err != nil {
		streak = &Streak{}
	}

	// 7. Confirm the plan to the user.
	if err := s.presenter.ShowPlanConfirmation(&plan, streak); err != nil {
		return nil, err
	}

	// 8. Return the saved plan.
	return &plan, nil
}
