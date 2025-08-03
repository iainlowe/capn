package captain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Captain is the main orchestrator that uses LLM reasoning to analyze goals and create sophisticated execution plans
type Captain struct {
	id           string
	config       *Config
	llmProvider  LLMProvider
	planner      PlanningEngine
	taskQueue    chan Task
	resultChan   chan Result
	logger       *zap.Logger
	mu           sync.RWMutex
	activeAgents map[string]bool
	shutdown     chan struct{}
}

// NewCaptain creates a new Captain instance
func NewCaptain(id string, config *Config, llmProvider LLMProvider, logger *zap.Logger) (*Captain, error) {
	if id == "" {
		return nil, fmt.Errorf("captain ID cannot be empty")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if llmProvider == nil {
		return nil, fmt.Errorf("LLM provider cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	planner := NewLLMPlanningEngine(llmProvider, config)
	
	return &Captain{
		id:           id,
		config:       config,
		llmProvider:  llmProvider,
		planner:      planner,
		taskQueue:    make(chan Task, config.MaxConcurrentAgents*2),
		resultChan:   make(chan Result, config.MaxConcurrentAgents*2),
		logger:       logger,
		activeAgents: make(map[string]bool),
		shutdown:     make(chan struct{}),
	}, nil
}

// PlanGoal analyzes a goal and creates an execution plan using LLM-powered reasoning
func (c *Captain) PlanGoal(ctx context.Context, goal string) (*ExecutionPlan, error) {
	c.logger.Info("Starting goal analysis", zap.String("goal", goal))
	
	start := time.Now()
	defer func() {
		c.logger.Info("Goal analysis completed", 
			zap.Duration("duration", time.Since(start)))
	}()

	// Use the planning engine to analyze the goal
	plan, err := c.planner.AnalyzeGoal(ctx, goal)
	if err != nil {
		c.logger.Error("Failed to analyze goal", 
			zap.String("goal", goal), 
			zap.Error(err))
		return nil, fmt.Errorf("goal analysis failed: %w", err)
	}

	c.logger.Info("Plan created successfully",
		zap.String("planID", plan.ID),
		zap.Int("taskCount", len(plan.Tasks)),
		zap.Duration("estimatedDuration", plan.EstimatedDuration))

	return plan, nil
}

// ValidatePlan validates an execution plan for feasibility and correctness
func (c *Captain) ValidatePlan(ctx context.Context, plan *ExecutionPlan) error {
	c.logger.Info("Validating execution plan", zap.String("planID", plan.ID))
	
	start := time.Now()
	defer func() {
		c.logger.Info("Plan validation completed", 
			zap.Duration("duration", time.Since(start)))
	}()

	if err := c.planner.ValidatePlan(ctx, plan); err != nil {
		c.logger.Error("Plan validation failed", 
			zap.String("planID", plan.ID), 
			zap.Error(err))
		return fmt.Errorf("plan validation failed: %w", err)
	}

	c.logger.Info("Plan validation successful", zap.String("planID", plan.ID))
	return nil
}

// OptimizePlan optimizes an execution plan for better performance
func (c *Captain) OptimizePlan(ctx context.Context, plan *ExecutionPlan) (*ExecutionPlan, error) {
	c.logger.Info("Optimizing execution plan", zap.String("planID", plan.ID))
	
	start := time.Now()
	defer func() {
		c.logger.Info("Plan optimization completed", 
			zap.Duration("duration", time.Since(start)))
	}()

	optimizedPlan, err := c.planner.OptimizePlan(ctx, plan)
	if err != nil {
		c.logger.Error("Plan optimization failed", 
			zap.String("planID", plan.ID), 
			zap.Error(err))
		return nil, fmt.Errorf("plan optimization failed: %w", err)
	}

	c.logger.Info("Plan optimization successful",
		zap.String("originalPlanID", plan.ID),
		zap.String("optimizedPlanID", optimizedPlan.ID),
		zap.Duration("originalDuration", plan.EstimatedDuration),
		zap.Duration("optimizedDuration", optimizedPlan.EstimatedDuration))

	return optimizedPlan, nil
}

// ExecutePlan executes an execution plan by spawning crew agents
func (c *Captain) ExecutePlan(ctx context.Context, plan *ExecutionPlan, dryRun bool) ([]Result, error) {
	if dryRun {
		return c.dryRunPlan(plan)
	}

	c.logger.Info("Starting plan execution", 
		zap.String("planID", plan.ID),
		zap.Int("taskCount", len(plan.Tasks)))

	// Validate plan before execution
	if err := c.ValidatePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("cannot execute invalid plan: %w", err)
	}

	// TODO: Implement actual execution logic with crew agents
	// For now, return mock results to complete the interface
	results := make([]Result, len(plan.Tasks))
	for i, task := range plan.Tasks {
		results[i] = Result{
			TaskID:    task.ID,
			Success:   true,
			Output:    fmt.Sprintf("Task %s would be executed: %s", task.ID, task.Command),
			Duration:  task.EstimatedDuration,
			Timestamp: time.Now().UTC(),
		}
	}

	c.logger.Info("Plan execution completed", 
		zap.String("planID", plan.ID),
		zap.Int("results", len(results)))

	return results, nil
}

// dryRunPlan performs a dry run of the execution plan
func (c *Captain) dryRunPlan(plan *ExecutionPlan) ([]Result, error) {
	c.logger.Info("Performing dry run", zap.String("planID", plan.ID))

	results := make([]Result, len(plan.Tasks))
	for i, task := range plan.Tasks {
		results[i] = Result{
			TaskID:    task.ID,
			Success:   true,
			Output:    fmt.Sprintf("DRY RUN: Would execute %s - %s", task.Title, task.Command),
			Duration:  0, // No actual execution time
			Timestamp: time.Now().UTC(),
		}
	}

	return results, nil
}

// GetActiveAgentCount returns the number of currently active agents
func (c *Captain) GetActiveAgentCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.activeAgents)
}

// GetMaxConcurrentAgents returns the maximum number of concurrent agents
func (c *Captain) GetMaxConcurrentAgents() int {
	return c.config.MaxConcurrentAgents
}

// GetID returns the captain's ID
func (c *Captain) GetID() string {
	return c.id
}

// GetConfig returns the captain's configuration
func (c *Captain) GetConfig() *Config {
	return c.config
}

// Shutdown gracefully shuts down the captain
func (c *Captain) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down captain", zap.String("id", c.id))
	
	select {
	case <-c.shutdown:
		// Already shut down
		return nil
	default:
		close(c.shutdown)
	}

	// Wait for active agents to complete or context to cancel
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Warn("Shutdown timed out, forcing exit", zap.String("id", c.id))
			return ctx.Err()
		case <-ticker.C:
			if c.GetActiveAgentCount() == 0 {
				c.logger.Info("Captain shutdown completed", zap.String("id", c.id))
				return nil
			}
		}
	}
}

// IsShutdown returns true if the captain is shut down
func (c *Captain) IsShutdown() bool {
	select {
	case <-c.shutdown:
		return true
	default:
		return false
	}
}