package main

import (
	"fmt"
	"os"

	"github.com/EwanGreer/day-planner/internal/integrations/taskwarrior"
	"github.com/EwanGreer/day-planner/internal/view/tui"
)

func main() {
	presenter := tui.New()
	defer func() {
		if err := presenter.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "presenter close error: %v\n", err)
		}
	}()

	tw := taskwarrior.New()

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
}
