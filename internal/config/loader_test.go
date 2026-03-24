package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EwanGreer/day-planner/internal/config"
)

// writeTOML writes content to a temp file and returns the path.
func writeTOML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

// TestDefaultConfig verifies that Load("") with no file returns defaults.
func TestDefaultConfig(t *testing.T) {
	// Point at a non-existent directory so the default path resolution
	// falls through to the missing-file branch.
	path := filepath.Join(t.TempDir(), "nonexistent", "config.toml")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.General.DataDir != "~/.day-planner" {
		t.Errorf("DataDir: got %q, want %q", cfg.General.DataDir, "~/.day-planner")
	}
	if cfg.General.StreakThreshold != 70 {
		t.Errorf("StreakThreshold: got %d, want 70", cfg.General.StreakThreshold)
	}
	if cfg.Nudges.IntervalMinutes != 90 {
		t.Errorf("IntervalMinutes: got %d, want 90", cfg.Nudges.IntervalMinutes)
	}
	if !cfg.Taskwarrior.Enabled {
		t.Error("Taskwarrior.Enabled: got false, want true")
	}
	if cfg.Jira.Enabled {
		t.Error("Jira.Enabled: got true, want false")
	}
}

// TestPartialConfig verifies that a TOML with only [general] uses defaults
// for all other sections.
func TestPartialConfig(t *testing.T) {
	path := writeTOML(t, `
[general]
data_dir = "/tmp/custom"
streak_threshold = 80
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.General.DataDir != "/tmp/custom" {
		t.Errorf("DataDir: got %q, want /tmp/custom", cfg.General.DataDir)
	}
	if cfg.General.StreakThreshold != 80 {
		t.Errorf("StreakThreshold: got %d, want 80", cfg.General.StreakThreshold)
	}
	// Sections not in the file must fall back to defaults.
	if cfg.Nudges.IntervalMinutes != 90 {
		t.Errorf("IntervalMinutes: got %d, want 90 (default)", cfg.Nudges.IntervalMinutes)
	}
	if !cfg.Taskwarrior.Enabled {
		t.Error("Taskwarrior.Enabled: got false, want true (default)")
	}
}

// TestFullConfig verifies that a TOML with all fields parses every field.
func TestFullConfig(t *testing.T) {
	path := writeTOML(t, `
[general]
data_dir = "/data/planner"
streak_threshold = 60

[nudges]
interval_minutes = 45
focus_windows = [
  { start = "09:00", end = "10:30", label = "deep work" },
  { start = "14:00", end = "15:00", label = "review" },
]

[taskwarrior]
enabled = false

[jira]
enabled = true
base_url = "https://example.atlassian.net"
email = "user@example.com"
api_token = "tok123"
project_keys = ["PROJ", "OPS"]
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.General.DataDir != "/data/planner" {
		t.Errorf("DataDir: got %q", cfg.General.DataDir)
	}
	if cfg.General.StreakThreshold != 60 {
		t.Errorf("StreakThreshold: got %d", cfg.General.StreakThreshold)
	}
	if cfg.Nudges.IntervalMinutes != 45 {
		t.Errorf("IntervalMinutes: got %d", cfg.Nudges.IntervalMinutes)
	}
	if len(cfg.Nudges.FocusWindows) != 2 {
		t.Fatalf("FocusWindows: got %d, want 2", len(cfg.Nudges.FocusWindows))
	}
	if cfg.Nudges.FocusWindows[0].Start != "09:00" {
		t.Errorf("FocusWindows[0].Start: got %q", cfg.Nudges.FocusWindows[0].Start)
	}
	if cfg.Nudges.FocusWindows[1].Label != "review" {
		t.Errorf("FocusWindows[1].Label: got %q", cfg.Nudges.FocusWindows[1].Label)
	}
	if cfg.Taskwarrior.Enabled {
		t.Error("Taskwarrior.Enabled: got true, want false")
	}
	if !cfg.Jira.Enabled {
		t.Error("Jira.Enabled: got false, want true")
	}
	if cfg.Jira.BaseURL != "https://example.atlassian.net" {
		t.Errorf("Jira.BaseURL: got %q", cfg.Jira.BaseURL)
	}
	if cfg.Jira.Email != "user@example.com" {
		t.Errorf("Jira.Email: got %q", cfg.Jira.Email)
	}
	if cfg.Jira.APIToken != "tok123" {
		t.Errorf("Jira.APIToken: got %q", cfg.Jira.APIToken)
	}
	if len(cfg.Jira.ProjectKeys) != 2 || cfg.Jira.ProjectKeys[0] != "PROJ" {
		t.Errorf("Jira.ProjectKeys: got %v", cfg.Jira.ProjectKeys)
	}
}

// TestEnvOverride verifies that JIRA_API_TOKEN env var overrides the file value.
func TestEnvOverride(t *testing.T) {
	path := writeTOML(t, `
[jira]
enabled = true
base_url = "https://example.atlassian.net"
email = "user@example.com"
api_token = "file-token"
`)
	t.Setenv("JIRA_API_TOKEN", "secret-env-token")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Jira.APIToken != "secret-env-token" {
		t.Errorf("APIToken: got %q, want secret-env-token", cfg.Jira.APIToken)
	}
}

// TestInvalidStreakThreshold verifies validation rejects out-of-range thresholds.
func TestInvalidStreakThreshold(t *testing.T) {
	path := writeTOML(t, `
[general]
streak_threshold = 150
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

// TestInvalidIntervalMinutes verifies validation rejects interval_minutes < 1.
func TestInvalidIntervalMinutes(t *testing.T) {
	path := writeTOML(t, `
[nudges]
interval_minutes = 0
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

// TestJiraEnabledWithoutURL verifies validation rejects Jira enabled without base_url.
func TestJiraEnabledWithoutURL(t *testing.T) {
	path := writeTOML(t, `
[jira]
enabled = true
email = "user@example.com"
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

// TestFocusWindowValidation_StartAfterEnd verifies that a focus window where
// start >= end is rejected by validation.
func TestFocusWindowValidation_StartAfterEnd(t *testing.T) {
	path := writeTOML(t, `
[nudges]
focus_windows = [
  { start = "14:00", end = "09:00", label = "bad window" },
]
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for start >= end, got nil")
	}
}

// TestFocusWindowValidation_InvalidFormat verifies that a focus window with a
// non-HH:MM start time is rejected.
func TestFocusWindowValidation_InvalidFormat(t *testing.T) {
	path := writeTOML(t, `
[nudges]
focus_windows = [
  { start = "9am", end = "10:00", label = "bad format" },
]
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for invalid HH:MM format, got nil")
	}
}

// TestMissingFile verifies that a non-existent path returns defaults without error.
func TestMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.toml")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("cfg is nil")
	}
	if cfg.General.StreakThreshold != 70 {
		t.Errorf("StreakThreshold: got %d, want 70", cfg.General.StreakThreshold)
	}
}
