package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/EwanGreer/day-planner/internal/config"
	"github.com/EwanGreer/day-planner/internal/core"
	"github.com/EwanGreer/day-planner/internal/integrations/jira"
	"github.com/EwanGreer/day-planner/internal/integrations/taskwarrior"
	"github.com/EwanGreer/day-planner/internal/nudge"
	"github.com/EwanGreer/day-planner/internal/store/sqlite"
	"github.com/EwanGreer/day-planner/internal/view/tui"
)

func main() {
	// Load configuration before opening any resources.
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	// Expand "~" in DataDir to the real home directory.
	dataDir, err := config.ExpandDataDir(cfg.General.DataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "expand data dir: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "create data dir: %v\n", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(dataDir, "planner.db")
	store, err := sqlite.New(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open store: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "store close error: %v\n", err)
		}
	}()

	wc, err := nudge.NewWindowChecker(cfg.Nudges.FocusWindows)
	if err != nil {
		fmt.Fprintf(os.Stderr, "focus windows: %v\n", err)
		os.Exit(1)
	}

	presenter := tui.New()
	defer func() {
		if err := presenter.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "presenter close error: %v\n", err)
		}
	}()

	if len(os.Args) < 2 {
		if err := presenter.ShowMessage("Usage: day-planner <start|status>"); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		runStart(store, presenter, cfg)
	case "status":
		runStatus(wc, presenter)
	default:
		if err := presenter.ShowMessage(fmt.Sprintf("Unknown command: %s", os.Args[1])); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
}

// runStart wires together the task providers and planning service, then
// executes the morning planning session for today.
func runStart(store core.Store, presenter core.PlanPresenter, cfg *core.Config) {
	var providers []core.TaskProvider

	if cfg.Taskwarrior.Enabled {
		providers = append(providers, taskwarrior.New())
	}

	if cfg.Jira.Enabled {
		providers = append(providers, jira.New(cfg.Jira))
	}

	svc := core.NewPlanningService(store, providers, presenter)
	if _, err := svc.StartMorning(time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "start morning: %v\n", err)
		os.Exit(1)
	}
}

// runStatus prints the current focus-window status message.
func runStatus(wc *nudge.WindowChecker, presenter core.PlanPresenter) {
	msg := wc.StatusMessage(time.Now())
	if err := presenter.ShowMessage(msg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
