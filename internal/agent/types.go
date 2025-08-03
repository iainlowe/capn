package agent

import (
	"context"
	"errors"
	"time"
)

// AgentType represents the type of agent
type AgentType int

const (
	Captain AgentType = iota
	FileAgent
	NetworkAgent
	ResearchAgent
)

// String returns the string representation of AgentType
func (at AgentType) String() string {
	switch at {
	case Captain:
		return "Captain"
	case FileAgent:
		return "FileAgent"
	case NetworkAgent:
		return "NetworkAgent"
	case ResearchAgent:
		return "ResearchAgent"
	default:
		return "Unknown"
	}
}

// AgentStatus represents the current status of an agent
type AgentStatus int

const (
	Idle AgentStatus = iota
	Running
	Stopped
	Error
)

// String returns the string representation of AgentStatus
func (as AgentStatus) String() string {
	switch as {
	case Idle:
		return "Idle"
	case Running:
		return "Running"
	case Stopped:
		return "Stopped"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}

// HealthStatus represents the health status of an agent
type HealthStatus int

const (
	Healthy HealthStatus = iota
	Unhealthy
	Unknown
)

// String returns the string representation of HealthStatus
func (hs HealthStatus) String() string {
	switch hs {
	case Healthy:
		return "Healthy"
	case Unhealthy:
		return "Unhealthy"
	case Unknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// MessageType represents the type of message
type MessageType int

const (
	TaskMessage MessageType = iota
	ResponseMessage
	StatusMessage
	ErrorMessage
)

// Message represents a communication message between agents
type Message struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Content   string                 `json:"content"`
	Type      MessageType            `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Validate validates the message
func (m *Message) Validate() error {
	if m.ID == "" {
		return errors.New("Message ID cannot be empty")
	}
	if m.From == "" {
		return errors.New("Message From cannot be empty")
	}
	if m.To == "" {
		return errors.New("Message To cannot be empty")
	}
	return nil
}

// Task represents a task to be executed by an agent
type Task struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Validate validates the task
func (t *Task) Validate() error {
	if t.ID == "" {
		return errors.New("Task ID cannot be empty")
	}
	if t.Description == "" {
		return errors.New("Task Description cannot be empty")
	}
	return nil
}

// Result represents the result of a task execution
type Result struct {
	TaskID  string      `json:"task_id"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// IsSuccess returns true if the result indicates success
func (r *Result) IsSuccess() bool {
	return r.Success
}

// IsError returns true if the result indicates an error
func (r *Result) IsError() bool {
	return !r.Success
}

// GetError returns the error if the result indicates failure
func (r *Result) GetError() error {
	if !r.Success && r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

// Agent interface defines the contract for all agent types
type Agent interface {
	// Identity methods
	ID() string
	Name() string
	Type() AgentType
	Status() AgentStatus
	
	// Execution methods
	Execute(ctx context.Context, task Task) Result
	
	// Communication methods
	SendMessage(to string, message Message) error
	ReceiveMessage(message Message) error
	
	// Lifecycle methods
	Stop() error
	Health() HealthStatus
}