package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EwanGreer/day-planner/internal/integrations/taskwarrior"
	"github.com/EwanGreer/day-planner/internal/store/sqlite"
	"github.com/EwanGreer/day-planner/internal/view/tui"
)

func main() {
	// Resolve the default database path: ~/.day-planner/planner.db
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve home dir: %v\n", err)
		os.Exit(1)
	}

	dbDir := filepath.Join(homeDir, ".day-planner")
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "create data dir: %v\n", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(dbDir, "planner.db")
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

	presenter := tui.New()
	defer func() {
		if err := presenter.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "presenter close error: %v\n", err)
		}
	}()

	tw := taskwarrior.New()

	if err := presenter.ShowMessage(fmt.Sprintf("Store opened at %s", dbPath)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := presenter.ShowMessage("Day Planner starting..."); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if tw.IsAvailable() {
		tasks, err := tw.FetchTasks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "taskwarrior error: %v\n", err)
			os.Exit(1)
		}
		if err := presenter.ShowTasks(tasks); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	_ = store // store will be passed to services in subsequent issues
}
