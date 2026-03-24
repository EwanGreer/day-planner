# Feature Report: Taskwarrior Integration

**Date**: 2026-03-24
**Language/Framework**: Go
**Issue**: #4
**Branch**: `feature/issue-4-taskwarrior`
**Status**: Complete

## Summary

Replaced the Taskwarrior adapter stub with a fully functional implementation.
The adapter shells out to `task status:pending export`, parses the JSON response,
maps Taskwarrior fields to `core.Task`, and returns results sorted by urgency
descending. A `Commander` interface decouples the adapter from `os/exec` so tests
run without a real `task` binary. `core.Task` was extended with four new fields
(`Urgency`, `Project`, `Tags`, `Due`) required by the adapter. The main entry
point gates Taskwarrior behind the config flag and emits a clear message when the
binary is absent from PATH.

## Changes Made

### internal/core

- `internal/core/domain.go` — Added `Urgency float64`, `Project string`,
  `Tags []string`, and `Due *time.Time` fields to `Task`. Existing fields
  untouched.

### internal/integrations/taskwarrior

- `internal/integrations/taskwarrior/commander.go` — New file. Defines the
  `Commander` interface (`Run(name string, args ...string) ([]byte, error)`) and
  the `ExecCommander` concrete implementation backed by `os/exec`.

- `internal/integrations/taskwarrior/adapter.go` — Full rewrite of the stub.
  Defines private `twTask` for JSON unmarshalling. `New()` wires `ExecCommander`;
  `NewWithCommander(cmd)` accepts a test double. `IsAvailable()` uses
  `exec.LookPath`. `FetchTasks()` runs `task status:pending export`, handles the
  binary-not-found case by returning `nil, nil`, unmarshals JSON, maps fields
  (including priority string to int and due-date parsing in both RFC3339 and
  Taskwarrior compact format), sorts by urgency descending, and always returns a
  non-nil slice on success.

### cmd/day-planner

- `cmd/day-planner/main.go` — Taskwarrior block now checks
  `cfg.Taskwarrior.Enabled` before calling `IsAvailable()`. When enabled but not
  in PATH, prints "Taskwarrior not found in PATH -- skipping" via the presenter
  instead of silently doing nothing.

## Tests Written

- `internal/integrations/taskwarrior/adapter_test.go` — 6 test cases using a
  `fakeCommander` test double:

  | Test | What is verified |
  |------|-----------------|
  | `TestFetchTasks_Success` | 3 tasks returned, sorted urgency desc, all fields mapped correctly |
  | `TestFetchTasks_Empty` | Empty JSON array yields non-nil empty slice, no error |
  | `TestFetchTasks_InvalidJSON` | Garbage output returns a non-nil error |
  | `TestFetchTasks_NotInstalled` | "executable file not found" error yields nil, nil |
  | `TestPriorityMapping` | H=3, M=2, L=1, ""=0 |
  | `TestIsAvailable_False` | Calls `IsAvailable()` on a real adapter; asserts no panic |

## Commits

| Hash | Message |
|------|---------|
| `743975f` | extend core.Task with Urgency, Project, Tags, Due fields |
| `c7b6cb4` | add Commander interface and ExecCommander for testable shell delegation |
| `457ad75` | implement real Taskwarrior adapter with JSON parsing and urgency sort |
| `0396fa0` | add unit tests for Taskwarrior adapter |
| `209a584` | wire Taskwarrior adapter into main with config-gating and PATH-missing message |

## Acceptance Criteria

- [x] `go build ./...` passes
- [x] `go test ./...` passes — all 6 adapter tests green
- [x] `IsAvailable` uses `exec.LookPath` (not hardcoded false)
- [x] `FetchTasks` returns tasks sorted by urgency descending
- [x] Binary-not-found returns `nil, nil` (not an error)
- [x] Commander interface enables testing without a real `task` binary

## Design Decisions

- **`ExitError` passthrough with non-empty stdout**: `task export` can exit
  non-zero (e.g., Taskwarrior versions that exit 1 when there are no tasks) but
  still write valid JSON to stdout. The adapter checks for an `*exec.ExitError`
  and, when stdout is non-empty, proceeds to JSON parse rather than returning the
  error. This avoids false failures on strict Taskwarrior versions.

- **Two due-date formats**: Taskwarrior exports dates as `"20260324T120000Z"`
  (compact UTC), but the JSON spec allows RFC3339 too. Trying RFC3339 first keeps
  the parser forward-compatible with tools that normalise the field.

- **`nil, nil` for missing binary**: Treating a missing binary as an empty
  integration (rather than an error) lets the planner start cleanly on machines
  that don't have Taskwarrior. The "not found in PATH" presenter message provides
  user-visible feedback without a fatal exit.

- **Non-nil empty slice on success**: `make([]core.Task, 0, len(raw))` ensures
  callers can safely range over the result without a nil check when Taskwarrior
  is present but has no pending tasks.

## Known Limitations / Follow-up Work

- The adapter only queries `status:pending`. Taskwarrior supports richer filters
  (project, tag, due-before); exposing filter configuration can be done in a
  follow-up once the config schema is stable.
- Due-date parsing silently returns `nil` for unrecognised formats. A log warning
  would help diagnose future format changes.
- `TestIsAvailable_False` does not assert the return value because `task` may be
  installed in any given environment. A more deterministic approach would be to
  add a `lookPath` function field to `Adapter` and inject a fake in tests.
