package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/EwanGreer/day-planner/internal/config"
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

	if wc.IsBlocked(time.Now()) {
		if err := presenter.ShowMessage(wc.StatusMessage(time.Now())); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	tw := taskwarrior.New()

	if err := presenter.ShowMessage(fmt.Sprintf("Data dir: %s", dataDir)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := presenter.ShowMessage(fmt.Sprintf("Store opened at %s", dbPath)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := presenter.ShowMessage("Day Planner starting..."); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if cfg.Taskwarrior.Enabled {
		if !tw.IsAvailable() {
			if err := presenter.ShowMessage("Taskwarrior not found in PATH -- skipping"); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else {
			tasks, err := tw.FetchTasks()
			if err != nil {
				fmt.Fprintf(os.Stderr, "taskwarrior error: %v\n", err)
				os.Exit(1)
			}
			if err := presenter.ShowTaskList(tasks); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}
	}

	_ = store // store will be passed to services in subsequent issues
}
