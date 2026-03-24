package core

// Config holds all runtime configuration for the day planner.
// It has no dependency on any file format.
type Config struct {
	General     GeneralConfig
	Nudges      NudgesConfig
	Taskwarrior TaskwarriorConfig
	Jira        JiraConfig
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	DataDir         string // default: "~/.day-planner"
	StreakThreshold int    // default: 70 (minimum % score for streak)
}

// NudgesConfig holds settings for intraday nudge behaviour.
type NudgesConfig struct {
	IntervalMinutes int
	FocusWindows    []FocusWindow
}

// FocusWindow defines a named time window for deep-focus scheduling.
type FocusWindow struct {
	Start string // "HH:MM" 24-hour
	End   string // "HH:MM" 24-hour
	Label string
}

// TaskwarriorConfig holds settings for the Taskwarrior integration.
type TaskwarriorConfig struct {
	Enabled bool // default: true
}

// JiraConfig holds settings for the optional Jira integration.
type JiraConfig struct {
	Enabled     bool
	BaseURL     string
	Email       string
	APIToken    string   // overridable via JIRA_API_TOKEN env var
	ProjectKeys []string
}
