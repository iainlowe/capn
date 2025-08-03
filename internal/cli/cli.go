package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/iainlowe/capn/internal/config"
	"github.com/iainlowe/capn/internal/task"
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

func (e *ExecuteCmd) Run(globals *GlobalOptions, logger *zap.Logger, cli *CLI) error {
	// Check if we're in planning mode (plan-only or global dry-run)
	planningMode := e.PlanOnly || globals.DryRun
	
	if planningMode {
		logger.Info("Creating plan", zap.String("goal", e.Goal))
		fmt.Fprintf(cli.output, "Planning: %s\n", e.Goal)
		// TODO: Implement planning logic
	} else {
		logger.Info("Executing", zap.String("goal", e.Goal))
		fmt.Fprintf(cli.output, "Executing: %s\n", e.Goal)
		// TODO: Implement execution logic
	}
	return nil
}

// StatusCmd represents the status command
type StatusCmd struct{}

func (s *StatusCmd) Run(globals *GlobalOptions, logger *zap.Logger, cli *CLI) error {
	logger.Info("Checking status")
	
	storage := cli.GetTaskStorage()
	if storage == nil {
		fmt.Println("No task storage configured")
		return nil
	}
	
	ctx := context.Background()
	filter := task.TaskFilter{} // Empty filter to get all tasks
	
	tasks, err := storage.ListTasks(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}
	
	// Count tasks by status
	statusCounts := make(map[task.TaskStatus]int)
	for _, task := range tasks {
		statusCounts[task.Status]++
	}
	
	total := len(tasks)
	running := statusCounts[task.TaskStatusRunning]
	completed := statusCounts[task.TaskStatusCompleted]
	queued := statusCounts[task.TaskStatusQueued]
	failed := statusCounts[task.TaskStatusFailed]
	
	fmt.Fprintf(cli.output, "%d tasks total: %d running, %d completed, %d queued", 
		total, running, completed, queued)
	
	if failed > 0 {
		fmt.Fprintf(cli.output, ", %d failed", failed)
	}
	fmt.Fprintln(cli.output)
	
	// Show current activity
	if running > 0 {
		for _, t := range tasks {
			if t.Status == task.TaskStatusRunning {
				fmt.Fprintf(cli.output, "Current activity: %s...\n", t.Goal)
				break
			}
		}
	}
	
	return nil
}

// TasksCmd represents the tasks command group
type TasksCmd struct {
	List TasksListCmd `cmd:"" help:"List all tasks with optional filtering"`
	Show TasksShowCmd `cmd:"" help:"Show detailed task information"`
	Logs TasksLogsCmd `cmd:"" help:"Show task execution logs and agent communications"`
}

// TasksListCmd represents the tasks list command
type TasksListCmd struct {
	Status   []string `help:"Filter by status (queued,running,completed,failed,cancelled)" short:"s"`
	Keywords []string `help:"Filter by keywords in goal" short:"k"`
	Limit    int      `help:"Limit number of results" default:"50"`
}

func (t *TasksListCmd) Run(globals *GlobalOptions, logger *zap.Logger, cli *CLI) error {
	logger.Info("Listing tasks", zap.Strings("status", t.Status), zap.Strings("keywords", t.Keywords))
	
	storage := cli.GetTaskStorage()
	if storage == nil {
		fmt.Println("No task storage configured")
		return nil
	}
	
	ctx := context.Background()
	filter := task.TaskFilter{
		Keywords: t.Keywords,
	}
	
	// Parse status filters
	if len(t.Status) > 0 {
		var statusFilters []task.TaskStatus
		for _, s := range t.Status {
			switch strings.ToLower(s) {
			case "queued":
				statusFilters = append(statusFilters, task.TaskStatusQueued)
			case "running":
				statusFilters = append(statusFilters, task.TaskStatusRunning)
			case "completed":
				statusFilters = append(statusFilters, task.TaskStatusCompleted)
			case "failed":
				statusFilters = append(statusFilters, task.TaskStatusFailed)
			case "cancelled":
				statusFilters = append(statusFilters, task.TaskStatusCancelled)
			default:
				return fmt.Errorf("invalid status: %s", s)
			}
		}
		filter.Status = statusFilters
	}
	
	tasks, err := storage.ListTasks(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}
	
	if len(tasks) == 0 {
		fmt.Fprintln(cli.output, "No tasks found")
		return nil
	}
	
	// Limit results
	if t.Limit > 0 && len(tasks) > t.Limit {
		tasks = tasks[:t.Limit]
	}
	
	// Display header
	fmt.Fprintf(cli.output, "%-20s %-12s %-40s %s\n", "ID", "Status", "Goal", "Started")
	fmt.Fprintln(cli.output, strings.Repeat("-", 80))
	
	// Display tasks
	for _, task := range tasks {
		elapsed := time.Since(task.Started)
		var timeStr string
		if elapsed.Hours() > 24 {
			timeStr = fmt.Sprintf("%.0fd ago", elapsed.Hours()/24)
		} else if elapsed.Hours() > 1 {
			timeStr = fmt.Sprintf("%.0fh ago", elapsed.Hours())
		} else if elapsed.Minutes() > 1 {
			timeStr = fmt.Sprintf("%.0fm ago", elapsed.Minutes())
		} else {
			timeStr = fmt.Sprintf("%.0fs ago", elapsed.Seconds())
		}
		
		goal := task.Goal
		if len(goal) > 40 {
			goal = goal[:37] + "..."
		}
		
		fmt.Fprintf(cli.output, "%-20s %-12s %-40s %s\n", task.ID, task.Status, goal, timeStr)
	}
	
	return nil
}

// TasksShowCmd represents the tasks show command
type TasksShowCmd struct {
	TaskID string `arg:"" help:"Task ID to show details for"`
}

func (t *TasksShowCmd) Run(globals *GlobalOptions, logger *zap.Logger, cli *CLI) error {
	logger.Info("Showing task details", zap.String("task_id", t.TaskID))
	
	storage := cli.GetTaskStorage()
	if storage == nil {
		fmt.Println("No task storage configured")
		return nil
	}
	
	ctx := context.Background()
	details, err := storage.GetTaskDetails(ctx, t.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task details: %w", err)
	}
	
	fmt.Fprintf(cli.output, "Task: %s\n", details.ID)
	fmt.Fprintf(cli.output, "Goal: %s\n", details.Goal)
	fmt.Fprintf(cli.output, "Status: %s", details.Status)
	
	if details.Progress > 0 {
		fmt.Fprintf(cli.output, " (%.0f%% complete)", details.Progress)
	}
	fmt.Fprintln(cli.output)
	
	fmt.Fprintf(cli.output, "Started: %s\n", details.Started.Format("2006-01-02 15:04:05"))
	if details.Finished != nil {
		fmt.Fprintf(cli.output, "Finished: %s\n", details.Finished.Format("2006-01-02 15:04:05"))
		duration := details.Finished.Sub(details.Started)
		fmt.Fprintf(cli.output, "Duration: %s\n", duration.Round(time.Second))
	}
	
	if details.Plan != "" {
		fmt.Fprintf(cli.output, "Plan: %s\n", details.Plan)
	}
	
	if details.TotalSteps > 0 {
		fmt.Fprintf(cli.output, "Progress: %d/%d steps\n", details.CurrentStep, details.TotalSteps)
	}
	
	if len(details.ActiveAgents) > 0 {
		fmt.Fprintf(cli.output, "Active Agents: %s\n", strings.Join(details.ActiveAgents, ", "))
	}
	
	return nil
}

// TasksLogsCmd represents the tasks logs command
type TasksLogsCmd struct {
	TaskID string `arg:"" help:"Task ID to show logs for"`
	Level  string `help:"Filter by log level (debug,info,warn,error)" short:"l"`
	Tail   int    `help:"Show last N log entries" short:"n" default:"100"`
}

func (t *TasksLogsCmd) Run(globals *GlobalOptions, logger *zap.Logger, cli *CLI) error {
	logger.Info("Showing task logs", zap.String("task_id", t.TaskID))
	
	storage := cli.GetTaskStorage()
	if storage == nil {
		fmt.Println("No task storage configured")
		return nil
	}
	
	ctx := context.Background()
	logs, err := storage.GetTaskLogs(ctx, t.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task logs: %w", err)
	}
	
	if len(logs) == 0 {
		fmt.Fprintln(cli.output, "No logs found for task")
		return nil
	}
	
	// Filter by level if specified
	if t.Level != "" {
		var levelFilter task.LogLevel
		switch strings.ToLower(t.Level) {
		case "debug":
			levelFilter = task.LogLevelDebug
		case "info":
			levelFilter = task.LogLevelInfo
		case "warn":
			levelFilter = task.LogLevelWarn
		case "error":
			levelFilter = task.LogLevelError
		default:
			return fmt.Errorf("invalid log level: %s", t.Level)
		}
		
		var filteredLogs []task.LogEntry
		for _, log := range logs {
			if log.Level == levelFilter {
				filteredLogs = append(filteredLogs, log)
			}
		}
		logs = filteredLogs
	}
	
	// Apply tail limit
	if t.Tail > 0 && len(logs) > t.Tail {
		logs = logs[len(logs)-t.Tail:]
	}
	
	// Display logs
	for _, log := range logs {
		timestamp := log.Timestamp.Format("15:04:05")
		level := strings.ToUpper(log.Level.String())
		fmt.Fprintf(cli.output, "[%s] %s [%s] %s\n", timestamp, level, log.Agent, log.Message)
		
		// Display metadata if present
		if len(log.Metadata) > 0 {
			for k, v := range log.Metadata {
				fmt.Fprintf(cli.output, "  %s: %v\n", k, v)
			}
		}
	}
	
	return nil
}

// AgentsCmd represents the agents command
type AgentsCmd struct{}

func (a *AgentsCmd) Run(globals *GlobalOptions, logger *zap.Logger, cli *CLI) error {
	logger.Info("Managing agents")
	fmt.Fprintln(cli.output, "Managing agents")
	// TODO: Implement agents management logic
	return nil
}

// MCPCmd represents the mcp command
type MCPCmd struct{}

func (m *MCPCmd) Run(globals *GlobalOptions, logger *zap.Logger, cli *CLI) error {
	logger.Info("Managing MCP servers")
	fmt.Fprintln(cli.output, "Managing MCP servers")
	// TODO: Implement MCP server management logic
	return nil
}
// CLI represents the main CLI structure
type CLI struct {
	GlobalOptions

	Execute ExecuteCmd `cmd:"" help:"Plan and execute goals (use --dry-run for planning only)"`
	Status  StatusCmd  `cmd:"" help:"Show current operation status"`
	Tasks   TasksCmd   `cmd:"" help:"Manage and query tasks"`
	Agents  AgentsCmd  `cmd:"" help:"Manage agent configurations"`
	MCP     MCPCmd     `cmd:"" help:"Manage MCP server connections"`

	output       io.Writer
	logger       *zap.Logger
	config       *config.Config
	taskStorage  task.TaskStorage
	callback     func(*GlobalOptions)
	exitOverride bool
	skipConfig   bool // Skip config loading for tests
}

// GetTaskStorage returns the task storage instance
func (c *CLI) GetTaskStorage() task.TaskStorage {
	return c.taskStorage
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		output:       os.Stdout,
		exitOverride: true, // Default to true for tests
		taskStorage:  task.NewInMemoryStorage(), // Initialize with in-memory storage
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
		kong.Bind(c),                // Bind CLI instance for access to storage
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
