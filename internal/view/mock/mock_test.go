package mock_test

import (
	"errors"
	"testing"
	"time"

	"github.com/EwanGreer/day-planner/internal/core"
	"github.com/EwanGreer/day-planner/internal/view/mock"
)

func TestMockPresenter_ShowTaskList(t *testing.T) {
	m := &mock.Presenter{}
	tasks := []core.Task{
		{ID: "1", Title: "Write tests", Priority: 1, Status: core.TaskStatusPending, Source: core.TaskSourceManual},
		{ID: "2", Title: "Review PR", Priority: 2, Status: core.TaskStatusPending, Source: core.TaskSourceManual},
	}

	if err := m.ShowTaskList(tasks); err != nil {
		t.Fatalf("ShowTaskList returned unexpected error: %v", err)
	}

	if len(m.TasksShown) != len(tasks) {
		t.Fatalf("expected %d tasks recorded, got %d", len(tasks), len(m.TasksShown))
	}
	for i, task := range tasks {
		if m.TasksShown[i].ID != task.ID {
			t.Errorf("TasksShown[%d].ID = %q, want %q", i, m.TasksShown[i].ID, task.ID)
		}
	}
}

func TestMockPresenter_PromptCommitments_Default(t *testing.T) {
	m := &mock.Presenter{}
	tasks := []core.Task{
		{ID: "1", Title: "Task A", Priority: 1, Status: core.TaskStatusPending, Source: core.TaskSourceManual},
		{ID: "2", Title: "Task B", Priority: 2, Status: core.TaskStatusPending, Source: core.TaskSourceManual},
	}

	result, err := m.PromptCommitments(tasks)
	if err != nil {
		t.Fatalf("PromptCommitments returned unexpected error: %v", err)
	}

	if len(result) != len(tasks) {
		t.Fatalf("expected %d tasks returned, got %d", len(tasks), len(result))
	}
	if len(m.CommitmentsPrompted) != len(tasks) {
		t.Fatalf("expected %d tasks recorded in CommitmentsPrompted, got %d", len(tasks), len(m.CommitmentsPrompted))
	}
}

func TestMockPresenter_PromptCommitments_Configured(t *testing.T) {
	configured := []core.Task{
		{ID: "99", Title: "Configured task", Priority: 1, Status: core.TaskStatusPending, Source: core.TaskSourceManual},
	}
	m := &mock.Presenter{
		PromptCommitmentsResult: configured,
	}
	tasks := []core.Task{
		{ID: "1", Title: "Task A", Priority: 1, Status: core.TaskStatusPending, Source: core.TaskSourceManual},
	}

	result, err := m.PromptCommitments(tasks)
	if err != nil {
		t.Fatalf("PromptCommitments returned unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].ID != "99" {
		t.Errorf("expected configured result with ID 99, got %+v", result)
	}
	// Input tasks are still recorded
	if len(m.CommitmentsPrompted) != len(tasks) {
		t.Errorf("expected %d tasks recorded in CommitmentsPrompted, got %d", len(tasks), len(m.CommitmentsPrompted))
	}
}

func TestMockPresenter_ShowError(t *testing.T) {
	m := &mock.Presenter{}
	err1 := errors.New("something went wrong")
	err2 := errors.New("another error")

	m.ShowError(err1)
	m.ShowError(err2)

	if len(m.ErrorsShown) != 2 {
		t.Fatalf("expected 2 errors recorded, got %d", len(m.ErrorsShown))
	}
	if m.ErrorsShown[0] != err1 {
		t.Errorf("ErrorsShown[0] = %v, want %v", m.ErrorsShown[0], err1)
	}
	if m.ErrorsShown[1] != err2 {
		t.Errorf("ErrorsShown[1] = %v, want %v", m.ErrorsShown[1], err2)
	}
}

func TestMockPresenter_Close(t *testing.T) {
	m := &mock.Presenter{}

	if m.Closed {
		t.Fatal("Closed should be false before calling Close")
	}

	if err := m.Close(); err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}

	if !m.Closed {
		t.Error("expected Closed to be true after calling Close")
	}
}

func TestMockPresenter_ShowPlanConfirmation(t *testing.T) {
	m := &mock.Presenter{}
	plan := &core.DayPlan{
		Date:  time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Tasks: []core.Task{{ID: "1", Title: "Task A", Status: core.TaskStatusPending}},
		Score: 0,
	}
	streak := &core.Streak{Current: 3, Longest: 7}

	if err := m.ShowPlanConfirmation(plan, streak); err != nil {
		t.Fatalf("ShowPlanConfirmation returned unexpected error: %v", err)
	}
	if m.PlanConfirmed != plan {
		t.Errorf("expected PlanConfirmed to be the supplied plan pointer")
	}
}

func TestMockPresenter_ShowNudge(t *testing.T) {
	m := &mock.Presenter{}
	commitments := []core.Task{
		{ID: "1", Title: "Focus task", Status: core.TaskStatusInProgress},
	}

	if err := m.ShowNudge(commitments); err != nil {
		t.Fatalf("ShowNudge returned unexpected error: %v", err)
	}
	if len(m.NudgesSent) != 1 {
		t.Fatalf("expected 1 nudge recorded, got %d", len(m.NudgesSent))
	}
	if len(m.NudgesSent[0]) != 1 || m.NudgesSent[0][0].ID != "1" {
		t.Errorf("unexpected nudge contents: %+v", m.NudgesSent[0])
	}
}

func TestMockPresenter_ShowEODSummary(t *testing.T) {
	m := &mock.Presenter{}
	plan := &core.DayPlan{
		Date:  time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Tasks: []core.Task{{ID: "1", Title: "Done task", Status: core.TaskStatusDone}},
		Score: 85,
	}

	if err := m.ShowEODSummary(plan, 85); err != nil {
		t.Fatalf("ShowEODSummary returned unexpected error: %v", err)
	}
	if len(m.EODSummariesShown) != 1 {
		t.Fatalf("expected 1 EOD summary recorded, got %d", len(m.EODSummariesShown))
	}
	if m.EODSummariesShown[0] != plan {
		t.Errorf("expected EODSummariesShown[0] to be the supplied plan pointer")
	}
}

func TestMockPresenter_ShowMessage(t *testing.T) {
	m := &mock.Presenter{}

	if err := m.ShowMessage("hello"); err != nil {
		t.Fatalf("ShowMessage returned unexpected error: %v", err)
	}
	if err := m.ShowMessage("world"); err != nil {
		t.Fatalf("ShowMessage returned unexpected error: %v", err)
	}

	if len(m.MessagesShown) != 2 {
		t.Fatalf("expected 2 messages recorded, got %d", len(m.MessagesShown))
	}
	if m.MessagesShown[0] != "hello" || m.MessagesShown[1] != "world" {
		t.Errorf("unexpected messages: %v", m.MessagesShown)
	}
}
