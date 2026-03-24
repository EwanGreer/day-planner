# Feature Report: Local Config File

**Date**: 2026-03-24
**Language/Framework**: Go
**Issue**: #3
**Branch**: `feature/issue-3-config-file`
**Status**: Complete

## Summary

Implemented a TOML-based configuration system for the day planner. A new
`core.Config` domain type carries all runtime configuration with no file-format
dependency. A new `internal/config` package provides `Load`, `DefaultConfig`,
and `ExpandDataDir`. The loader merges file values over defaults, applies the
`JIRA_API_TOKEN` env-var override, and validates constraints before returning.
Missing config files return defaults without error. `main.go` now loads config
before opening the store and derives the data directory from `cfg.General.DataDir`.

## Changes Made

### internal/core

- `internal/core/config.go` — New file. Defines `Config`, `GeneralConfig`,
  `NudgesConfig`, `FocusWindow`, `TaskwarriorConfig`, and `JiraConfig`. Zero
  imports — no file-format or IO dependencies.

### internal/config (new package)

- `internal/config/loader.go` — `Load(path string) (*core.Config, error)`,
  `DefaultConfig() *core.Config`, `ExpandDataDir(path string) (string, error)`.
  Uses `github.com/BurntSushi/toml` for parsing. Internal `rawConfig` mirrors
  the TOML shape with pointer fields so absent keys fall back to defaults rather
  than overwriting them with zero values. `mergeRaw` layers parsed values over
  defaults; `applyEnvOverrides` applies `JIRA_API_TOKEN`; `validate` enforces
  range and completeness constraints.
- `internal/config/default.toml` — Commented TOML template documenting all
  options with their defaults. Not loaded at runtime; serves as reference
  documentation.

### cmd/day-planner

- `cmd/day-planner/main.go` — Calls `config.Load("")` before opening the store,
  expands `DataDir` with `config.ExpandDataDir`, uses the expanded path as the
  base for `dbPath`, and prints `"Data dir: <path>"` via the presenter.

## Tests Written

- `internal/config/loader_test.go` — 8 test cases, each using an isolated
  `t.TempDir()`:

  | Test | What is verified |
  |------|-----------------|
  | `TestDefaultConfig` | Load with non-existent path returns all defaults |
  | `TestPartialConfig` | Only `[general]` present; other sections use defaults |
  | `TestFullConfig` | All fields parsed correctly including focus windows and project keys |
  | `TestEnvOverride` | `JIRA_API_TOKEN` env var overrides file value |
  | `TestInvalidStreakThreshold` | `streak_threshold = 150` returns a validation error |
  | `TestInvalidIntervalMinutes` | `interval_minutes = 0` returns a validation error |
  | `TestJiraEnabledWithoutURL` | `enabled = true` without `base_url` returns a validation error |
  | `TestMissingFile` | Non-existent path returns defaults, not an error |

## Commits

| Hash | Message |
|------|---------|
| `8cf1d86` | add Config domain types to internal/core |
| `dff5d58` | implement config loader with defaults, TOML parsing, env override, validation |
| `74113d1` | add unit tests for config loader |
| `b30ddc4` | wire config into main: load config, use DataDir for db path, show data dir |

## Acceptance Criteria

- [x] `go build ./...` passes
- [x] `go test ./...` passes — all 8 config tests green
- [x] `internal/core/config.go` has zero imports from `internal/config`
- [x] `internal/config` imports `internal/core` but nothing from `internal/view` or `internal/store`
- [x] `JIRA_API_TOKEN` env var overrides the file value
- [x] Missing config file returns defaults without error
- [x] Validation errors are clear and actionable

## Design Decisions

- **Pointer fields in `rawConfig`**: Using `*string` and `*int` for all raw
  fields lets `mergeRaw` distinguish "key absent from file" (nil) from "key
  present with zero value" (non-nil). This keeps default logic in one place
  (`DefaultConfig`) rather than scattering sentinel checks.
- **`ExpandDataDir` exported separately**: `main.go` needs to expand the path
  returned by `DefaultConfig`; keeping expansion as a standalone function avoids
  expanding inside `Load` (which would silently depend on `os.UserHomeDir` for
  every path, including already-expanded absolute paths).
- **`default.toml` as documentation only**: The file is committed for
  discoverability but is not embedded or read at runtime, keeping the binary
  free of hidden file-load paths.
- **`JIRA_API_TOKEN` override applied to missing-file path too**: The env
  override is applied in `applyEnvOverrides`, which is called for both the
  file-present and file-absent branches, so the override works even when no
  config file exists.

## Known Limitations / Follow-up Work

- `FocusWindow.Start` and `End` are stored as raw strings ("HH:MM"). No
  format validation is performed at load time; the nudge scheduler (issue #7)
  should parse and validate these when it sets up timers.
- `ExpandDataDir` only handles a leading `~` shorthand. Paths like `~other`
  (another user's home) are passed through unchanged, which is the correct
  conservative behaviour for a single-user tool.
- Config hot-reload is not implemented. The config is read once at startup;
  a restart is required to pick up changes.
