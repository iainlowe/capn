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

// PlanCmd represents the plan command for goal analysis and execution planning
type PlanCmd struct {
	Goal string `arg:"" help:"Goal to analyze and create execution plan for"`
}

func (p *PlanCmd) Run(globals *GlobalOptions, logger *zap.Logger) error {
	logger.Info("Creating execution plan", zap.String("goal", p.Goal))
	
	// Load configuration
	var cfg *config.Config
	var err error
	
	if globals.Config != "" {
		cfg, err = config.LoadConfig(globals.Config)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		cfg = config.NewConfig()
	}

	// Check if OpenAI API key is configured
	if cfg.Captain.OpenAIAPIKey == "" {
		fmt.Println("OpenAI API key not configured. Set it in config file or OPENAI_API_KEY environment variable.")
		fmt.Println("Running in demonstration mode with mock planning...")
		
		// Show what the plan would look like
		fmt.Printf("\n=== EXECUTION PLAN FOR: %s ===\n", p.Goal)
		fmt.Println("Goal Analysis: This would use OpenAI to break down the goal into actionable tasks")
		fmt.Println("\nSample Plan Structure:")
		fmt.Println("1. Task decomposition using chain-of-thought reasoning")
		fmt.Println("2. Dependency analysis and task ordering") 
		fmt.Println("3. Resource allocation and timing estimates")
		fmt.Println("4. Validation and feasibility checking")
		fmt.Println("\nTo enable full LLM-powered planning, configure your OpenAI API key.")
		return nil
	}

	// Create captain configuration
	captainConfig := &captain.Config{
		OpenAIAPIKey:        cfg.Captain.OpenAIAPIKey,
		Model:               cfg.Captain.Model,
		MaxTokens:           cfg.Captain.MaxTokens,
		Temperature:         cfg.Captain.Temperature,
		MaxConcurrentAgents: cfg.Captain.MaxConcurrentAgents,
		PlanningTimeout:     cfg.Captain.PlanningTimeout,
		RetryAttempts:       cfg.Captain.RetryAttempts,
		RetryDelay:          cfg.Captain.RetryDelay,
	}

	// Create OpenAI provider
	llmProvider, err := captain.NewOpenAIProvider(captainConfig)
	if err != nil {
		return fmt.Errorf("failed to create OpenAI provider: %w", err)
	}

	// Create Captain
	capt, err := captain.NewCaptain("main-captain", captainConfig, llmProvider, logger)
	if err != nil {
		return fmt.Errorf("failed to create captain: %w", err)
	}
	defer capt.Shutdown(nil)

	// Create execution plan
	ctx := context.Background()
	plan, err := capt.PlanGoal(ctx, p.Goal)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// Display the plan
	fmt.Printf("\n=== EXECUTION PLAN ===\n")
	fmt.Printf("Goal: %s\n", plan.Goal)
	fmt.Printf("Plan ID: %s\n", plan.ID)
	fmt.Printf("Created: %s\n", plan.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Estimated Duration: %s\n", plan.EstimatedDuration)
	
	if plan.Reasoning != "" {
		fmt.Printf("\n=== REASONING ===\n%s\n", plan.Reasoning)
	}

	fmt.Printf("\n=== TASKS (%d) ===\n", len(plan.Tasks))
	for i, task := range plan.Tasks {
		fmt.Printf("\n%d. %s (ID: %s)\n", i+1, task.Title, task.ID)
		fmt.Printf("   Description: %s\n", task.Description)
		if task.Command != "" {
			fmt.Printf("   Command: %s\n", task.Command)
		}
		if len(task.Dependencies) > 0 {
			fmt.Printf("   Dependencies: %v\n", task.Dependencies)
		}
		fmt.Printf("   Duration: %s\n", task.EstimatedDuration)
		fmt.Printf("   Priority: %d\n", task.Priority)
	}

	// Validate the plan
	if err := capt.ValidatePlan(ctx, plan); err != nil {
		fmt.Printf("\n⚠️  Plan validation warnings: %s\n", err.Error())
	} else {
		fmt.Printf("\n✅ Plan validation: PASSED\n")
	}

	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Total tasks: %d\n", len(plan.Tasks))
	fmt.Printf("Estimated time: %s\n", plan.EstimatedDuration)
	fmt.Printf("\nTo execute this plan, use: capn execute \"%s\"\n", p.Goal)
	fmt.Printf("To execute in dry-run mode, use: capn execute --dry-run \"%s\"\n", p.Goal)

	return nil
}

// ExecuteCmd represents the execute command (with optional planning mode)
type ExecuteCmd struct {
	PlanOnly bool   `help:"Plan only, don't execute" short:"n" name:"plan-only"`
	Goal     string `arg:"" help:"Goal to execute"`
}

func (e *ExecuteCmd) Run(globals *GlobalOptions, logger *zap.Logger) error {
	// Check if we're in planning mode (plan-only or global dry-run)
	planningMode := e.PlanOnly || globals.DryRun
	
	if planningMode {
		logger.Info("Creating plan", zap.String("goal", e.Goal))
		
		// Load configuration
		var cfg *config.Config
		var err error
		
		if globals.Config != "" {
			cfg, err = config.LoadConfig(globals.Config)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
		} else {
			cfg = config.NewConfig()
		}

		// Check if OpenAI API key is configured
		if cfg.Captain.OpenAIAPIKey == "" {
			fmt.Printf("Planning: %s (Demo mode - OpenAI not configured)\n", e.Goal)
			fmt.Println("Would create detailed execution plan using LLM-powered analysis...")
			return nil
		}

		// Create captain configuration
		captainConfig := &captain.Config{
			OpenAIAPIKey:        cfg.Captain.OpenAIAPIKey,
			Model:               cfg.Captain.Model,
			MaxTokens:           cfg.Captain.MaxTokens,
			Temperature:         cfg.Captain.Temperature,
			MaxConcurrentAgents: cfg.Captain.MaxConcurrentAgents,
			PlanningTimeout:     cfg.Captain.PlanningTimeout,
			RetryAttempts:       cfg.Captain.RetryAttempts,
			RetryDelay:          cfg.Captain.RetryDelay,
		}

		// Create OpenAI provider
		llmProvider, err := captain.NewOpenAIProvider(captainConfig)
		if err != nil {
			return fmt.Errorf("failed to create OpenAI provider: %w", err)
		}

		// Create Captain
		capt, err := captain.NewCaptain("main-captain", captainConfig, llmProvider, logger)
		if err != nil {
			return fmt.Errorf("failed to create captain: %w", err)
		}
		defer capt.Shutdown(context.Background())

		// Create and execute plan in dry-run mode
		ctx := context.Background()
		plan, err := capt.PlanGoal(ctx, e.Goal)
		if err != nil {
			return fmt.Errorf("failed to create plan: %w", err)
		}

		fmt.Printf("=== DRY RUN EXECUTION PLAN ===\n")
		fmt.Printf("Goal: %s\n", plan.Goal)
		
		results, err := capt.ExecutePlan(ctx, plan, true)
		if err != nil {
			return fmt.Errorf("failed to execute dry run: %w", err)
		}

		fmt.Printf("\n=== DRY RUN RESULTS ===\n") 
		for _, result := range results {
			status := "✅"
			if !result.Success {
				status = "❌"
			}
			fmt.Printf("%s %s\n", status, result.Output)
		}
		
	} else {
		logger.Info("Executing", zap.String("goal", e.Goal))
		fmt.Printf("Executing: %s\n", e.Goal)
		// TODO: Implement actual execution logic with crew agents
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

	Plan    PlanCmd    `cmd:"" help:"Analyze goal and create intelligent execution plan"`
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
