package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/EwanGreer/day-planner/internal/core"
	"github.com/EwanGreer/day-planner/internal/view"
)

var _ view.Presenter = (*TUI)(nil)

// TUI implements view.Presenter by printing to stdout.
// This is a stub; a full terminal UI will be added in a later issue.
type TUI struct{}

// New returns a new TUI presenter.
func New() *TUI {
	return &TUI{}
}

// ShowTaskList prints a numbered list of tasks to stdout.
func (t *TUI) ShowTaskList(tasks []core.Task) error {
	fmt.Println("Tasks:")
	for i, task := range tasks {
		fmt.Printf("  %d. [%s] %s (priority: %d, source: %s)\n",
			i+1, task.Status, task.Title, task.Priority, task.Source)
	}
	return nil
}

// PromptCommitments presents the task list and reads the user's selection from
// stdin. The user enters a comma-separated list of 1-based indices. Exactly
// 2 or 3 tasks must be selected; invalid input is rejected with a retry.
func (t *TUI) PromptCommitments(tasks []core.Task) ([]core.Task, error) {
	fmt.Println("\nSelect 2-3 tasks to commit to today (e.g. 1,3):")
	for i, task := range tasks {
		fmt.Printf("  [%d] %-50s urgency: %.1f  (%s)\n",
			i+1, task.Title, task.Urgency, task.Source)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nYour selection: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read input: %w", err)
		}
		selected, err := parseSelection(strings.TrimSpace(line), len(tasks))
		if err != nil {
			fmt.Printf("Invalid selection: %v. Please try again.\n", err)
			continue
		}
		if len(selected) < 2 || len(selected) > 3 {
			fmt.Println("Please select between 2 and 3 tasks.")
			continue
		}
		result := make([]core.Task, len(selected))
		for i, idx := range selected {
			result[i] = tasks[idx-1]
		}
		return result, nil
	}
}

// parseSelection parses a comma-separated list of 1-based indices, returning
// an error if any index is out of range or duplicated.
func parseSelection(s string, max int) ([]int, error) {
	parts := strings.Split(s, ",")
	seen := map[int]bool{}
	var indices []int
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 1 || n > max {
			return nil, fmt.Errorf("invalid index %q (valid: 1-%d)", p, max)
		}
		if seen[n] {
			return nil, fmt.Errorf("duplicate index %d", n)
		}
		seen[n] = true
		indices = append(indices, n)
	}
	return indices, nil
}

// ShowPlanConfirmation prints the confirmed day plan and streak info.
func (t *TUI) ShowPlanConfirmation(plan *core.DayPlan, streak *core.Streak) error {
	if plan == nil {
		return fmt.Errorf("ShowPlanConfirmation: nil plan")
	}
	fmt.Printf("Plan confirmed for %s\n", plan.Date.Format("2006-01-02"))
	fmt.Println("Committed tasks:")
	for _, task := range plan.Tasks {
		fmt.Printf("  - [%s] %s\n", task.Status, task.Title)
	}
	if streak != nil {
		fmt.Printf("Current streak: %d day(s) (longest: %d)\n", streak.Current, streak.Longest)
	}
	return nil
}

// ShowNudge prints a mid-day reminder of the user's commitments.
func (t *TUI) ShowNudge(commitments []core.Task) error {
	fmt.Println("Reminder: your commitments for today:")
	for _, task := range commitments {
		fmt.Printf("  - [%s] %s\n", task.Status, task.Title)
	}
	return nil
}

// ShowEODSummary prints the end-of-day review summary.
func (t *TUI) ShowEODSummary(plan *core.DayPlan, score int) error {
	if plan == nil {
		return fmt.Errorf("ShowEODSummary: nil plan")
	}
	fmt.Printf("End of day summary for %s\n", plan.Date.Format("2006-01-02"))
	fmt.Println("Tasks:")
	for _, task := range plan.Tasks {
		fmt.Printf("  - [%s] %s\n", task.Status, task.Title)
	}
	fmt.Printf("Score: %d\n", score)
	return nil
}

// ShowMessage prints a plain message to stdout.
func (t *TUI) ShowMessage(msg string) error {
	fmt.Println(msg)
	return nil
}

// ShowError prints an error to stderr.
func (t *TUI) ShowError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// Close is a no-op for the stdout stub.
func (t *TUI) Close() error {
	return nil
}
