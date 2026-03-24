package taskwarrior

import "os/exec"

// Commander runs external commands and returns their combined output.
// The interface exists so tests can inject a fake without spawning real processes.
type Commander interface {
	Run(name string, args ...string) ([]byte, error)
}

// ExecCommander is the real implementation using os/exec.
type ExecCommander struct{}

// Run executes name with the given args and returns its stdout output.
func (ExecCommander) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}
