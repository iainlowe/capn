package task

import (
	"context"
	"time"
)

// TaskQuery defines the interface for querying tasks
type TaskQuery interface {
	// ListTasks returns a list of task summaries matching the filter
	ListTasks(ctx context.Context, filter TaskFilter) ([]*TaskSummary, error)
	
	// GetTaskDetails returns detailed information about a specific task
	GetTaskDetails(ctx context.Context, taskID string) (*TaskDetails, error)
	
	// GetTaskLogs returns the log entries for a specific task
	GetTaskLogs(ctx context.Context, taskID string) ([]LogEntry, error)
	
	// SearchTasks searches for tasks matching the query string
	SearchTasks(ctx context.Context, query string) ([]*TaskSummary, error)
}

// TaskStorage extends TaskQuery with storage operations
type TaskStorage interface {
	TaskQuery
	
	// StoreTask stores a task execution
	StoreTask(ctx context.Context, task *TaskExecution) error
	
	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, task *TaskExecution) error
	
	// AddLogEntry adds a log entry to a task
	AddLogEntry(ctx context.Context, taskID string, entry LogEntry) error
}

// NotificationService defines the interface for task notifications
type NotificationService interface {
	// NotifyCompletion notifies when a task completes successfully
	NotifyCompletion(task *TaskExecution) error
	
	// NotifyError notifies when a task encounters an error
	NotifyError(task *TaskExecution, err error) error
	
	// ConfigureNotifications configures notification preferences
	ConfigureNotifications(prefs NotificationPreferences) error
}

// TaskExecution represents a complete task execution with all details
type TaskExecution struct {
	ID           string                 `json:"id"`
	Status       TaskStatus             `json:"status"`
	Goal         string                 `json:"goal"`
	Started      time.Time              `json:"started"`
	Finished     *time.Time             `json:"finished,omitempty"`
	Plan         string                 `json:"plan,omitempty"`
	Progress     float64                `json:"progress"`
	CurrentStep  int                    `json:"current_step"`
	TotalSteps   int                    `json:"total_steps"`
	ActiveAgents []string               `json:"active_agents,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Error        *string                `json:"error,omitempty"`
}

// ToSummary converts a TaskExecution to a TaskSummary
func (te *TaskExecution) ToSummary() *TaskSummary {
	return &TaskSummary{
		ID:       te.ID,
		Status:   te.Status,
		Goal:     te.Goal,
		Started:  te.Started,
		Finished: te.Finished,
	}
}

// ToDetails converts a TaskExecution to TaskDetails
func (te *TaskExecution) ToDetails() *TaskDetails {
	return &TaskDetails{
		TaskSummary:  *te.ToSummary(),
		Plan:         te.Plan,
		Progress:     te.Progress,
		CurrentStep:  te.CurrentStep,
		TotalSteps:   te.TotalSteps,
		ActiveAgents: te.ActiveAgents,
		Metadata:     te.Metadata,
	}
}

// NotificationPreferences holds user preferences for notifications
type NotificationPreferences struct {
	EnableCompletion bool   `json:"enable_completion"`
	EnableErrors     bool   `json:"enable_errors"`
	OutputFormat     string `json:"output_format"` // "console", "system", etc.
}