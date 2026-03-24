package nudge_test

import (
	"strings"
	"testing"
	"time"

	"github.com/EwanGreer/day-planner/internal/core"
	"github.com/EwanGreer/day-planner/internal/nudge"
)

// makeTime creates a time.Time for today at the given hour and minute.
func makeTime(hour, minute int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
}

// TestIsBlocked_NoWindows verifies that an empty config never blocks.
func TestIsBlocked_NoWindows(t *testing.T) {
	wc, err := nudge.NewWindowChecker(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc.IsBlocked(makeTime(10, 0)) {
		t.Error("IsBlocked: got true with no windows configured, want false")
	}
}

// TestIsBlocked_InsideWindow verifies that a time inside a window is blocked.
func TestIsBlocked_InsideWindow(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "09:00", End: "11:00", Label: "deep work"},
	}
	wc, err := nudge.NewWindowChecker(windows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wc.IsBlocked(makeTime(10, 0)) {
		t.Error("IsBlocked: got false for time inside window, want true")
	}
}

// TestIsBlocked_OutsideWindow verifies that a time outside all windows is not blocked.
func TestIsBlocked_OutsideWindow(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "09:00", End: "11:00", Label: "deep work"},
	}
	wc, err := nudge.NewWindowChecker(windows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc.IsBlocked(makeTime(12, 0)) {
		t.Error("IsBlocked: got true for time outside window, want false")
	}
}

// TestIsBlocked_AtStartBoundary verifies that a time exactly at start is blocked (inclusive).
func TestIsBlocked_AtStartBoundary(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "09:00", End: "11:00", Label: "deep work"},
	}
	wc, err := nudge.NewWindowChecker(windows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wc.IsBlocked(makeTime(9, 0)) {
		t.Error("IsBlocked: got false at start boundary, want true (start inclusive)")
	}
}

// TestIsBlocked_AtEndBoundary verifies that a time exactly at end is not blocked (exclusive).
func TestIsBlocked_AtEndBoundary(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "09:00", End: "11:00", Label: "deep work"},
	}
	wc, err := nudge.NewWindowChecker(windows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc.IsBlocked(makeTime(11, 0)) {
		t.Error("IsBlocked: got true at end boundary, want false (end exclusive)")
	}
}

// TestIsBlocked_OverlappingWindows verifies that a time in the second of two windows is blocked.
func TestIsBlocked_OverlappingWindows(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "09:00", End: "10:00", Label: "morning"},
		{Start: "14:00", End: "16:00", Label: "afternoon"},
	}
	wc, err := nudge.NewWindowChecker(windows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wc.IsBlocked(makeTime(15, 0)) {
		t.Error("IsBlocked: got false for time in second window, want true")
	}
	if wc.IsBlocked(makeTime(11, 0)) {
		t.Error("IsBlocked: got true for time between windows, want false")
	}
}

// TestActiveWindow_ReturnsFirst verifies that a time in the first window returns that window.
func TestActiveWindow_ReturnsFirst(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "09:00", End: "11:00", Label: "deep work"},
		{Start: "14:00", End: "16:00", Label: "review"},
	}
	wc, err := nudge.NewWindowChecker(windows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w := wc.ActiveWindow(makeTime(10, 0))
	if w == nil {
		t.Fatal("ActiveWindow: got nil, want window")
	}
	if w.Label != "deep work" {
		t.Errorf("ActiveWindow: got label %q, want %q", w.Label, "deep work")
	}
}

// TestActiveWindow_None verifies that a time outside all windows returns nil.
func TestActiveWindow_None(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "09:00", End: "11:00", Label: "deep work"},
	}
	wc, err := nudge.NewWindowChecker(windows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w := wc.ActiveWindow(makeTime(12, 0)); w != nil {
		t.Errorf("ActiveWindow: got %+v, want nil", w)
	}
}

// TestNewWindowChecker_InvalidFormat verifies that an invalid HH:MM string returns an error.
func TestNewWindowChecker_InvalidFormat(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "9am", End: "10:00", Label: "bad"},
	}
	_, err := nudge.NewWindowChecker(windows)
	if err == nil {
		t.Fatal("NewWindowChecker: expected error for invalid start format, got nil")
	}
}

// TestStatusMessage_InWindow verifies that the status message contains the label.
func TestStatusMessage_InWindow(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "09:00", End: "11:00", Label: "deep work"},
	}
	wc, err := nudge.NewWindowChecker(windows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	msg := wc.StatusMessage(makeTime(10, 0))
	if !strings.Contains(msg, "deep work") {
		t.Errorf("StatusMessage: %q does not contain label %q", msg, "deep work")
	}
}

// TestStatusMessage_NoWindow verifies the message when outside all windows.
func TestStatusMessage_NoWindow(t *testing.T) {
	windows := []core.FocusWindow{
		{Start: "09:00", End: "11:00", Label: "deep work"},
	}
	wc, err := nudge.NewWindowChecker(windows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	const want = "No active focus window"
	msg := wc.StatusMessage(makeTime(14, 0))
	if msg != want {
		t.Errorf("StatusMessage: got %q, want %q", msg, want)
	}
}
