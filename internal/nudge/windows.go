package nudge

import (
	"fmt"
	"time"

	"github.com/EwanGreer/day-planner/internal/core"
)

// WindowChecker determines whether a given time falls within any configured
// focus or DND window. It is stateless and safe for concurrent use.
type WindowChecker struct {
	windows []core.FocusWindow
}

// NewWindowChecker creates a WindowChecker from the config's focus windows.
// It validates that each window's Start and End are in valid HH:MM format.
func NewWindowChecker(windows []core.FocusWindow) (*WindowChecker, error) {
	for i, w := range windows {
		if _, _, err := parseTimeOfDay(w.Start); err != nil {
			return nil, fmt.Errorf("focus_windows[%d]: invalid start %q: %w", i, w.Start, err)
		}
		if _, _, err := parseTimeOfDay(w.End); err != nil {
			return nil, fmt.Errorf("focus_windows[%d]: invalid end %q: %w", i, w.End, err)
		}
	}
	return &WindowChecker{windows: windows}, nil
}

// IsBlocked reports whether t falls within any configured focus window.
// The window interval is start-inclusive, end-exclusive: [start, end).
// Returns false when no windows are configured.
func (wc *WindowChecker) IsBlocked(t time.Time) bool {
	return wc.ActiveWindow(t) != nil
}

// ActiveWindow returns the first window that contains t, or nil if none.
// The window interval is start-inclusive, end-exclusive: [start, end).
func (wc *WindowChecker) ActiveWindow(t time.Time) *core.FocusWindow {
	h, m := t.Hour(), t.Minute()
	current := h*60 + m

	for i := range wc.windows {
		sh, sm, _ := parseTimeOfDay(wc.windows[i].Start)
		eh, em, _ := parseTimeOfDay(wc.windows[i].End)
		start := sh*60 + sm
		end := eh*60 + em
		if current >= start && current < end {
			return &wc.windows[i]
		}
	}
	return nil
}

// StatusMessage returns a human-readable string describing whether the user
// is currently in a focus window. Used by `day-planner status`.
func (wc *WindowChecker) StatusMessage(t time.Time) string {
	if w := wc.ActiveWindow(t); w != nil {
		return fmt.Sprintf("In focus window: %s (%s-%s)", w.Label, w.Start, w.End)
	}
	return "No active focus window"
}

// parseTimeOfDay parses an "HH:MM" string into hour and minute integers.
func parseTimeOfDay(s string) (hour, minute int, err error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, 0, fmt.Errorf("expected HH:MM 24-hour format")
	}
	return t.Hour(), t.Minute(), nil
}
