package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/EwanGreer/day-planner/internal/core"
	"github.com/EwanGreer/day-planner/internal/integrations"
)

// compile-time assertion that Adapter satisfies the Integration interface.
var _ integrations.Integration = (*Adapter)(nil)

// HTTPClient is an interface so tests can inject a fake HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Adapter implements integrations.Integration for Jira.
type Adapter struct {
	cfg      core.JiraConfig
	client   HTTPClient
	lastCall time.Time
}

// New returns a new Jira Adapter backed by the real http.Client.
func New(cfg core.JiraConfig) *Adapter {
	return &Adapter{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewWithClient returns an Adapter that uses the provided HTTPClient.
// Intended for use in tests to inject a fake client.
func NewWithClient(cfg core.JiraConfig, client HTTPClient) *Adapter {
	return &Adapter{cfg: cfg, client: client}
}

// Name returns the integration name.
func (a *Adapter) Name() string {
	return "jira"
}

// IsAvailable reports whether all required Jira config fields are set.
func (a *Adapter) IsAvailable() bool {
	return a.cfg.Enabled &&
		a.cfg.BaseURL != "" &&
		a.cfg.Email != "" &&
		a.cfg.APIToken != ""
}

// FetchTasks retrieves open tasks assigned to the current user from Jira.
//
// Returns nil, nil when the adapter is not available or when the HTTP request
// fails with a non-2xx status (graceful fallback — does not abort the session).
func (a *Adapter) FetchTasks() ([]core.Task, error) {
	if !a.IsAvailable() {
		return nil, nil
	}

	// Simple rate-limit guard: sleep 1s if this adapter has been called before.
	if !a.lastCall.IsZero() {
		time.Sleep(time.Second)
	}
	a.lastCall = time.Now()

	jql := buildJQL(a.cfg.ProjectKeys)

	searchURL := strings.TrimRight(a.cfg.BaseURL, "/") +
		"/rest/api/3/search?jql=" + url.QueryEscape(jql) +
		"&maxResults=50&fields=summary,priority,project,status"

	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("jira: build request: %w", err)
	}
	req.SetBasicAuth(a.cfg.Email, a.cfg.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		log.Printf("jira: http request failed — skipping: %v", err)
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("jira: non-2xx response %d — skipping", resp.StatusCode)
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jira: read response body: %w", err)
	}

	var result searchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("jira: parse response: %w", err)
	}

	tasks := make([]core.Task, 0, len(result.Issues))
	for _, issue := range result.Issues {
		tasks = append(tasks, mapIssue(issue))
	}

	return tasks, nil
}

// buildJQL constructs the JQL query string for the search request.
func buildJQL(projectKeys []string) string {
	base := `assignee = currentUser() AND status NOT IN (Done, Closed, "Won't Do") ORDER BY priority DESC`
	if len(projectKeys) == 0 {
		return base
	}

	// Insert project filter before the ORDER BY clause.
	keys := make([]string, len(projectKeys))
	copy(keys, projectKeys)
	projectFilter := "project IN (" + strings.Join(keys, ", ") + ")"

	// Insert after the status clause, before ORDER BY.
	orderIdx := strings.Index(base, " ORDER BY")
	return base[:orderIdx] + " AND " + projectFilter + base[orderIdx:]
}

// mapIssue converts a Jira issue into a core.Task.
func mapIssue(issue jiraIssue) core.Task {
	return core.Task{
		ID:       issue.Key,
		Title:    issue.Fields.Summary,
		Status:   core.TaskStatusPending,
		Source:   core.TaskSourceJira,
		Priority: mapPriority(issue.Fields.Priority.Name),
		Project:  issue.Fields.Project.Key,
	}
}

// mapPriority converts a Jira priority name to an integer priority level.
func mapPriority(name string) int {
	switch name {
	case "Highest", "High":
		return 3
	case "Medium":
		return 2
	case "Low", "Lowest":
		return 1
	default:
		return 0
	}
}

// Jira REST API response types.

type searchResponse struct {
	Issues []jiraIssue `json:"issues"`
}

type jiraIssue struct {
	Key    string     `json:"key"`
	Fields jiraFields `json:"fields"`
}

type jiraFields struct {
	Summary  string       `json:"summary"`
	Priority jiraPriority `json:"priority"`
	Project  jiraProject  `json:"project"`
}

type jiraPriority struct {
	Name string `json:"name"`
}

type jiraProject struct {
	Key string `json:"key"`
}
