package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iainlowe/capn/internal/testutil"
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
	testCases := []testutil.ValidationTestCase[*Config]{
		{
			Name:      "valid config",
			Input:     NewConfig(),
			WantError: false,
		},
		{
			Name: "negative max concurrent agents",
			Input: &Config{
				Global: GlobalConfig{
					Parallel: 5, // Set valid value
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: -1,                   // This should trigger the error
					PlanningTimeout:     30 * time.Second, // Set valid value
				},
			},
			WantError: true,
			ErrorMsg:  "max_concurrent_agents must be positive",
		},
		{
			Name: "zero planning timeout",
			Input: &Config{
				Global: GlobalConfig{
					Parallel: 5, // Set valid parallel value
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5, // Set valid max concurrent agents
					PlanningTimeout:     0,  // This should trigger the error
				},
			},
			WantError: true,
			ErrorMsg:  "planning_timeout must be positive",
		},
		{
			Name: "negative parallel setting",
			Input: &Config{
				Global: GlobalConfig{
					Parallel: -1, // This should trigger the error
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,                    // Set valid value
					PlanningTimeout:     30 * time.Second, // Set valid value
				},
			},
			WantError: true,
			ErrorMsg:  "parallel must be positive",
		},
		{
			Name: "OpenAI config with valid settings",
			Input: &Config{
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
			WantError: false,
		},
		{
			Name: "OpenAI config with missing model",
			Input: &Config{
				Global: GlobalConfig{
					Parallel: 5,
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,
					PlanningTimeout:     30 * time.Second,
				},
				OpenAI: OpenAIConfig{
					APIKey: "test-key",
					Model:  "",
				},
			},
			WantError: true,
			ErrorMsg:  "openai model cannot be empty",
		},
		{
			Name: "OpenAI config with invalid temperature",
			Input: &Config{
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
					Temperature: 2.0,
				},
			},
			WantError: true,
			ErrorMsg:  "openai temperature must be between 0 and 1",
		},
		{
			Name: "OpenAI config with negative max retries",
			Input: &Config{
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
					MaxRetries: -1,
				},
			},
			WantError: true,
			ErrorMsg:  "openai max_retries cannot be negative",
		},
		{
			Name: "OpenAI config with empty API key",
			Input: &Config{
				Global: GlobalConfig{
					Parallel: 5,
				},
				Captain: CaptainConfig{
					MaxConcurrentAgents: 5,
					PlanningTimeout:     30 * time.Second,
				},
				OpenAI: OpenAIConfig{
					APIKey: "",
					Model:  "",
				},
			},
			WantError: false,
		},
	}

	testutil.RunValidationTests(t, testCases, func(cfg *Config) error {
		return cfg.Validate()
	})
}
