package taskwarrior

import (
	"errors"
	"testing"
)

// fakeCommander is a test double for Commander.
type fakeCommander struct {
	out []byte
	err error
}

func (f fakeCommander) Run(_ string, _ ...string) ([]byte, error) {
	return f.out, f.err
}

// fakeNotFound simulates the error returned when the binary is missing.
// exec.LookPath wraps os.ErrNotExist so we match on string instead.
var errNotFound = errors.New("exec: \"task\": executable file not found in $PATH")

// threeTaskJSON is a pre-canned export with three tasks at different urgencies.
const threeTaskJSON = `[
  {
    "uuid": "aaa-111",
    "description": "Low urgency task",
    "status": "pending",
    "urgency": 1.5,
    "priority": "L",
    "project": "home",
    "tags": ["chores"],
    "due": ""
  },
  {
    "uuid": "bbb-222",
    "description": "High urgency task",
    "status": "pending",
    "urgency": 9.8,
    "priority": "H",
    "project": "work",
    "tags": ["urgent", "dev"],
    "due": "20260325T120000Z"
  },
  {
    "uuid": "ccc-333",
    "description": "Medium urgency task",
    "status": "pending",
    "urgency": 4.2,
    "priority": "M",
    "project": "work",
    "tags": [],
    "due": ""
  }
]`

func TestFetchTasks_Success(t *testing.T) {
	a := NewWithCommander(fakeCommander{out: []byte(threeTaskJSON)})

	tasks, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	// Tasks must be sorted by urgency descending.
	if tasks[0].Urgency != 9.8 {
		t.Errorf("expected first task urgency 9.8, got %f", tasks[0].Urgency)
	}
	if tasks[1].Urgency != 4.2 {
		t.Errorf("expected second task urgency 4.2, got %f", tasks[1].Urgency)
	}
	if tasks[2].Urgency != 1.5 {
		t.Errorf("expected third task urgency 1.5, got %f", tasks[2].Urgency)
	}

	// Field mapping for the highest-urgency task.
	top := tasks[0]
	if top.ID != "bbb-222" {
		t.Errorf("expected ID bbb-222, got %s", top.ID)
	}
	if top.Title != "High urgency task" {
		t.Errorf("expected title 'High urgency task', got %s", top.Title)
	}
	if top.Priority != 3 {
		t.Errorf("expected priority 3 (H), got %d", top.Priority)
	}
	if top.Project != "work" {
		t.Errorf("expected project 'work', got %s", top.Project)
	}
	if len(top.Tags) != 2 || top.Tags[0] != "urgent" || top.Tags[1] != "dev" {
		t.Errorf("unexpected tags: %v", top.Tags)
	}
	if top.Due == nil {
		t.Error("expected non-nil Due for task with due date")
	}
}

func TestFetchTasks_Empty(t *testing.T) {
	a := NewWithCommander(fakeCommander{out: []byte("[]")})

	tasks, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestFetchTasks_InvalidJSON(t *testing.T) {
	a := NewWithCommander(fakeCommander{out: []byte("not valid json {{{")})

	_, err := a.FetchTasks()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestFetchTasks_NotInstalled(t *testing.T) {
	a := NewWithCommander(fakeCommander{err: errNotFound})

	tasks, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("expected nil error when binary not found, got: %v", err)
	}
	if tasks != nil {
		t.Fatalf("expected nil tasks when binary not found, got %v", tasks)
	}
}

func TestPriorityMapping(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"H", 3},
		{"M", 2},
		{"L", 1},
		{"", 0},
	}

	for _, tc := range cases {
		got := mapPriority(tc.input)
		if got != tc.want {
			t.Errorf("mapPriority(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestIsAvailable_False(t *testing.T) {
	a := New()
	// We only assert no panic; `task` may or may not be installed in CI.
	// The boolean result is environment-dependent and cannot be asserted.
	_ = a.IsAvailable()
}
