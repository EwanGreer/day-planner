package core_test

import (
	"errors"
	"testing"
	"time"

	"github.com/EwanGreer/day-planner/internal/core"
)

// fakeStore is an in-memory implementation of core.Store for testing.
type fakeStore struct {
	plans       map[string]*core.DayPlan // keyed by date "2006-01-02"
	streak      *core.Streak
	streakErr   error
	savePlanErr error
}

func newFakeStore() *fakeStore {
	return &fakeStore{plans: make(map[string]*core.DayPlan)}
}

func (s *fakeStore) SaveDayPlan(plan *core.DayPlan) error {
	if s.savePlanErr != nil {
		return s.savePlanErr
	}
	key := plan.Date.Format("2006-01-02")
	s.plans[key] = plan
	return nil
}

func (s *fakeStore) LoadDayPlan(date time.Time) (*core.DayPlan, error) {
	key := date.Format("2006-01-02")
	return s.plans[key], nil
}

func (s *fakeStore) SaveCompletion(record core.CompletionRecord) error   { return nil }
func (s *fakeStore) LoadCompletions(_ time.Time) ([]core.CompletionRecord, error) {
	return nil, nil
}
func (s *fakeStore) SaveStreak(_ core.Streak) error { return nil }
func (s *fakeStore) LoadStreak() (*core.Streak, error) {
	return s.streak, s.streakErr
}
func (s *fakeStore) SaveReflection(_ core.Reflection) error { return nil }
func (s *fakeStore) LoadReflections(_ time.Time) ([]core.Reflection, error) {
	return nil, nil
}
func (s *fakeStore) SaveNudgeWindow(_ core.NudgeWindow) error { return nil }
func (s *fakeStore) LoadNudgeWindows() ([]core.NudgeWindow, error) {
	return nil, nil
}
func (s *fakeStore) Close() error { return nil }

// fakeProvider is a configurable core.TaskProvider.
type fakeProvider struct {
	name      string
	available bool
	tasks     []core.Task
	err       error
}

func (p *fakeProvider) Name() string               { return p.name }
func (p *fakeProvider) IsAvailable() bool          { return p.available }
func (p *fakeProvider) FetchTasks() ([]core.Task, error) { return p.tasks, p.err }

// fakePresenter records all presenter calls and returns configurable responses.
type fakePresenter struct {
	taskListShown    []core.Task
	commitmentTasks  []core.Task
	planConfirmed    *core.DayPlan
	streakShown      *core.Streak
	messages         []string
	errors           []error

	commitmentResult []core.Task
	commitmentErr    error
}

func (p *fakePresenter) ShowTaskList(tasks []core.Task) error {
	p.taskListShown = append(p.taskListShown, tasks...)
	return nil
}

func (p *fakePresenter) PromptCommitments(tasks []core.Task) ([]core.Task, error) {
	p.commitmentTasks = append(p.commitmentTasks, tasks...)
	if p.commitmentResult != nil {
		return p.commitmentResult, p.commitmentErr
	}
	return tasks, p.commitmentErr
}

func (p *fakePresenter) ShowPlanConfirmation(plan *core.DayPlan, streak *core.Streak) error {
	p.planConfirmed = plan
	p.streakShown = streak
	return nil
}

func (p *fakePresenter) ShowMessage(msg string) error {
	p.messages = append(p.messages, msg)
	return nil
}

func (p *fakePresenter) ShowError(err error) {
	p.errors = append(p.errors, err)
}

// TestStartMorning_HappyPath verifies the full happy-path: tasks fetched from
// a provider, user selects 2 commitments, plan saved, confirmation shown.
func TestStartMorning_HappyPath(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)

	tasks := []core.Task{
		{ID: "1", Title: "Task A", Source: core.TaskSourceTaskwarrior},
		{ID: "2", Title: "Task B", Source: core.TaskSourceTaskwarrior},
		{ID: "3", Title: "Task C", Source: core.TaskSourceTaskwarrior},
	}

	provider := &fakeProvider{name: "taskwarrior", available: true, tasks: tasks}
	store := newFakeStore()
	store.streak = &core.Streak{Current: 3, Longest: 5}

	presenter := &fakePresenter{
		commitmentResult: tasks[:2], // user picks first two
	}

	svc := core.NewPlanningService(store, []core.TaskProvider{provider}, presenter)
	plan, err := svc.StartMorning(today)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == nil {
		t.Fatal("expected non-nil plan")
	}
	if len(plan.Tasks) != 2 {
		t.Errorf("expected 2 committed tasks, got %d", len(plan.Tasks))
	}
	if presenter.planConfirmed == nil {
		t.Error("expected ShowPlanConfirmation to be called")
	}
	if presenter.streakShown == nil || presenter.streakShown.Current != 3 {
		t.Errorf("expected streak current=3, got %+v", presenter.streakShown)
	}
	// plan should be persisted in the store
	saved, _ := store.LoadDayPlan(today)
	if saved == nil {
		t.Error("expected plan to be saved in store")
	}
}

// TestStartMorning_ExistingPlan verifies that when a plan already exists for
// today, the service shows a message and returns the existing plan without
// saving a new one.
func TestStartMorning_ExistingPlan(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	existing := &core.DayPlan{
		Date:  today,
		Tasks: []core.Task{{ID: "x", Title: "Existing Task"}},
	}

	store := newFakeStore()
	store.plans[today.Format("2006-01-02")] = existing

	presenter := &fakePresenter{}
	provider := &fakeProvider{name: "tw", available: true, tasks: []core.Task{}}

	svc := core.NewPlanningService(store, []core.TaskProvider{provider}, presenter)
	plan, err := svc.StartMorning(today)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan != existing {
		t.Error("expected the existing plan to be returned")
	}
	if len(presenter.messages) == 0 {
		t.Error("expected an 'already exists' message to be shown")
	}
	// Confirm that no new plan was committed via PromptCommitments.
	if len(presenter.commitmentTasks) != 0 {
		t.Error("PromptCommitments should not have been called")
	}
}

// TestStartMorning_NoTasks verifies that when all providers return no tasks,
// the task list is still shown (empty), the user is still prompted, and a plan
// with zero committed tasks is saved.
func TestStartMorning_NoTasks(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)

	provider := &fakeProvider{name: "tw", available: true, tasks: nil}
	store := newFakeStore()
	presenter := &fakePresenter{
		commitmentResult: []core.Task{}, // user commits to nothing
	}

	svc := core.NewPlanningService(store, []core.TaskProvider{provider}, presenter)
	plan, err := svc.StartMorning(today)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == nil {
		t.Fatal("expected non-nil plan")
	}
	if len(plan.Tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(plan.Tasks))
	}
	if presenter.planConfirmed == nil {
		t.Error("expected ShowPlanConfirmation to be called")
	}
}

// TestStartMorning_SourceError verifies that when one provider errors, the
// error is shown via the presenter but the session continues with tasks from
// other providers.
func TestStartMorning_SourceError(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)

	goodTasks := []core.Task{
		{ID: "g1", Title: "Good Task", Source: core.TaskSourceTaskwarrior},
	}
	badProvider := &fakeProvider{
		name:      "broken",
		available: true,
		err:       errors.New("connection refused"),
	}
	goodProvider := &fakeProvider{
		name:      "taskwarrior",
		available: true,
		tasks:     goodTasks,
	}

	store := newFakeStore()
	presenter := &fakePresenter{}

	svc := core.NewPlanningService(store, []core.TaskProvider{badProvider, goodProvider}, presenter)
	plan, err := svc.StartMorning(today)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(presenter.errors) == 0 {
		t.Error("expected the provider error to be shown via ShowError")
	}
	if len(presenter.taskListShown) != 1 {
		t.Errorf("expected 1 task from the good provider, got %d", len(presenter.taskListShown))
	}
	if plan == nil {
		t.Fatal("expected non-nil plan")
	}
}

// TestStartMorning_SaveError verifies that a store.SaveDayPlan error is
// propagated back to the caller.
func TestStartMorning_SaveError(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)

	store := newFakeStore()
	store.savePlanErr = errors.New("disk full")

	presenter := &fakePresenter{}
	provider := &fakeProvider{name: "tw", available: true, tasks: []core.Task{
		{ID: "1", Title: "Task A"},
	}}

	svc := core.NewPlanningService(store, []core.TaskProvider{provider}, presenter)
	_, err := svc.StartMorning(today)

	if err == nil {
		t.Fatal("expected an error from SaveDayPlan, got nil")
	}
	if !errors.Is(err, store.savePlanErr) {
		t.Errorf("expected savePlanErr, got %v", err)
	}
}
