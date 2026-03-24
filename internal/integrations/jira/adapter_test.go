package jira

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/EwanGreer/day-planner/internal/core"
)

// TestMain skips integration tests when JIRA_BASE_URL is not set.
// Unit tests run unconditionally; only tests that explicitly call
// t.Skip on JIRA_BASE_URL would be affected by this guard — but we
// centralise the env check here for clarity.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// fakeHTTPClient implements HTTPClient for testing.
type fakeHTTPClient struct {
	statusCode int
	body       string
	err        error
}

func (f *fakeHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &http.Response{
		StatusCode: f.statusCode,
		Body:       io.NopCloser(strings.NewReader(f.body)),
	}, nil
}

// baseConfig returns a fully populated JiraConfig for testing.
func baseConfig() core.JiraConfig {
	return core.JiraConfig{
		Enabled:  true,
		BaseURL:  "https://example.atlassian.net",
		Email:    "user@example.com",
		APIToken: "token123",
	}
}

const threeIssueJSON = `{
  "issues": [
    {
      "key": "ENG-1",
      "fields": {
        "summary": "First task",
        "priority": {"name": "High"},
        "project": {"key": "ENG"}
      }
    },
    {
      "key": "ENG-2",
      "fields": {
        "summary": "Second task",
        "priority": {"name": "Medium"},
        "project": {"key": "ENG"}
      }
    },
    {
      "key": "ENG-3",
      "fields": {
        "summary": "Third task",
        "priority": {"name": "Low"},
        "project": {"key": "ENG"}
      }
    }
  ]
}`

func TestFetchTasks_Success(t *testing.T) {
	a := NewWithClient(baseConfig(), &fakeHTTPClient{
		statusCode: 200,
		body:       threeIssueJSON,
	})

	tasks, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	// Check field mapping for the first issue.
	first := tasks[0]
	if first.ID != "ENG-1" {
		t.Errorf("expected ID ENG-1, got %s", first.ID)
	}
	if first.Title != "First task" {
		t.Errorf("expected title 'First task', got %s", first.Title)
	}
	if first.Status != core.TaskStatusPending {
		t.Errorf("expected status pending, got %s", first.Status)
	}
	if first.Source != core.TaskSourceJira {
		t.Errorf("expected source jira, got %s", first.Source)
	}
	if first.Priority != 3 {
		t.Errorf("expected priority 3 (High), got %d", first.Priority)
	}
	if first.Project != "ENG" {
		t.Errorf("expected project ENG, got %s", first.Project)
	}

	// Second task — Medium priority.
	if tasks[1].Priority != 2 {
		t.Errorf("expected priority 2 (Medium), got %d", tasks[1].Priority)
	}

	// Third task — Low priority.
	if tasks[2].Priority != 1 {
		t.Errorf("expected priority 1 (Low), got %d", tasks[2].Priority)
	}
}

func TestFetchTasks_Disabled(t *testing.T) {
	cfg := baseConfig()
	cfg.Enabled = false
	a := NewWithClient(cfg, &fakeHTTPClient{statusCode: 200, body: threeIssueJSON})

	tasks, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks != nil {
		t.Errorf("expected nil tasks when disabled, got %v", tasks)
	}
}

func TestFetchTasks_EmptyProjectKeys(t *testing.T) {
	var capturedReq *http.Request

	// Use a capturing client to inspect the request URL.
	cap := &capFakeClient{
		statusCode: 200,
		body:       `{"issues":[]}`,
		captured:   &capturedReq,
	}

	cfg := baseConfig()
	cfg.ProjectKeys = nil
	a := NewWithClient(cfg, cap)

	_, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq == nil {
		t.Fatal("no request was made")
	}

	jql := capturedReq.URL.Query().Get("jql")
	if strings.Contains(jql, "project IN") {
		t.Errorf("expected no project filter in JQL, got: %s", jql)
	}
}

func TestFetchTasks_ProjectKeysInJQL(t *testing.T) {
	var capturedReq *http.Request

	cap := &capFakeClient{
		statusCode: 200,
		body:       `{"issues":[]}`,
		captured:   &capturedReq,
	}

	cfg := baseConfig()
	cfg.ProjectKeys = []string{"ENG", "OPS"}
	a := NewWithClient(cfg, cap)

	_, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq == nil {
		t.Fatal("no request was made")
	}

	jql := capturedReq.URL.Query().Get("jql")
	if !strings.Contains(jql, "project IN") {
		t.Errorf("expected project IN filter in JQL, got: %s", jql)
	}
	if !strings.Contains(jql, "ENG") || !strings.Contains(jql, "OPS") {
		t.Errorf("expected project keys ENG and OPS in JQL, got: %s", jql)
	}
}

func TestFetchTasks_HTTPError(t *testing.T) {
	a := NewWithClient(baseConfig(), &fakeHTTPClient{
		statusCode: 500,
		body:       `{"errorMessages":["Internal Server Error"]}`,
	})

	tasks, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("expected nil error on HTTP 500, got: %v", err)
	}
	if tasks != nil {
		t.Errorf("expected nil tasks on HTTP 500, got %v", tasks)
	}
}

func TestFetchTasks_NetworkError(t *testing.T) {
	a := NewWithClient(baseConfig(), &fakeHTTPClient{
		err: fmt.Errorf("dial tcp: connection refused"),
	})

	tasks, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("expected nil error on network failure (graceful fallback), got: %v", err)
	}
	if tasks != nil {
		t.Errorf("expected nil tasks on network failure, got %v", tasks)
	}
}

func TestFetchTasks_InvalidJSON(t *testing.T) {
	a := NewWithClient(baseConfig(), &fakeHTTPClient{
		statusCode: 200,
		body:       `not valid json {{{`,
	})

	_, err := a.FetchTasks()
	if err == nil {
		t.Fatal("expected error for invalid JSON response, got nil")
	}
}

func TestPriorityMapping(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"Highest", 3},
		{"High", 3},
		{"Medium", 2},
		{"Low", 1},
		{"Lowest", 1},
		{"", 0},
		{"Unknown", 0},
	}

	for _, tc := range cases {
		got := mapPriority(tc.input)
		if got != tc.want {
			t.Errorf("mapPriority(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestIsAvailable_AllFieldsSet(t *testing.T) {
	a := New(baseConfig())
	if !a.IsAvailable() {
		t.Error("expected IsAvailable() = true when all fields set")
	}
}

func TestIsAvailable_MissingToken(t *testing.T) {
	cfg := baseConfig()
	cfg.APIToken = ""
	a := New(cfg)
	if a.IsAvailable() {
		t.Error("expected IsAvailable() = false when APIToken is empty")
	}
}

// capFakeClient captures the outgoing request and returns a canned response.
type capFakeClient struct {
	statusCode int
	body       string
	captured   **http.Request
}

func (c *capFakeClient) Do(req *http.Request) (*http.Response, error) {
	*c.captured = req
	return &http.Response{
		StatusCode: c.statusCode,
		Body:       io.NopCloser(strings.NewReader(c.body)),
	}, nil
}

// TestIntegration_FetchTasks is an integration test that hits a real Jira instance.
// It is skipped unless JIRA_BASE_URL is set in the environment.
func TestIntegration_FetchTasks(t *testing.T) {
	baseURL := os.Getenv("JIRA_BASE_URL")
	if baseURL == "" {
		t.Skip("JIRA_BASE_URL not set — skipping integration test")
	}
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")

	cfg := core.JiraConfig{
		Enabled:  true,
		BaseURL:  baseURL,
		Email:    email,
		APIToken: apiToken,
	}

	a := New(cfg)
	tasks, err := a.FetchTasks()
	if err != nil {
		t.Fatalf("FetchTasks error: %v", err)
	}
	t.Logf("fetched %d tasks from Jira", len(tasks))
}
