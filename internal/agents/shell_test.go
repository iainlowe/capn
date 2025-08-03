package agents

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShellCommand_String(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		timeout  time.Duration
		args     []string
		expected string
	}{
		{
			name:     "simple command",
			command:  "ls",
			timeout:  10 * time.Second,
			args:     []string{"-la"},
			expected: "timeout 10s ls -la",
		},
		{
			name:     "test command",
			command:  "go",
			timeout:  300 * time.Second,
			args:     []string{"test", "./..."},
			expected: "timeout 300s go test ./...",
		},
		{
			name:     "build command",
			command:  "go",
			timeout:  600 * time.Second,
			args:     []string{"build", "./cmd/capn"},
			expected: "timeout 600s go build ./cmd/capn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewShellCommand(tt.command, tt.timeout, tt.args...)
			assert.Equal(t, tt.expected, cmd.String())
		})
	}
}

func TestShellCommand_Execute(t *testing.T) {
	// Test with a simple echo command
	cmd := NewShellCommand("echo", 10*time.Second, "hello", "world")
	output, err := cmd.Execute()
	
	assert.NoError(t, err)
	assert.Equal(t, "hello world", output)
}

func TestShellCommand_NewShellCommand(t *testing.T) {
	cmd := NewShellCommand("ls", 30*time.Second, "-la", "/tmp")
	
	assert.Equal(t, "ls", cmd.Command)
	assert.Equal(t, 30*time.Second, cmd.Timeout)
	assert.Equal(t, []string{"-la", "/tmp"}, cmd.Args)
}