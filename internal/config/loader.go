package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/EwanGreer/day-planner/internal/core"
)

// rawConfig mirrors the TOML structure for parsing.
// Fields are pointers so we can distinguish zero-value from absent.
type rawConfig struct {
	General     rawGeneralConfig     `toml:"general"`
	Nudges      rawNudgesConfig      `toml:"nudges"`
	Taskwarrior rawTaskwarriorConfig `toml:"taskwarrior"`
	Jira        rawJiraConfig        `toml:"jira"`
}

type rawGeneralConfig struct {
	DataDir         *string `toml:"data_dir"`
	StreakThreshold *int    `toml:"streak_threshold"`
}

type rawNudgesConfig struct {
	IntervalMinutes *int              `toml:"interval_minutes"`
	FocusWindows    []rawFocusWindow  `toml:"focus_windows"`
}

type rawFocusWindow struct {
	Start string `toml:"start"`
	End   string `toml:"end"`
	Label string `toml:"label"`
}

type rawTaskwarriorConfig struct {
	Enabled *bool `toml:"enabled"`
}

type rawJiraConfig struct {
	Enabled     *bool    `toml:"enabled"`
	BaseURL     *string  `toml:"base_url"`
	Email       *string  `toml:"email"`
	APIToken    *string  `toml:"api_token"`
	ProjectKeys []string `toml:"project_keys"`
}

// DefaultConfig returns a Config populated with all default values.
// DataDir is returned unexpanded ("~/.day-planner") — callers must expand it
// before use via ExpandDataDir.
func DefaultConfig() *core.Config {
	return &core.Config{
		General: core.GeneralConfig{
			DataDir:         "~/.day-planner",
			StreakThreshold: 70,
		},
		Nudges: core.NudgesConfig{
			IntervalMinutes: 90,
		},
		Taskwarrior: core.TaskwarriorConfig{
			Enabled: true,
		},
	}
}

// ExpandDataDir replaces a leading "~" in path with the current user's home
// directory. The path is returned unchanged if it does not start with "~".
func ExpandDataDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, path[1:]), nil
}

// Load reads the TOML config file at path, merges it with defaults, applies
// environment variable overrides, validates the result, and returns the Config.
//
// If path is empty, the default path (~/.day-planner/config.toml) is used.
// If the file does not exist, Load returns DefaultConfig() without error.
func Load(path string) (*core.Config, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home dir: %w", err)
		}
		path = filepath.Join(home, ".day-planner", "config.toml")
	}

	cfg := DefaultConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		applyEnvOverrides(cfg)
		return cfg, nil
	}

	var raw rawConfig
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", path, err)
	}

	mergeRaw(cfg, &raw)
	applyEnvOverrides(cfg)

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// mergeRaw copies non-nil fields from raw into cfg, layering over defaults.
func mergeRaw(cfg *core.Config, raw *rawConfig) {
	if raw.General.DataDir != nil {
		cfg.General.DataDir = *raw.General.DataDir
	}
	if raw.General.StreakThreshold != nil {
		cfg.General.StreakThreshold = *raw.General.StreakThreshold
	}

	if raw.Nudges.IntervalMinutes != nil {
		cfg.Nudges.IntervalMinutes = *raw.Nudges.IntervalMinutes
	}
	if len(raw.Nudges.FocusWindows) > 0 {
		windows := make([]core.FocusWindow, len(raw.Nudges.FocusWindows))
		for i, w := range raw.Nudges.FocusWindows {
			windows[i] = core.FocusWindow{
				Start: w.Start,
				End:   w.End,
				Label: w.Label,
			}
		}
		cfg.Nudges.FocusWindows = windows
	}

	if raw.Taskwarrior.Enabled != nil {
		cfg.Taskwarrior.Enabled = *raw.Taskwarrior.Enabled
	}

	if raw.Jira.Enabled != nil {
		cfg.Jira.Enabled = *raw.Jira.Enabled
	}
	if raw.Jira.BaseURL != nil {
		cfg.Jira.BaseURL = *raw.Jira.BaseURL
	}
	if raw.Jira.Email != nil {
		cfg.Jira.Email = *raw.Jira.Email
	}
	if raw.Jira.APIToken != nil {
		cfg.Jira.APIToken = *raw.Jira.APIToken
	}
	if len(raw.Jira.ProjectKeys) > 0 {
		cfg.Jira.ProjectKeys = raw.Jira.ProjectKeys
	}
}

// applyEnvOverrides applies environment variable overrides to cfg.
func applyEnvOverrides(cfg *core.Config) {
	if token, ok := os.LookupEnv("JIRA_API_TOKEN"); ok {
		cfg.Jira.APIToken = token
	}
}

// validate checks cfg for constraint violations and returns the first error.
func validate(cfg *core.Config) error {
	if cfg.General.StreakThreshold < 0 || cfg.General.StreakThreshold > 100 {
		return fmt.Errorf(
			"config: general.streak_threshold must be between 0 and 100, got %d",
			cfg.General.StreakThreshold,
		)
	}
	if cfg.Nudges.IntervalMinutes < 1 {
		return fmt.Errorf(
			"config: nudges.interval_minutes must be at least 1, got %d",
			cfg.Nudges.IntervalMinutes,
		)
	}
	if cfg.Jira.Enabled {
		if cfg.Jira.BaseURL == "" {
			return fmt.Errorf("config: jira.base_url is required when jira.enabled is true")
		}
		if cfg.Jira.Email == "" {
			return fmt.Errorf("config: jira.email is required when jira.enabled is true")
		}
	}
	return nil
}
