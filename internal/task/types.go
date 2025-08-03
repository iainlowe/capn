package task

import (
	"context"
	"time"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	StatusQueued    TaskStatus = "queued"
	StatusPlanning  TaskStatus = "planning"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusCancelled TaskStatus = "cancelled"
)

// TaskResult represents the result of a task execution step
type TaskResult struct {
	Step        string    `json:"step"`
	Status      string    `json:"status"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
	CompletedAt time.Time `json:"completed_at"`
}

// CommunicationLog represents a message in the task execution log
type CommunicationLog struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source,omitempty"`
}

// ExecutionPlan represents the plan for executing a task
type ExecutionPlan struct {
	Steps       []string  `json:"steps"`
	CreatedAt   time.Time `json:"created_at"`
	EstimatedDuration time.Duration `json:"estimated_duration,omitempty"`
}

// TaskExecution represents a running or completed task
type TaskExecution struct {
	ID        string             `json:"id"`
	Goal      string             `json:"goal"`
	Status    TaskStatus         `json:"status"`
	StartTime time.Time          `json:"start_time"`
	EndTime   *time.Time         `json:"end_time,omitempty"`
	Plan      *ExecutionPlan     `json:"plan,omitempty"`
	Results   []TaskResult       `json:"results"`
	Messages  []CommunicationLog `json:"messages"`
	Error     string             `json:"error,omitempty"`
	
	// Internal fields for execution management
	ctx    context.Context
	cancel context.CancelFunc
}

// TaskFilter represents filtering options for listing tasks
type TaskFilter struct {
	Status    *TaskStatus `json:"status,omitempty"`
	Limit     int         `json:"limit,omitempty"`
	Offset    int         `json:"offset,omitempty"`
	SinceTime *time.Time  `json:"since_time,omitempty"`
}

// TaskManager interface defines the contract for task management
type TaskManager interface {
	StartTask(ctx context.Context, goal string) (*TaskExecution, error)
	GetTask(taskID string) (*TaskExecution, error)
	ListTasks(filter TaskFilter) ([]*TaskExecution, error)
	CancelTask(taskID string) error
	Close() error
}