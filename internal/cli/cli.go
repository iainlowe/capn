package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/iainlowe/capn/internal/config"
	"github.com/iainlowe/capn/internal/task"
)

var (
	globalTaskManager task.TaskManager
	taskManagerMu     sync.RWMutex
	globalCLI         *CLI
	cliMu             sync.RWMutex
)

// setGlobalTaskManager sets the global task manager instance
func setGlobalTaskManager(tm task.TaskManager) {
	taskManagerMu.Lock()
	defer taskManagerMu.Unlock()
	globalTaskManager = tm
}

// getGlobalTaskManager gets the global task manager instance
func getGlobalTaskManager() task.TaskManager {
	taskManagerMu.RLock()
	defer taskManagerMu.RUnlock()
	return globalTaskManager
}

// setCurrentCLI sets the current CLI instance for command access
func setCurrentCLI(cli *CLI) {
	cliMu.Lock()
	defer cliMu.Unlock()
	globalCLI = cli
}

// getCurrentCLI gets the current CLI instance
func getCurrentCLI() *CLI {
	cliMu.RLock()
	defer cliMu.RUnlock()
	return globalCLI
}

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

func (e *ExecuteCmd) Run(globals *GlobalOptions, logger *zap.Logger) error {
	// Check if we're in planning mode (plan-only or global dry-run)
	planningMode := e.PlanOnly || globals.DryRun
	
	// Get current CLI instance for output
	cli := getCurrentCLI()
	
	if planningMode {
		logger.Info("Creating plan", zap.String("goal", e.Goal))
		fmt.Fprintf(cli.getOutput(), "Planning: %s\n", e.Goal)
		// TODO: Implement planning logic
	} else {
		logger.Info("Starting task", zap.String("goal", e.Goal))
		
		// Get task manager from global singleton (we'll implement this)
		taskManager := getGlobalTaskManager()
		if taskManager == nil {
			// Fallback to old behavior for compatibility
			logger.Info("Task manager not available, falling back to old behavior")
			fmt.Fprintf(cli.getOutput(), "Executing: %s\n", e.Goal)
			return nil
		}
		
		// Start fire-and-forget task execution
		ctx := context.Background()
		taskExec, err := taskManager.StartTask(ctx, e.Goal)
		if err != nil {
			return fmt.Errorf("failed to start task: %w", err)
		}
		
		fmt.Fprintf(cli.getOutput(), "Task started: %s\n", taskExec.ID)
		fmt.Fprintf(cli.getOutput(), "Captain has begun planning...\n")
	}
	return nil
}

// StatusCmd represents the status command
type StatusCmd struct{}

func (s *StatusCmd) Run(globals *GlobalOptions, logger *zap.Logger) error {
	logger.Info("Checking status")
	
	// Get current CLI instance for output
	cli := getCurrentCLI()
	
	// Get task manager from global singleton
	taskManager := getGlobalTaskManager()
	if taskManager == nil {
		// Fallback to old behavior
		fmt.Fprintf(cli.getOutput(), "No active tasks\n")
		return nil
	}
	
	// List all active tasks
	tasks, err := taskManager.ListTasks(task.TaskFilter{})
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}
	
	if len(tasks) == 0 {
		fmt.Fprintf(cli.getOutput(), "No active tasks\n")
		return nil
	}
	
	fmt.Fprintf(cli.getOutput(), "Active tasks (%d):\n", len(tasks))
	for _, t := range tasks {
		fmt.Fprintf(cli.getOutput(), "  %s: %s [%s]\n", t.ID, t.Goal, t.Status)
		if t.Plan != nil && len(t.Plan.Steps) > 0 {
			fmt.Fprintf(cli.getOutput(), "    Steps: %d/%d completed\n", len(t.Results), len(t.Plan.Steps))
		}
	}
	
	return nil
}

// AgentsCmd represents the agents command
type AgentsCmd struct{}

func (a *AgentsCmd) Run(globals *GlobalOptions, logger *zap.Logger) error {
	logger.Info("Managing agents")
	cli := getCurrentCLI()
	fmt.Fprintf(cli.getOutput(), "Managing agents\n")
	// TODO: Implement agents management logic
	return nil
}

// MCPCmd represents the mcp command
type MCPCmd struct{}

func (m *MCPCmd) Run(globals *GlobalOptions, logger *zap.Logger) error {
	logger.Info("Managing MCP servers")
	cli := getCurrentCLI()
	fmt.Fprintf(cli.getOutput(), "Managing MCP servers\n")
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
	taskManager  task.TaskManager
	callback     func(*GlobalOptions)
	exitOverride bool
	skipConfig   bool // Skip config loading for tests
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		output:       os.Stdout,
		exitOverride: true, // Default to true for tests
		taskManager:  task.NewManager(),
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

// getOutput returns the output writer for the CLI
func (c *CLI) getOutput() io.Writer {
	if c.output == nil {
		return os.Stdout
	}
	return c.output
}

// SetSkipConfigForTests disables config file loading for testing
func (c *CLI) SetSkipConfigForTests(skip bool) {
	c.skipConfig = skip
}

// Parse runs the CLI with the given arguments
func (c *CLI) Parse(args []string) error {
	// Initialize logger first (needed for binding)
	c.logger = c.createLogger()
	
	// Initialize task manager if not already done
	if c.taskManager == nil {
		c.taskManager = task.NewManager()
	}
	
	// Set global task manager for command access
	setGlobalTaskManager(c.taskManager)
	
	// Set global CLI for command access
	setCurrentCLI(c)
	
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
	err = ctx.Run()
	
	// Cleanup task manager if needed
	if c.taskManager != nil && !c.skipConfig {
		// Only close for real execution, not tests
		defer c.taskManager.Close()
	}
	
	return err
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
