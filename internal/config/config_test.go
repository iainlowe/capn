package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_DefaultValues(t *testing.T) {
	cfg := NewConfig()
	
	// Test default values are set correctly
	assert.Equal(t, 5, cfg.Captain.MaxConcurrentAgents)
	assert.Equal(t, 30*time.Second, cfg.Captain.PlanningTimeout)
	assert.Equal(t, 3, cfg.MCP.RetryCount)
	assert.Equal(t, 10*time.Second, cfg.MCP.Timeout)
	assert.False(t, cfg.Global.Verbose)
	assert.False(t, cfg.Global.DryRun)
	assert.Equal(t, 5, cfg.Global.Parallel)
	assert.Equal(t, 5*time.Minute, cfg.Global.Timeout)
}

func TestConfig_LoadFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `
global:
  verbose: true
  dry_run: true
  parallel: 10
  timeout: 600s

captain:
  max_concurrent_agents: 8
  planning_timeout: 60s

crew:
  timeouts:
    research: 300s
    code: 600s

mcp:
  timeout: 15s
  retry_count: 5
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Load config from file
	cfg, err := LoadConfig(configFile)
	require.NoError(t, err)
	
	// Verify loaded values
	assert.True(t, cfg.Global.Verbose)
	assert.True(t, cfg.Global.DryRun)
	assert.Equal(t, 10, cfg.Global.Parallel)
	assert.Equal(t, 10*time.Minute, cfg.Global.Timeout)
	assert.Equal(t, 8, cfg.Captain.MaxConcurrentAgents)
	assert.Equal(t, 60*time.Second, cfg.Captain.PlanningTimeout)
	assert.Equal(t, 15*time.Second, cfg.MCP.Timeout)
	assert.Equal(t, 5, cfg.MCP.RetryCount)
}

func TestConfig_LoadNonExistentFile(t *testing.T) {
	_, err := LoadConfig("/non/existent/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")
	
	// Create invalid YAML
	invalidYAML := `
global:
  verbose: true
  invalid_yaml: [
`
	
	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	require.NoError(t, err)
	
	_, err = LoadConfig(configFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      NewConfig(),
			expectError: false,
		},
		{
			name: "negative max concurrent agents",
			config: &Config{
				Captain: CaptainConfig{
					MaxConcurrentAgents: -1,
				},
			},
			expectError: true,
			errorMsg:    "max_concurrent_agents must be positive",
		},
		{
			name: "zero planning timeout",
			config: &Config{
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,
					PlanningTimeout:     0,
				},
			},
			expectError: true,
			errorMsg:    "planning_timeout must be positive",
		},
		{
			name: "negative parallel setting",
			config: &Config{
				Global: GlobalConfig{
					Parallel: -1,
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,
					PlanningTimeout:     30 * time.Second,
				},
			},
			expectError: true,
			errorMsg:    "parallel must be positive",
		},
		{
			name: "OpenAI config with valid settings",
			config: &Config{
				Global: GlobalConfig{
					Parallel: 5,
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,
					PlanningTimeout:     30 * time.Second,
				},
				OpenAI: OpenAIConfig{
					APIKey:      "test-key",
					Model:       "gpt-3.5-turbo",
					Temperature: 0.7,
					MaxRetries:  3,
				},
			},
			expectError: false,
		},
		{
			name: "OpenAI config with missing model",
			config: &Config{
				Global: GlobalConfig{
					Parallel: 5,
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,
					PlanningTimeout:     30 * time.Second,
				},
				OpenAI: OpenAIConfig{
					APIKey: "test-key",
					Model:  "", // missing model
				},
			},
			expectError: true,
			errorMsg:    "openai model cannot be empty when api_key is set",
		},
		{
			name: "OpenAI config with invalid temperature",
			config: &Config{
				Global: GlobalConfig{
					Parallel: 5,
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,
					PlanningTimeout:     30 * time.Second,
				},
				OpenAI: OpenAIConfig{
					APIKey:      "test-key",
					Model:       "gpt-3.5-turbo",
					Temperature: 2.0, // invalid temperature > 1
				},
			},
			expectError: true,
			errorMsg:    "openai temperature must be between 0 and 1",
		},
		{
			name: "OpenAI config with negative max retries",
			config: &Config{
				Global: GlobalConfig{
					Parallel: 5,
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,
					PlanningTimeout:     30 * time.Second,
				},
				OpenAI: OpenAIConfig{
					APIKey:     "test-key",
					Model:      "gpt-3.5-turbo",
					MaxRetries: -1, // negative retries
				},
			},
			expectError: true,
			errorMsg:    "openai max_retries cannot be negative",
		},
		{
			name: "OpenAI config with empty API key (should not validate OpenAI)",
			config: &Config{
				Global: GlobalConfig{
					Parallel: 5,
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,
					PlanningTimeout:     30 * time.Second,
				},
				OpenAI: OpenAIConfig{
					APIKey: "", // empty API key, so OpenAI validation should be skipped
					Model:  "", // this would normally be invalid, but should be ignored
				},
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
