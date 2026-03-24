# Feature Report: Jira Integration

**Date**: 2026-03-24
**Language/Framework**: Go
**Issue**: #12
**Branch**: `feature/issue-12-jira`
**Status**: Complete

## Summary

Implemented a Jira REST API adapter that satisfies the existing
`integrations.Integration` interface. The adapter fetches open issues assigned
to the current user, maps them to `core.Task`, and integrates gracefully: it
returns `nil, nil` when disabled or when the HTTP request fails with a non-2xx
status. All required config fields (`Enabled`, `BaseURL`, `Email`, `APIToken`)
are checked before any network call is made. A compile-time assertion confirms
interface compliance. The adapter is wired into `main.go` behind the
`cfg.Jira.Enabled` flag.

## Changes Made

### internal/integrations/jira

- `internal/integrations/jira/adapter.go` — New package. Defines:
  - `HTTPClient` interface for test injection
  - `Adapter` struct with `cfg`, `client`, and `lastCall` fields
  - `New(cfg)` constructor using a real `http.Client` with a 10-second timeout
  - `NewWithClient(cfg, client)` constructor for testing
  - `Name()` returning `"jira"`
  - `IsAvailable()` checking all four required config fields
  - `FetchTasks()` building JQL, calling the Jira search API, and mapping results
  - `buildJQL(projectKeys)` constructing JQL with optional `project IN (...)` filter
  - `mapIssue(issue)` mapping a Jira issue to `core.Task`
  - `mapPriority(name)` mapping Jira priority names to int levels
  - Private response types: `searchResponse`, `jiraIssue`, `jiraFields`,
    `jiraPriority`, `jiraProject`
  - Compile-time assertion: `var _ integrations.Integration = (*Adapter)(nil)`

### cmd/day-planner

- `cmd/day-planner/main.go` — Added `jira` import and a block that constructs
  `jira.New(cfg.Jira)` when `cfg.Jira.Enabled` is true. The adapter variable is
  held ready to be passed to the planning service in a subsequent issue.

## Tests Written

- `internal/integrations/jira/adapter_test.go` — 9 unit tests + 1 skipped
  integration test using a `fakeHTTPClient` and `capFakeClient` test doubles.

  | Test | What is verified |
  |------|-----------------|
  | `TestFetchTasks_Success` | 3 issues returned, all fields mapped correctly (ID, Title, Status, Source, Priority, Project) |
  | `TestFetchTasks_Disabled` | `cfg.Enabled=false` returns `nil, nil` without an HTTP call |
  | `TestFetchTasks_EmptyProjectKeys` | JQL contains no `project IN` filter when `ProjectKeys` is nil |
  | `TestFetchTasks_ProjectKeysInJQL` | JQL contains `project IN (ENG, OPS)` when keys are provided |
  | `TestFetchTasks_HTTPError` | HTTP 500 response returns `nil, nil` (graceful fallback) |
  | `TestFetchTasks_InvalidJSON` | Garbage response body returns a non-nil error |
  | `TestPriorityMapping` | Highest/High=3, Medium=2, Low/Lowest=1, unknown/empty=0 |
  | `TestIsAvailable_AllFieldsSet` | Returns `true` when all four fields populated |
  | `TestIsAvailable_MissingToken` | Returns `false` when `APIToken` is empty |
  | `TestIntegration_FetchTasks` | Skipped unless `JIRA_BASE_URL` env var is set |

## Commits

| Hash | Message |
|------|---------|
| `f7018fb` | add Jira integration adapter |
| `6dba10c` | add unit tests for Jira adapter |
| `bbe3774` | wire Jira adapter into main |

## Acceptance Criteria

- [x] `go build ./...` passes
- [x] `go test ./...` passes — all 9 unit tests green
- [x] `IsAvailable` returns false when any required config field is missing
- [x] Non-2xx HTTP response returns `nil, nil` (graceful fallback)
- [x] `Adapter` implements `integrations.Integration` (compile-time assertion)
- [x] Integration test skipped unless `JIRA_BASE_URL` env var is set
- [x] No Jira imports leak into `internal/core`

## Design Decisions

- **Graceful fallback on HTTP errors**: A non-2xx status logs a warning and
  returns `nil, nil` rather than propagating an error. This matches the design
  goal that optional integrations must never abort the morning planning session.

- **Rate-limit guard via `lastCall`**: The adapter tracks the time of its last
  call and sleeps one second if called again within the same session. In practice
  the adapter is called once per session, but the guard prevents accidental
  thundering-herd behaviour if call patterns change.

- **`capFakeClient` for request inspection**: Two tests need to assert on the
  outgoing request URL (specifically the `jql` query parameter). A thin
  `capFakeClient` captures the `*http.Request` pointer while still returning a
  canned response, avoiding the overhead of `httptest.NewServer`.

- **JQL insertion point**: The project filter is inserted before the `ORDER BY`
  clause by locating the substring `" ORDER BY"` in the base JQL string. This is
  deterministic and avoids fragile string concatenation.

- **`TaskMain` is minimal**: The `TestMain` function in the test file exists to
  document the integration-test skip convention in one place. Unit tests run
  unconditionally.

## Known Limitations / Follow-up Work

- The `lastCall` rate-limit guard resets when the process exits. If the adapter
  is ever called from a long-running daemon, a persistent rate-limit store would
  be needed.
- `FetchTasks` always requests `maxResults=50`. Pagination support (using
  `startAt` + `total`) is deferred until a real usage pattern shows it is needed.
- The `status NOT IN (Done, Closed, "Won't Do")` filter is hardcoded. Exposing
  this as a config option is left for a follow-up issue.
