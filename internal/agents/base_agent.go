package agents

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BaseAgent provides a base implementation of the Agent interface
type BaseAgent struct {
	id        string
	name      string
	agentType AgentType
	
	mu              sync.RWMutex
	status          AgentStatus
	receivedMsgs    []Message
	router          *MessageRouter
	startTime       time.Time
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(id, name string, agentType AgentType) *BaseAgent {
	return &BaseAgent{
		id:           id,
		name:         name,
		agentType:    agentType,
		status:       AgentStatusIdle,
		receivedMsgs: make([]Message, 0),
		startTime:    time.Now(),
	}
}

// ID returns the agent's unique identifier
func (b *BaseAgent) ID() string {
	return b.id
}

// Name returns the agent's human-readable name
func (b *BaseAgent) Name() string {
	return b.name
}

// Type returns the agent's type
func (b *BaseAgent) Type() AgentType {
	return b.agentType
}

// Status returns the current status of the agent
func (b *BaseAgent) Status() AgentStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

// Health returns the health status of the agent
func (b *BaseAgent) Health() HealthStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	var healthState HealthState
	var message string
	
	switch b.status {
	case AgentStatusIdle, AgentStatusBusy:
		healthState = HealthStatusHealthy
		message = "Agent is operational"
	case AgentStatusError:
		healthState = HealthStatusUnhealthy
		message = "Agent has encountered an error"
	case AgentStatusStopped:
		healthState = HealthStatusDegraded
		message = "Agent has been stopped"
	default:
		healthState = HealthStatusUnhealthy
		message = "Agent status is unknown"
	}
	
	return HealthStatus{
		Status:    healthState,
		Message:   message,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"uptime": time.Since(b.startTime),
			"status": b.status,
		},
	}
}

// Execute executes a task (base implementation - can be overridden)
func (b *BaseAgent) Execute(ctx context.Context, task Task) Result {
	b.mu.Lock()
	oldStatus := b.status
	b.status = AgentStatusBusy
	b.mu.Unlock()
	
	defer func() {
		b.mu.Lock()
		b.status = oldStatus
		b.mu.Unlock()
	}()
	
	// Base implementation just acknowledges the task
	return Result{
		TaskID:    task.ID,
		Success:   true,
		Output:    fmt.Sprintf("Task %s executed by %s", task.ID, b.name),
		Duration:  time.Millisecond * 10, // Simulate quick execution
		Timestamp: time.Now(),
	}
}

// SendMessage sends a message to another agent via the router
func (b *BaseAgent) SendMessage(to string, message Message) error {
	b.mu.RLock()
	router := b.router
	b.mu.RUnlock()
	
	if router == nil {
		return fmt.Errorf("no router configured for agent %s", b.id)
	}
	
	// Ensure the message is properly formatted
	message.From = b.id
	message.To = to
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	
	return router.RouteMessage(message)
}

// ReceiveMessage receives a message from another agent
func (b *BaseAgent) ReceiveMessage(message Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	// Store the message for later processing/retrieval
	b.receivedMsgs = append(b.receivedMsgs, message)
	
	return nil
}

// Stop gracefully stops the agent
func (b *BaseAgent) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.status != AgentStatusStopped {
		b.status = AgentStatusStopped
	}
	
	return nil
}

// SetRouter sets the message router for this agent
func (b *BaseAgent) SetRouter(router *MessageRouter) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.router = router
}

// GetReceivedMessages returns a copy of all received messages (for testing/debugging)
func (b *BaseAgent) GetReceivedMessages() []Message {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	messages := make([]Message, len(b.receivedMsgs))
	copy(messages, b.receivedMsgs)
	return messages
}

// SetStatus sets the agent status (for internal use)
func (b *BaseAgent) SetStatus(status AgentStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = status
}

// ClearReceivedMessages clears all received messages (for testing/debugging)
func (b *BaseAgent) ClearReceivedMessages() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.receivedMsgs = make([]Message, 0)
}