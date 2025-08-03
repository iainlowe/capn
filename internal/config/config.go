package config

import (
	"fmt"
	"os"
	"time"

	"github.com/iainlowe/capn/internal/common"
	yaml "gopkg.in/yaml.v3"
)

// GlobalConfig holds global command-line options
type GlobalConfig struct {
	Config   string        `yaml:"config" kong:"help='Configuration file path',short='c'"`
	Verbose  bool          `yaml:"verbose" kong:"help='Verbose logging',short='v'"`
	DryRun   bool          `yaml:"dry_run" kong:"help='Plan without execution'"`
	Parallel int           `yaml:"parallel" kong:"help='Maximum parallel agents',short='p',default='5'"`
	Timeout  time.Duration `yaml:"timeout" kong:"help='Global timeout duration',default='5m'"`
}

// CaptainConfig holds Captain agent configuration
type CaptainConfig struct {
	MaxConcurrentAgents int           `yaml:"max_concurrent_agents"`
	PlanningTimeout     time.Duration `yaml:"planning_timeout"`
}

// CrewConfig holds Crew agent configuration
type CrewConfig struct {
	Timeouts map[string]time.Duration `yaml:"timeouts"`
}

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Timeout    time.Duration `yaml:"timeout"`
	RetryCount int           `yaml:"retry_count"`
}

// OpenAIConfig holds OpenAI configuration
type OpenAIConfig struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	BaseURL     string  `yaml:"base_url,omitempty"`
	MaxRetries  int     `yaml:"max_retries"`
	Temperature float64 `yaml:"temperature"`
}

// Config is the main configuration structure
type Config struct {
	Global  GlobalConfig  `yaml:"global"`
	Captain CaptainConfig `yaml:"captain"`
	Crew    CrewConfig    `yaml:"crew"`
	MCP     MCPConfig     `yaml:"mcp"`
	OpenAI  OpenAIConfig  `yaml:"openai"`
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		Global: GlobalConfig{
			Verbose:  false,
			DryRun:   false,
			Parallel: 5,
			Timeout:  5 * time.Minute,
		},
		Captain: CaptainConfig{
			MaxConcurrentAgents: 5,
			PlanningTimeout:     30 * time.Second,
		},
		Crew: CrewConfig{
			Timeouts: make(map[string]time.Duration),
		},
		MCP: MCPConfig{
			Timeout:    10 * time.Second,
			RetryCount: 3,
		},
		OpenAI: OpenAIConfig{
			Model:       "gpt-3.5-turbo",
			MaxRetries:  3,
			Temperature: 0.7,
		},
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	config := NewConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration using the common validation framework
func (c *Config) Validate() error {
	validator := common.NewValidator()
	validator.AddRule("max_concurrent_agents", common.Positive("max_concurrent_agents"))
	validator.AddRule("planning_timeout", common.Positive("planning_timeout"))
	validator.AddRule("parallel", common.Positive("parallel"))

	// Validate basic fields
	err := validator.Validate(map[string]interface{}{
		"max_concurrent_agents": c.Captain.MaxConcurrentAgents,
		"planning_timeout":      c.Captain.PlanningTimeout.Nanoseconds(),
		"parallel":              c.Global.Parallel,
	})

	if err != nil {
		return err
	}

	// Validate OpenAI config if API key is provided
	if c.OpenAI.APIKey != "" {
		openaiValidator := common.NewValidator()
		openaiValidator.AddRule("model", common.Required("openai model"))
		openaiValidator.AddRule("temperature", common.Range("openai temperature", 0, 1))
		openaiValidator.AddRule("max_retries", func(value interface{}) error {
			if maxRetries, ok := value.(int); ok && maxRetries < 0 {
				return fmt.Errorf("openai max_retries cannot be negative")
			}
			return nil
		})

		return openaiValidator.Validate(map[string]interface{}{
			"model":       c.OpenAI.Model,
			"temperature": c.OpenAI.Temperature,
			"max_retries": c.OpenAI.MaxRetries,
		})
	}

	return nil
}
