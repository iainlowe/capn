package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_HelpCommand(t *testing.T) {
	var buf bytes.Buffer
	cli := NewCLI()
	cli.SetOutput(&buf)
	
	args := []string{"--help"}
	err := cli.Parse(args)
	
	// Help should return an error in Kong (it's expected)
	assert.Error(t, err)
	
	output := buf.String()
	assert.Contains(t, output, "capn")
	assert.Contains(t, output, "Flags")
	assert.Contains(t, output, "--config")
	assert.Contains(t, output, "--verbose")
	assert.Contains(t, output, "--dry-run")
	assert.Contains(t, output, "--parallel")
	assert.Contains(t, output, "--timeout")
	assert.Contains(t, output, "Commands")
	assert.Contains(t, output, "execute")
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "agents")
	assert.Contains(t, output, "mcp")
}

func TestCLI_GlobalOptions(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		checkFn  func(t *testing.T, options *GlobalOptions)
	}{
		{
			name: "verbose flag",
			args: []string{"--verbose", "execute", "test"},
			checkFn: func(t *testing.T, options *GlobalOptions) {
				assert.True(t, options.Verbose)
			},
		},
		{
			name: "verbose short flag",
			args: []string{"-v", "execute", "test"},
			checkFn: func(t *testing.T, options *GlobalOptions) {
				assert.True(t, options.Verbose)
			},
		},
		{
			name: "dry-run flag",
			args: []string{"--dry-run", "execute", "test"},
			checkFn: func(t *testing.T, options *GlobalOptions) {
				assert.True(t, options.DryRun)
			},
		},
		{
			name: "parallel flag",
			args: []string{"--parallel", "10", "execute", "test"},
			checkFn: func(t *testing.T, options *GlobalOptions) {
				assert.Equal(t, 10, options.Parallel)
			},
		},
		{
			name: "config flag",
			args: []string{"--config", "/path/to/config.yaml", "status"},
			checkFn: func(t *testing.T, options *GlobalOptions) {
				assert.Equal(t, "/path/to/config.yaml", options.Config)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI()
			cli.SetSkipConfigForTests(true) // Skip config loading for all flag tests
			
			// Capture the parsed options
			var capturedOptions *GlobalOptions
			cli.SetGlobalOptionsCallback(func(opts *GlobalOptions) {
				capturedOptions = opts
			})
			
			err := cli.Parse(tt.args)
			require.NoError(t, err)
			require.NotNil(t, capturedOptions)
			
			tt.checkFn(t, capturedOptions)
		})
	}
}

func TestCLI_SubCommands(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "execute command (planning mode)",
			args:        []string{"execute", "--plan-only", "test goal"},
			expectError: false,
		},
		{
			name:        "execute command (execution mode)",
			args:        []string{"execute", "test goal"},
			expectError: false,
		},
		{
			name:        "status command",
			args:        []string{"status"},
			expectError: false,
		},
		{
			name:        "agents command",
			args:        []string{"agents"},
			expectError: false,
		},
		{
			name:        "mcp command",
			args:        []string{"mcp"},
			expectError: false,
		},
		{
			name:        "unknown command",
			args:        []string{"unknown"},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI()
			err := cli.Parse(tt.args)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCLI_ConfigFileLoading(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")
	
	configContent := `
global:
  verbose: true
  parallel: 8
  timeout: 120s

captain:
  max_concurrent_agents: 10
  planning_timeout: 45s
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)
	
	cli := NewCLI()
	
	var capturedOptions *GlobalOptions
	cli.SetGlobalOptionsCallback(func(opts *GlobalOptions) {
		capturedOptions = opts
	})
	
	args := []string{"--config", configFile, "execute", "test"}
	err = cli.Parse(args)
	
	require.NoError(t, err)
	require.NotNil(t, capturedOptions)
	
	// Config file should be loaded
	assert.Equal(t, configFile, capturedOptions.Config)
}

func TestCLI_InvalidConfig(t *testing.T) {
	cli := NewCLI()
	
	args := []string{"--config", "/non/existent/config.yaml", "execute", "test"}
	err := cli.Parse(args)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config")
}
