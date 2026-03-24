package sqlite_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/EwanGreer/day-planner/internal/core"
	"github.com/EwanGreer/day-planner/internal/store/sqlite"
)

// newTestStore opens a fresh SQLiteStore in a temp directory.
func newTestStore(t *testing.T) *sqlite.SQLiteStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := sqlite.New(path)
	if err != nil {
		t.Fatalf("sqlite.New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// mustDate parses a YYYY-MM-DD string or fails the test.
func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("mustDate(%q): %v", s, err)
	}
	return d
}

func TestSaveDayPlan_and_Load(t *testing.T) {
	store := newTestStore(t)

	now := time.Now().Truncate(time.Second)
	plan := &core.DayPlan{
		Date: mustDate(t, "2026-03-24"),
		Tasks: []core.Task{
			{ID: "t1", Title: "Write tests", Priority: 1, Status: core.TaskStatusPending, Source: core.TaskSourceManual},
			{ID: "t2", Title: "Ship feature", Priority: 2, Status: core.TaskStatusInProgress, Source: core.TaskSourceJira},
		},
		Goals:     []string{"finish issue #2", "review PRs"},
		Score:     7,
		CreatedAt: now,
	}

	if err := store.SaveDayPlan(plan); err != nil {
		t.Fatalf("SaveDayPlan: %v", err)
	}

	got, err := store.LoadDayPlan(mustDate(t, "2026-03-24"))
	if err != nil {
		t.Fatalf("LoadDayPlan: %v", err)
	}
	if got == nil {
		t.Fatal("LoadDayPlan returned nil, want plan")
	}

	if !got.Date.Equal(plan.Date) {
		t.Errorf("Date: got %v, want %v", got.Date, plan.Date)
	}
	if got.Score != plan.Score {
		t.Errorf("Score: got %d, want %d", got.Score, plan.Score)
	}
	if len(got.Tasks) != 2 {
		t.Fatalf("Tasks len: got %d, want 2", len(got.Tasks))
	}
	if got.Tasks[0].ID != "t1" || got.Tasks[1].ID != "t2" {
		t.Errorf("Task IDs: got %q %q, want t1 t2", got.Tasks[0].ID, got.Tasks[1].ID)
	}
	if len(got.Goals) != 2 {
		t.Fatalf("Goals len: got %d, want 2", len(got.Goals))
	}
	if got.Goals[0] != "finish issue #2" {
		t.Errorf("Goals[0]: got %q, want %q", got.Goals[0], "finish issue #2")
	}
	if !got.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, now)
	}
}

func TestLoadDayPlan_NotFound(t *testing.T) {
	store := newTestStore(t)

	got, err := store.LoadDayPlan(mustDate(t, "2099-01-01"))
	if err != nil {
		t.Fatalf("LoadDayPlan: unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("LoadDayPlan: got %+v, want nil", got)
	}
}

func TestSaveCompletion_and_Load(t *testing.T) {
	store := newTestStore(t)

	date := mustDate(t, "2026-03-24")
	now := time.Now().Truncate(time.Second)

	records := []core.CompletionRecord{
		{Date: date, TaskID: "t1", Status: core.CompletionStatusComplete, CompletedAt: now},
		{Date: date, TaskID: "t2", Status: core.CompletionStatusSkipped, CompletedAt: now},
	}
	for _, r := range records {
		if err := store.SaveCompletion(r); err != nil {
			t.Fatalf("SaveCompletion(%q): %v", r.TaskID, err)
		}
	}

	got, err := store.LoadCompletions(date)
	if err != nil {
		t.Fatalf("LoadCompletions: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len: got %d, want 2", len(got))
	}
	if got[0].TaskID != "t1" {
		t.Errorf("TaskID[0]: got %q, want t1", got[0].TaskID)
	}
	if got[0].Status != core.CompletionStatusComplete {
		t.Errorf("Status[0]: got %q, want complete", got[0].Status)
	}
	if got[1].TaskID != "t2" {
		t.Errorf("TaskID[1]: got %q, want t2", got[1].TaskID)
	}
	if got[1].Status != core.CompletionStatusSkipped {
		t.Errorf("Status[1]: got %q, want skipped", got[1].Status)
	}
	if !got[0].CompletedAt.Equal(now) {
		t.Errorf("CompletedAt[0]: got %v, want %v", got[0].CompletedAt, now)
	}
}

func TestSaveStreak_and_Load(t *testing.T) {
	store := newTestStore(t)

	lastActive := mustDate(t, "2026-03-23")
	streak := core.Streak{
		Current:        5,
		Longest:        12,
		LastActiveDate: lastActive,
	}

	if err := store.SaveStreak(streak); err != nil {
		t.Fatalf("SaveStreak: %v", err)
	}

	got, err := store.LoadStreak()
	if err != nil {
		t.Fatalf("LoadStreak: %v", err)
	}
	if got == nil {
		t.Fatal("LoadStreak returned nil")
	}
	if got.Current != 5 {
		t.Errorf("Current: got %d, want 5", got.Current)
	}
	if got.Longest != 12 {
		t.Errorf("Longest: got %d, want 12", got.Longest)
	}
	if !got.LastActiveDate.Equal(lastActive) {
		t.Errorf("LastActiveDate: got %v, want %v", got.LastActiveDate, lastActive)
	}
}

func TestLoadStreak_Empty(t *testing.T) {
	store := newTestStore(t)

	got, err := store.LoadStreak()
	if err != nil {
		t.Fatalf("LoadStreak: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("LoadStreak returned nil, want zero-value struct")
	}
	if got.Current != 0 {
		t.Errorf("Current: got %d, want 0", got.Current)
	}
	if got.Longest != 0 {
		t.Errorf("Longest: got %d, want 0", got.Longest)
	}
}

func TestSaveReflection_and_Load(t *testing.T) {
	store := newTestStore(t)

	date := mustDate(t, "2026-03-24")
	now := time.Now().Truncate(time.Second)
	refl := core.Reflection{
		Date:      date,
		Text:      "Good progress today, focused well.",
		CreatedAt: now,
	}

	if err := store.SaveReflection(refl); err != nil {
		t.Fatalf("SaveReflection: %v", err)
	}

	got, err := store.LoadReflections(date)
	if err != nil {
		t.Fatalf("LoadReflections: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len: got %d, want 1", len(got))
	}
	if got[0].Text != refl.Text {
		t.Errorf("Text: got %q, want %q", got[0].Text, refl.Text)
	}
	if !got[0].Date.Equal(date) {
		t.Errorf("Date: got %v, want %v", got[0].Date, date)
	}
	if !got[0].CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: got %v, want %v", got[0].CreatedAt, now)
	}
}

func TestSaveNudgeWindow_and_Load(t *testing.T) {
	store := newTestStore(t)

	now := time.Now().Truncate(time.Second)
	windows := []core.NudgeWindow{
		{StartTime: now, EndTime: now.Add(2 * time.Hour), Label: "focus"},
		{StartTime: now.Add(4 * time.Hour), EndTime: now.Add(5 * time.Hour), Label: "dnd"},
	}
	for _, w := range windows {
		if err := store.SaveNudgeWindow(w); err != nil {
			t.Fatalf("SaveNudgeWindow(%q): %v", w.Label, err)
		}
	}

	got, err := store.LoadNudgeWindows()
	if err != nil {
		t.Fatalf("LoadNudgeWindows: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len: got %d, want 2", len(got))
	}
	if got[0].Label != "focus" {
		t.Errorf("Label[0]: got %q, want focus", got[0].Label)
	}
	if got[1].Label != "dnd" {
		t.Errorf("Label[1]: got %q, want dnd", got[1].Label)
	}
	if !got[0].StartTime.Equal(windows[0].StartTime) {
		t.Errorf("StartTime[0]: got %v, want %v", got[0].StartTime, windows[0].StartTime)
	}
}

func TestMigrations_Idempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "idempotent.db")

	s1, err := sqlite.New(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	s1.Close()

	s2, err := sqlite.New(path)
	if err != nil {
		t.Fatalf("second open (migrations not idempotent): %v", err)
	}
	s2.Close()
}
