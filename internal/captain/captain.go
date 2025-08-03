package captain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iainlowe/capn/internal/config"
)

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusBusy    AgentStatus = "busy" 
	AgentStatusStopped AgentStatus = "stopped"
	AgentStatusError   AgentStatus = "error"
)

// CaptainStatus represents the current status of the Captain
type CaptainStatus struct {
	ID          string      `json:"id"`
	Status      AgentStatus `json:"status"`
	ActiveTasks int         `json:"active_tasks"`
	QueuedTasks int         `json:"queued_tasks"`
	Uptime      time.Duration `json:"uptime"`
}

// ExecutionResult represents the result of executing a plan
type ExecutionResult struct {
	PlanID      string   `json:"plan_id"`
	Success     bool     `json:"success"`
	DryRun      bool     `json:"dry_run"`
	TaskResults []Result `json:"task_results"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	Error       string   `json:"error,omitempty"`
}

// Captain is the main orchestrator agent that uses LLM for planning
type Captain struct {
	id          string
	config      *config.Config
	llmProvider LLMProvider
	planner     *PlanningEngine
	taskQueue   chan Task
	resultChan  chan Result
	
	// State management
	mu         sync.RWMutex
	status     AgentStatus
	activeTasks map[string]bool
	startTime  time.Time
	
	// Shutdown management
	ctx        context.Context
	cancel     context.CancelFunc
	stopped    chan struct{}
}

// NewCaptain creates a new Captain agent
func NewCaptain(id string, config *config.Config, openaiConfig OpenAIConfig) (*Captain, error) {
	if id == "" {
		return nil, fmt.Errorf("captain ID cannot be empty")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create OpenAI provider
	llmProvider, err := NewOpenAIProvider(openaiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
	}

	// Create planning engine
	planner := NewPlanningEngine(llmProvider)

	ctx, cancel := context.WithCancel(context.Background())

	captain := &Captain{
		id:          id,
		config:      config,
		llmProvider: llmProvider,
		planner:     planner,
		taskQueue:   make(chan Task, 1000), // Buffered channel for tasks
		resultChan:  make(chan Result, 1000), // Buffered channel for results
		
		status:      AgentStatusIdle,
		activeTasks: make(map[string]bool),
		startTime:   time.Now(),
		
		ctx:     ctx,
		cancel:  cancel,
		stopped: make(chan struct{}),
	}

	return captain, nil
}

// ID returns the Captain's ID
func (c *Captain) ID() string {
	return c.id
}

// CreatePlan creates an execution plan from a goal using LLM reasoning
func (c *Captain) CreatePlan(ctx context.Context, goal string) (*ExecutionPlan, error) {
	if goal == "" {
		return nil, fmt.Errorf("goal cannot be empty")
	}

	// Use the planning engine to create the plan
	plan, err := c.planner.CreatePlan(ctx, goal)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	return plan, nil
}

// ExecutePlan executes an execution plan, optionally in dry-run mode
func (c *Captain) ExecutePlan(ctx context.Context, plan *ExecutionPlan, dryRun bool) (*ExecutionResult, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan cannot be nil")
	}

	// Validate the plan first
	if err := c.planner.ValidatePlan(plan); err != nil {
		return nil, fmt.Errorf("invalid plan: %w", err)
	}

	startTime := time.Now()
	result := &ExecutionResult{
		PlanID:      plan.ID,
		Success:     true,
		DryRun:      dryRun,
		TaskResults: make([]Result, len(plan.Tasks)),
		StartTime:   startTime,
	}

	// Update status
	c.mu.Lock()
	c.status = AgentStatusBusy
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.status = AgentStatusIdle
		c.mu.Unlock()
		
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
	}()

	// Execute tasks (in dry-run mode, just simulate)
	for i, task := range plan.Tasks {
		taskResult := Result{
			TaskID:    task.ID,
			Success:   true,
			Timestamp: time.Now(),
		}

		if dryRun {
			// Simulate task execution in dry-run mode
			taskResult.Output = fmt.Sprintf("DRY RUN: Would execute task %s of type %s with priority %s", 
				task.ID, task.Type, task.Priority)
			taskResult.Duration = time.Millisecond * 100 // Simulate quick execution
		} else {
			// TODO: Implement actual task execution with crew agents
			taskResult.Output = fmt.Sprintf("Task %s executed successfully", task.ID)
			taskResult.Duration = time.Second * 5 // Simulate longer execution
		}

		result.TaskResults[i] = taskResult
	}

	return result, nil
}

// Status returns the current status of the Captain
func (c *Captain) Status() CaptainStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CaptainStatus{
		ID:          c.id,
		Status:      c.status,
		ActiveTasks: len(c.activeTasks),
		QueuedTasks: len(c.taskQueue),
		Uptime:      time.Since(c.startTime),
	}
}

// Stop gracefully stops the Captain
func (c *Captain) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status == AgentStatusStopped {
		return nil // Already stopped
	}

	// Cancel context to signal shutdown
	c.cancel()
	c.status = AgentStatusStopped

	// Close channels
	close(c.taskQueue)
	close(c.resultChan)
	close(c.stopped)

	return nil
}