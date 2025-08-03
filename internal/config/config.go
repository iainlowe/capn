package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
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

// Config is the main configuration structure
type Config struct {
	Global  GlobalConfig  `yaml:"global"`
	Captain CaptainConfig `yaml:"captain"`
	Crew    CrewConfig    `yaml:"crew"`
	MCP     MCPConfig     `yaml:"mcp"`
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

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Captain.MaxConcurrentAgents <= 0 {
		return fmt.Errorf("max_concurrent_agents must be positive")
	}

	if c.Captain.PlanningTimeout <= 0 {
		return fmt.Errorf("planning_timeout must be positive")
	}

	if c.Global.Parallel <= 0 {
		return fmt.Errorf("parallel must be positive")
	}

	return nil
}
