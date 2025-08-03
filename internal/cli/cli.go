package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/iainlowe/capn/internal/captain"
	"github.com/iainlowe/capn/internal/config"
)

// GlobalOptions holds all global command-line options
type GlobalOptions struct {
	Config   string        `help:"Configuration file path" short:"c"`
	Verbose  bool          `help:"Verbose logging" short:"v"`
	DryRun   bool          `help:"Plan without execution"`
	Parallel int           `help:"Maximum parallel agents" short:"p" default:"5"`
	Timeout  time.Duration `help:"Global timeout duration" default:"5m"`
}

// ExecuteCmd represents the execute command (with optional planning mode)
type ExecuteCmd struct {
	PlanOnly bool   `help:"Plan only, don't execute" short:"n" name:"plan-only"`
	Goal     string `arg:"" help:"Goal to execute"`
}

func (e *ExecuteCmd) Run(globals *GlobalOptions, logger *zap.Logger, config *config.Config) error {
	// Check if we're in planning mode (plan-only or global dry-run)
	planningMode := e.PlanOnly || globals.DryRun
	
	// Check if OpenAI is configured (either in config or environment)
	openaiAPIKey := config.OpenAI.APIKey
	if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		openaiAPIKey = envKey
	}
	
	if openaiAPIKey == "" {
		if planningMode {
			logger.Info("Creating basic plan (OpenAI not configured)", zap.String("goal", e.Goal))
			fmt.Printf("Planning: %s\n", e.Goal)
			fmt.Printf("Note: Set OPENAI_API_KEY environment variable or configure OpenAI in config file for LLM-powered planning.\n")
			return nil
		} else {
			logger.Info("Basic execution (OpenAI not configured)", zap.String("goal", e.Goal))
			fmt.Printf("Executing: %s\n", e.Goal)
			fmt.Printf("Note: Set OPENAI_API_KEY environment variable or configure OpenAI in config file for intelligent planning.\n")
			return nil
		}
	}

	// Create OpenAI config with proper priority (env > config)
	openaiConfig := captain.OpenAIConfig{
		APIKey:      config.OpenAI.APIKey,
		Model:       config.OpenAI.Model,
		BaseURL:     config.OpenAI.BaseURL,
		MaxRetries:  config.OpenAI.MaxRetries,
		Temperature: config.OpenAI.Temperature,
	}
	
	if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		openaiConfig.APIKey = envKey
	}

	// Create Captain
	cap, err := captain.NewCaptain("main-captain", config, openaiConfig)
	if err != nil {
		return fmt.Errorf("failed to create captain: %w", err)
	}
	defer cap.Stop()

	// Create execution plan
	logger.Info("Creating execution plan", zap.String("goal", e.Goal))
	ctx := context.Background()
	plan, err := cap.CreatePlan(ctx, e.Goal)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	if planningMode {
		logger.Info("Plan created successfully", zap.String("plan_id", plan.ID))
		fmt.Printf("=== Execution Plan ===\n")
		fmt.Printf("Goal: %s\n", plan.Goal)
		fmt.Printf("Strategy: %s\n", plan.Strategy.Type)
		fmt.Printf("Estimated Duration: %s\n", plan.Timeline.EstimatedDuration)
		fmt.Printf("Tasks (%d):\n", len(plan.Tasks))
		
		for _, task := range plan.Tasks {
			fmt.Printf("  [%s] %s (Priority: %s)\n", 
				task.Type, task.Payload["description"], task.Priority)
			if len(task.Dependencies) > 0 {
				fmt.Printf("     Dependencies: %v\n", task.Dependencies)
			}
		}
		fmt.Printf("\nNote: This is a dry run. Use without --plan-only or --dry-run to execute.\n")
	} else {
		logger.Info("Executing plan", zap.String("plan_id", plan.ID))
		fmt.Printf("Executing plan: %s\n", plan.Goal)
		
		result, err := cap.ExecutePlan(ctx, plan, false)
		if err != nil {
			return fmt.Errorf("failed to execute plan: %w", err)
		}

		fmt.Printf("=== Execution Results ===\n")
		fmt.Printf("Plan: %s\n", result.PlanID)
		fmt.Printf("Success: %t\n", result.Success)
		fmt.Printf("Duration: %s\n", result.Duration)
		fmt.Printf("Tasks completed: %d\n", len(result.TaskResults))
		
		for _, taskResult := range result.TaskResults {
			status := "✓"
			if !taskResult.Success {
				status = "✗"
			}
			fmt.Printf("  %s Task %s: %s\n", status, taskResult.TaskID, taskResult.Output)
		}
		
		if !result.Success {
			fmt.Printf("Execution completed with errors. Check logs for details.\n")
		}
	}
	
	return nil
}

// StatusCmd represents the status command
type StatusCmd struct{}

func (s *StatusCmd) Run(globals *GlobalOptions, logger *zap.Logger) error {
	logger.Info("Checking status")
	// TODO: Implement status logic
	return nil
}

// AgentsCmd represents the agents command
type AgentsCmd struct{}

func (a *AgentsCmd) Run(globals *GlobalOptions, logger *zap.Logger) error {
	logger.Info("Managing agents")
	fmt.Println("Managing agents")
	// TODO: Implement agents management logic
	return nil
}

// MCPCmd represents the mcp command
type MCPCmd struct{}

func (m *MCPCmd) Run(globals *GlobalOptions, logger *zap.Logger) error {
	logger.Info("Managing MCP servers")
	fmt.Println("Managing MCP servers")
	// TODO: Implement MCP server management logic
	return nil
}

// CLI represents the main CLI structure
type CLI struct {
	GlobalOptions

	Execute ExecuteCmd `cmd:"" help:"Plan and execute goals (use --dry-run for planning only)"`
	Status  StatusCmd  `cmd:"" help:"Show current operation status"`
	Agents  AgentsCmd  `cmd:"" help:"Manage agent configurations"`
	MCP     MCPCmd     `cmd:"" help:"Manage MCP server connections"`

	output       io.Writer
	logger       *zap.Logger
	config       *config.Config
	callback     func(*GlobalOptions)
	exitOverride bool
	skipConfig   bool // Skip config loading for tests
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		output:       os.Stdout,
		exitOverride: true, // Default to true for tests
	}
}

// SetOutput sets the output writer for the CLI
func (c *CLI) SetOutput(w io.Writer) {
	c.output = w
}

// SetExitOverride sets whether to override os.Exit for tests
func (c *CLI) SetExitOverride(override bool) {
	c.exitOverride = override
}

// SetGlobalOptionsCallback sets a callback to capture global options for testing
func (c *CLI) SetGlobalOptionsCallback(callback func(*GlobalOptions)) {
	c.callback = callback
}

// SetSkipConfigForTests disables config file loading for testing
func (c *CLI) SetSkipConfigForTests(skip bool) {
	c.skipConfig = skip
}

// Parse runs the CLI with the given arguments
func (c *CLI) Parse(args []string) error {
	// Initialize logger first (needed for binding)
	c.logger = c.createLogger()
	
	// Create parser with bindings for command methods
	options := []kong.Option{
		kong.Name("capn"),
		kong.Description("Distributed CLI Agent System"),
		kong.UsageOnError(),
		kong.Writers(c.output, c.output),
		kong.Bind(&c.GlobalOptions), // Bind global options
		kong.Bind(c.logger),         // Bind logger
	}
	
	// Add exit override for tests
	if c.exitOverride {
		options = append(options, kong.Exit(func(int) {}))
	}
	
	parser := kong.Must(c, options...)
	
	// Parse command line arguments
	ctx, err := parser.Parse(args)
	if err != nil {
		return err
	}
	
	// Load configuration if specified
	if c.Config != "" && !c.skipConfig {
		c.config, err = config.LoadConfig(c.Config)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		
		// Merge config file values with command line options
		c.mergeConfigWithOptions()
	} else {
		// Use default configuration
		c.config = config.NewConfig()
		c.mergeOptionsWithConfig()
	}
	
	// Bind config for commands that need it
	ctx.Bind(c.config)
	
	// Call callback for testing
	if c.callback != nil {
		c.callback(&c.GlobalOptions)
	}
	
	// Run the selected command
	return ctx.Run()
}

// createLogger creates a zap logger based on verbose setting
func (c *CLI) createLogger() *zap.Logger {
	var logger *zap.Logger
	
	if c.Verbose {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, _ = config.Build()
	} else {
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		logger, _ = config.Build()
	}
	
	return logger
}

// mergeConfigWithOptions merges configuration file values with command line options
func (c *CLI) mergeConfigWithOptions() {
	// Command line options take precedence over config file
	// Only update from config if the option wasn't explicitly set on command line
	
	if !c.wasSetExplicitly("verbose") {
		c.Verbose = c.config.Global.Verbose
	}
	if !c.wasSetExplicitly("dry-run") {
		c.DryRun = c.config.Global.DryRun
	}
	if !c.wasSetExplicitly("parallel") {
		c.Parallel = c.config.Global.Parallel
	}
	if !c.wasSetExplicitly("timeout") {
		c.Timeout = c.config.Global.Timeout
	}
}

// mergeOptionsWithConfig updates config with command line options
func (c *CLI) mergeOptionsWithConfig() {
	c.config.Global.Verbose = c.Verbose
	c.config.Global.DryRun = c.DryRun
	c.config.Global.Parallel = c.Parallel
	c.config.Global.Timeout = c.Timeout
	c.config.Global.Config = c.Config
}

// wasSetExplicitly checks if an option was explicitly set on command line
// For now, this is a simplified implementation
func (c *CLI) wasSetExplicitly(option string) bool {
	// TODO: Implement proper detection of explicitly set flags
	// This would require extending Kong or tracking flag usage
	return false
}
