package captain

import (
	"time"
)

// TaskType represents the type of task to be executed
type TaskType string

const (
	TaskTypeAnalysis   TaskType = "analysis"
	TaskTypeExecution  TaskType = "execution"
	TaskTypeValidation TaskType = "validation"
	TaskTypeReporting  TaskType = "reporting"
)

// Priority represents task priority levels
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// StrategyType represents execution strategy types
type StrategyType string

const (
	StrategySequential StrategyType = "sequential"
	StrategyParallel   StrategyType = "parallel"
	StrategyHybrid     StrategyType = "hybrid"
)

// Task represents a single executable task
type Task struct {
	ID           string            `json:"id"`
	Type         TaskType          `json:"type"`
	Priority     Priority          `json:"priority"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Payload      map[string]any    `json:"payload,omitempty"`
	Deadline     time.Time         `json:"deadline,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ExecutionTimeline represents the timeline for plan execution
type ExecutionTimeline struct {
	EstimatedDuration time.Duration `json:"estimated_duration"`
	StartTime         time.Time     `json:"start_time,omitempty"`
	EndTime           time.Time     `json:"end_time,omitempty"`
}

// ResourceAllocation represents resource requirements for plan execution
type ResourceAllocation struct {
	MaxAgents     int      `json:"max_agents"`
	RequiredTools []string `json:"required_tools,omitempty"`
	EstimatedCost float64  `json:"estimated_cost,omitempty"`
}

// ExecutionStrategy represents the strategy for executing the plan
type ExecutionStrategy struct {
	Type        StrategyType   `json:"type"`
	Description string         `json:"description,omitempty"`
	Options     map[string]any `json:"options,omitempty"`
}

// ExecutionPlan represents a complete execution plan
type ExecutionPlan struct {
	ID        string             `json:"id"`
	Goal      string             `json:"goal"`
	Tasks     []Task             `json:"tasks"`
	Timeline  ExecutionTimeline  `json:"timeline"`
	Resources ResourceAllocation `json:"resources"`
	Strategy  ExecutionStrategy  `json:"strategy"`
}

// Result represents the result of a task execution
type Result struct {
	TaskID    string         `json:"task_id"`
	Success   bool           `json:"success"`
	Output    string         `json:"output,omitempty"`
	Error     string         `json:"error,omitempty"`
	Duration  time.Duration  `json:"duration"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}