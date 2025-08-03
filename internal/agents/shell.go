package agents

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ShellCommand represents a shell command with timeout
type ShellCommand struct {
	Command string
	Timeout time.Duration
	Args    []string
}

// NewShellCommand creates a shell command with timeout prefix
func NewShellCommand(command string, timeout time.Duration, args ...string) *ShellCommand {
	return &ShellCommand{
		Command: command,
		Timeout: timeout,
		Args:    args,
	}
}

// Execute runs the shell command with timeout prefix
func (sc *ShellCommand) Execute() (string, error) {
	// Build command with timeout prefix
	timeoutStr := fmt.Sprintf("%.0fs", sc.Timeout.Seconds())
	fullArgs := append([]string{timeoutStr, sc.Command}, sc.Args...)
	
	cmd := exec.Command("timeout", fullArgs...)
	output, err := cmd.CombinedOutput()
	
	return strings.TrimSpace(string(output)), err
}

// String returns the command as it would be executed
func (sc *ShellCommand) String() string {
	timeoutStr := fmt.Sprintf("%.0fs", sc.Timeout.Seconds())
	allArgs := append([]string{"timeout", timeoutStr, sc.Command}, sc.Args...)
	return strings.Join(allArgs, " ")
}