package common_test

import (
	"fmt"
	"testing"

	"github.com/iainlowe/capn/internal/common"
	"github.com/iainlowe/capn/internal/testutil"
)

// Example config struct (similar to OpenAIConfig)
type ExampleConfig struct {
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxRetries  int     `json:"max_retries"`
}

// validateConfig demonstrates how to use the validation framework
func validateConfig(config ExampleConfig) error {
	validator := common.NewValidator()
	validator.AddRule("api_key", common.Required("api_key"))
	validator.AddRule("model", common.Required("model"))
	validator.AddRule("temperature", common.Range("temperature", 0, 1))
	validator.AddRule("max_retries", common.Positive("max_retries"))

	return validator.Validate(map[string]interface{}{
		"api_key":     config.APIKey,
		"model":       config.Model,
		"temperature": config.Temperature,
		"max_retries": config.MaxRetries,
	})
}

// Example of how the validation framework would replace existing validation code
func TestValidationFrameworkExample(t *testing.T) {

	// Test cases using the new test helper
	testCases := []testutil.ValidationTestCase[ExampleConfig]{
		{
			Name: "valid config",
			Input: ExampleConfig{
				APIKey:      "test-key",
				Model:       "gpt-3.5-turbo",
				Temperature: 0.7,
				MaxRetries:  3,
			},
			WantError: false,
		},
		{
			Name: "missing API key",
			Input: ExampleConfig{
				Model:       "gpt-3.5-turbo",
				Temperature: 0.7,
				MaxRetries:  3,
			},
			WantError: true,
			ErrorMsg:  "api_key cannot be empty",
		},
		{
			Name: "invalid temperature",
			Input: ExampleConfig{
				APIKey:      "test-key",
				Model:       "gpt-3.5-turbo",
				Temperature: 2.0,
				MaxRetries:  3,
			},
			WantError: true,
			ErrorMsg:  "temperature must be between",
		},
	}

	// Validation function using the new framework
	validateConfig := func(config ExampleConfig) error {
		validator := common.NewValidator()
		validator.AddRule("api_key", common.Required("api_key"))
		validator.AddRule("model", common.Required("model"))
		validator.AddRule("temperature", common.Range("temperature", 0, 1))
		validator.AddRule("max_retries", common.Positive("max_retries"))

		return validator.Validate(map[string]interface{}{
			"api_key":     config.APIKey,
			"model":       config.Model,
			"temperature": config.Temperature,
			"max_retries": config.MaxRetries,
		})
	}

	// Single line to run all test cases
	testutil.RunValidationTests(t, testCases, validateConfig)
}

// Example of how the config builder would replace existing config creation
func TestConfigBuilderExample(t *testing.T) {
	type DatabaseConfig struct {
		Host     string
		Port     int
		Username string
		Password string
		Timeout  int
	}

	// Before: Manual config creation with validation scattered
	// After: Fluent builder with centralized validation
	config, err := common.NewConfigBuilder(DatabaseConfig{
		Host:    "localhost", // default
		Port:    5432,        // default
		Timeout: 30,          // default
	}).
		With(func(c *DatabaseConfig) {
			c.Username = "user"
			c.Password = "secret"
		}).
		Validate(func(c DatabaseConfig) error {
			if c.Username == "" {
				return fmt.Errorf("username required")
			}
			if c.Password == "" {
				return fmt.Errorf("password required")
			}
			return nil
		}).
		Build()

	if err != nil {
		t.Fatalf("failed to build config: %v", err)
	}

	// Config is now validated and ready to use
	if config.Username != "user" {
		t.Errorf("expected username 'user', got %s", config.Username)
	}
}
