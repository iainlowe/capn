package agents

import (
	"context"
	"fmt"
	"time"
)

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeCaptain  AgentType = "captain"
	AgentTypeFile     AgentType = "file"
	AgentTypeNetwork  AgentType = "network"
	AgentTypeResearch AgentType = "research"
)

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusBusy    AgentStatus = "busy"
	AgentStatusStopped AgentStatus = "stopped"
	AgentStatusError   AgentStatus = "error"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeText    MessageType = "text"
	MessageTypeCommand MessageType = "command"
	MessageTypeResult  MessageType = "result"
	MessageTypeStatus  MessageType = "status"
)

// Priority represents task priority levels
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// HealthState represents the health state of an agent
type HealthState string

const (
	HealthStatusHealthy   HealthState = "healthy"
	HealthStatusDegraded  HealthState = "degraded"
	HealthStatusUnhealthy HealthState = "unhealthy"
)

// Message represents a message between agents
type Message struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Content   string                 `json:"content"`
	Type      MessageType            `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Validate validates the message
func (m *Message) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("message ID cannot be empty")
	}
	if m.From == "" {
		return fmt.Errorf("message from cannot be empty")
	}
	if m.To == "" {
		return fmt.Errorf("message to cannot be empty")
	}
	if m.Content == "" {
		return fmt.Errorf("message content cannot be empty")
	}
	return nil
}

// Task represents a task to be executed by an agent
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Priority    Priority               `json:"priority"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Deadline    time.Time              `json:"deadline,omitempty"`
}

// Validate validates the task
func (t *Task) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if t.Type == "" {
		return fmt.Errorf("task type cannot be empty")
	}
	if t.Description == "" {
		return fmt.Errorf("task description cannot be empty")
	}
	return nil
}

// Result represents the result of task execution
type Result struct {
	TaskID    string                 `json:"task_id"`
	Success   bool                   `json:"success"`
	Output    string                 `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// HealthStatus represents the health status of an agent
type HealthStatus struct {
	Status    HealthState `json:"status"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// MessageLog represents a logged message with metadata
type MessageLog struct {
	Message   Message   `json:"message"`
	Formatted string    `json:"formatted"`
	Indexed   time.Time `json:"indexed"`
}

// Agent interface defines the contract for all agents in the system
type Agent interface {
	// Identity methods
	ID() string
	Name() string
	Type() AgentType
	
	// Status methods
	Status() AgentStatus
	Health() HealthStatus
	
	// Task execution
	Execute(ctx context.Context, task Task) Result
	
	// Communication methods
	SendMessage(to string, message Message) error
	ReceiveMessage(message Message) error
	
	// Lifecycle methods
	Stop() error
}

// CommunicationLogger interface defines the contract for logging agent communications
type CommunicationLogger interface {
	LogMessage(from, to string, message Message)
	GetHistory(agentID string) []MessageLog
	SearchMessages(query string) []MessageLog
	GetAllMessages() []MessageLog
	Clear() error
}