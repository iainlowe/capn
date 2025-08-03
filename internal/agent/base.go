package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MessageRouter interface for routing messages between agents
type MessageRouter interface {
	RouteMessage(message Message) error
}

// BaseAgent provides a basic implementation of the Agent interface
type BaseAgent struct {
	id            string
	name          string
	agentType     AgentType
	status        AgentStatus
	health        HealthStatus
	router        MessageRouter
	messages      []Message
	mutex         sync.RWMutex
	stopChan      chan struct{}
	stopped       bool
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(id, name string, agentType AgentType) *BaseAgent {
	return &BaseAgent{
		id:        id,
		name:      name,
		agentType: agentType,
		status:    Idle,
		health:    Healthy,
		messages:  make([]Message, 0),
		stopChan:  make(chan struct{}),
		stopped:   false,
	}
}

// ID returns the agent's unique identifier
func (a *BaseAgent) ID() string {
	return a.id
}

// Name returns the agent's name
func (a *BaseAgent) Name() string {
	return a.name
}

// Type returns the agent's type
func (a *BaseAgent) Type() AgentType {
	return a.agentType
}

// Status returns the agent's current status
func (a *BaseAgent) Status() AgentStatus {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.status
}

// Health returns the agent's health status
func (a *BaseAgent) Health() HealthStatus {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.health
}

// SetStatus sets the agent's status
func (a *BaseAgent) SetStatus(status AgentStatus) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.status = status
}

// SetHealth sets the agent's health status
func (a *BaseAgent) SetHealth(health HealthStatus) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.health = health
}

// SetRouter sets the message router for this agent
func (a *BaseAgent) SetRouter(router MessageRouter) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.router = router
}

// Execute executes a task - base implementation returns not implemented error
func (a *BaseAgent) Execute(ctx context.Context, task Task) Result {
	return Result{
		TaskID:  task.ID,
		Success: false,
		Error:   fmt.Sprintf("Execute method not implemented for %s", a.agentType.String()),
	}
}

// SendMessage sends a message to another agent via the router
func (a *BaseAgent) SendMessage(to string, message Message) error {
	a.mutex.RLock()
	router := a.router
	a.mutex.RUnlock()
	
	if router == nil {
		return fmt.Errorf("no message router configured for agent %s", a.id)
	}
	
	// Set timestamp if not already set
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	
	// Ensure From and To are set correctly
	message.From = a.id
	message.To = to
	
	return router.RouteMessage(message)
}

// ReceiveMessage receives a message from another agent
func (a *BaseAgent) ReceiveMessage(message Message) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	// Check if agent is stopped
	if a.stopped {
		return fmt.Errorf("agent %s is stopped and cannot receive messages", a.id)
	}
	
	// Store the message
	a.messages = append(a.messages, message)
	
	return nil
}

// GetReceivedMessages returns all received messages (for testing/debugging)
func (a *BaseAgent) GetReceivedMessages() []Message {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	
	// Return a copy to avoid race conditions
	messages := make([]Message, len(a.messages))
	copy(messages, a.messages)
	return messages
}

// Stop stops the agent
func (a *BaseAgent) Stop() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	if a.stopped {
		return nil // Already stopped
	}
	
	a.stopped = true
	a.status = Stopped
	close(a.stopChan)
	
	return nil
}

// IsRunning returns true if the agent is running
func (a *BaseAgent) IsRunning() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return !a.stopped
}