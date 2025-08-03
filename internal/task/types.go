package task

import (
	"strings"
	"time"
)

// TaskStatus represents the current state of a task
type TaskStatus int

const (
	TaskStatusQueued TaskStatus = iota
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCancelled
)

// String returns the string representation of TaskStatus
func (ts TaskStatus) String() string {
	switch ts {
	case TaskStatusQueued:
		return "queued"
	case TaskStatusRunning:
		return "running"
	case TaskStatusCompleted:
		return "completed"
	case TaskStatusFailed:
		return "failed"
	case TaskStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// IsActive returns true if the task is still active (queued or running)
func (ts TaskStatus) IsActive() bool {
	return ts == TaskStatusQueued || ts == TaskStatusRunning
}

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of LogLevel
func (ll LogLevel) String() string {
	switch ll {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	default:
		return "unknown"
	}
}

// AgentType represents the type of agent involved in task execution
type AgentType string

const (
	AgentTypeCaptain AgentType = "captain"
	AgentTypeCrew    AgentType = "crew"
	AgentTypeMCP     AgentType = "mcp"
)

// TaskSummary provides a brief overview of a task
type TaskSummary struct {
	ID       string     `json:"id"`
	Status   TaskStatus `json:"status"`
	Goal     string     `json:"goal"`
	Started  time.Time  `json:"started"`
	Finished *time.Time `json:"finished,omitempty"`
}

// TaskDetails provides comprehensive information about a task
type TaskDetails struct {
	TaskSummary
	Plan         string                 `json:"plan"`
	Progress     float64                `json:"progress"`
	CurrentStep  int                    `json:"current_step"`
	TotalSteps   int                    `json:"total_steps"`
	ActiveAgents []string               `json:"active_agents"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LogEntry represents a single log entry from task execution
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Agent     string                 `json:"agent"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DateRange represents a time range for filtering
type DateRange struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

// Contains checks if a time falls within the date range
func (dr DateRange) Contains(t time.Time) bool {
	if dr.Start != nil && t.Before(*dr.Start) {
		return false
	}
	if dr.End != nil && t.After(*dr.End) {
		return false
	}
	return true
}

// TaskFilter defines criteria for filtering tasks
type TaskFilter struct {
	Status    []TaskStatus `json:"status,omitempty"`
	DateRange DateRange    `json:"date_range,omitempty"`
	Keywords  []string     `json:"keywords,omitempty"`
	AgentType []AgentType  `json:"agent_type,omitempty"`
}

// Matches checks if a task summary matches the filter criteria
func (tf TaskFilter) Matches(summary *TaskSummary) bool {
	// Check status filter
	if len(tf.Status) > 0 {
		found := false
		for _, status := range tf.Status {
			if summary.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check date range filter
	if !tf.DateRange.Contains(summary.Started) {
		return false
	}

	// Check keywords filter
	if len(tf.Keywords) > 0 {
		goalLower := strings.ToLower(summary.Goal)
		found := false
		for _, keyword := range tf.Keywords {
			if strings.Contains(goalLower, strings.ToLower(keyword)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// TODO: Implement agent type filtering when we have agent information in TaskSummary

	return true
}