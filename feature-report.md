# Feature Report: Local SQLite Data Store

**Date**: 2026-03-24
**Language/Framework**: Go
**Issue**: #2
**Branch**: `feature/issue-2-local-data-store`
**Status**: Complete

## Summary

Implemented a SQLite-backed local data store using `modernc.org/sqlite`
(pure Go, no CGO). The store satisfies the expanded `core.Store` interface
covering five entity types: `DayPlan`, `CompletionRecord`, `Streak`,
`Reflection`, and `NudgeWindow`. Schema migrations are versioned and tracked
in a `schema_version` table. The store is wired into `main.go` and opens at
`~/.day-planner/planner.db` by default.

## Changes Made

### internal/core

- `internal/core/domain.go` — Added `CompletionRecord`, `CompletionStatus`,
  `Streak`, `Reflection`, `NudgeWindow` domain types, and `CreatedAt` field
  on `DayPlan`.
- `internal/core/store.go` — Replaced the two-method stub with the full
  `Store` interface covering all five entity types plus a `Close()` lifecycle
  method.

### internal/store/sqlite (new package)

- `internal/store/sqlite/migrations.go` — Integer-versioned migration system.
  Migrations are applied in ascending order; already-applied versions are
  skipped. Schema version 1 creates six tables: `schema_version`,
  `day_plans`, `completion_records`, `streaks`, `reflections`,
  `nudge_windows`.
- `internal/store/sqlite/store.go` — `SQLiteStore` struct with a `New(path)`
  constructor and all `core.Store` methods. Dates stored as `YYYY-MM-DD`
  text; timestamps as RFC3339. `DayPlan.Tasks` and `Goals` stored as JSON
  blobs. Compile-time interface assertion at package level.

### cmd/day-planner

- `cmd/day-planner/main.go` — Resolves `~/.day-planner/planner.db` via
  `os.UserHomeDir()`, creates the directory with `os.MkdirAll`, opens the
  store, defers `Close()`, and logs the store path through the presenter.

## Tests Written

- `internal/store/sqlite/store_test.go` — 8 test cases, each using an
  isolated `t.TempDir()` database:

  | Test | What is verified |
  |------|-----------------|
  | `TestSaveDayPlan_and_Load` | Round-trip: tasks, goals, score, created_at |
  | `TestLoadDayPlan_NotFound` | Missing date returns `nil, nil` (not an error) |
  | `TestSaveCompletion_and_Load` | Two records saved; count and fields correct |
  | `TestSaveStreak_and_Load` | Current, Longest, LastActiveDate round-trip |
  | `TestLoadStreak_Empty` | No data returns zero-value `&Streak{}`, not nil |
  | `TestSaveReflection_and_Load` | Text, date, created_at round-trip |
  | `TestSaveNudgeWindow_and_Load` | Two windows saved; count and labels correct |
  | `TestMigrations_Idempotent` | Opening the same DB twice does not error |

## Commits

| Hash | Message |
|------|---------|
| `dfed3e6` | expand domain types and Store interface in internal/core |
| `5b0fd5f` | add SQLite store implementation with versioned migrations |
| `08a3d76` | add unit tests for SQLiteStore (8 cases) |
| `c70f3b9` | wire SQLite store into main startup |

## Acceptance Criteria

- [x] `go build ./...` passes
- [x] `go test ./...` passes — all 8 store tests green
- [x] `internal/core` has zero new imports from `internal/store` or any storage engine
- [x] `SQLiteStore` has compile-time assertion it implements `core.Store`
- [x] Schema migrations are versioned (version tracked in `schema_version` table)
- [x] Store location defaults to `~/.day-planner/planner.db` but is injectable via `New(path)`
- [x] `LoadDayPlan` returns `nil, nil` for missing dates (not an error)
- [x] `LoadStreak` returns zero-value struct for missing data (not nil, not an error)

## Design Decisions

- **`modernc.org/sqlite` over `mattn/go-sqlite3`**: Pure Go driver removes the
  CGO requirement, keeping the binary portable across machines without a C
  toolchain.
- **`dateFmt = "2006-01-02"` for dates, RFC3339 for timestamps**: Dates that
  represent calendar days (plan date, completion date, streak date) are stored
  without a time component to avoid timezone ambiguity during day-boundary
  comparisons. Full timestamps (created_at, completed_at, nudge windows)
  use RFC3339.
- **`INSERT OR REPLACE` for `SaveDayPlan`**: Natural upsert semantics — the
  caller always provides a complete plan for a given day; there is no partial
  update path at this layer.
- **Singleton streak row (`id = 1`)**: There is only one streak per user.
  Using a fixed primary key makes the upsert trivial and avoids an extra
  `SELECT` before every save.
- **`_ = store` in main**: The store variable is retained but not yet passed
  anywhere. Subsequent issues (planning ritual, check-in commands, etc.) will
  inject it into the service layer.

## Known Limitations / Follow-up Work

- No transaction support yet. Bulk operations (e.g., saving a plan and its
  completions atomically) will require wrapping in a `sql.Tx`; a future issue
  can expose a `WithTx` helper if needed.
- `LoadCompletions`, `LoadReflections`, and `LoadNudgeWindows` return `nil`
  slices (not empty slices) when there are no rows. Callers should treat both
  as empty; this can be normalised to `[]T{}` in a later clean-up if callers
  find `nil` inconvenient.
- The database path is hardcoded in `main.go`; issue #3 (config file) will
  make this configurable.
