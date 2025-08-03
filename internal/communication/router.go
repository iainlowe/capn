package communication

import (
	"fmt"
	"sync"
	"time"

	"github.com/iainlowe/capn/internal/agent"
)

// MessageRouter handles routing messages between agents
type MessageRouter struct {
	agents map[string]agent.Agent
	logger CommunicationLogger
	mutex  sync.RWMutex
}

// NewMessageRouter creates a new message router
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		agents: make(map[string]agent.Agent),
		logger: NewInMemoryLogger(), // Default to in-memory logger
	}
}

// SetLogger sets the communication logger
func (r *MessageRouter) SetLogger(logger CommunicationLogger) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.logger = logger
}

// RegisterAgent registers an agent with the router
func (r *MessageRouter) RegisterAgent(a agent.Agent) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.agents[a.ID()]; exists {
		return fmt.Errorf("agent with ID %s is already registered", a.ID())
	}
	
	r.agents[a.ID()] = a
	return nil
}

// UnregisterAgent removes an agent from the router
func (r *MessageRouter) UnregisterAgent(agentID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}
	
	delete(r.agents, agentID)
	return nil
}

// GetRegisteredAgents returns a list of registered agent IDs
func (r *MessageRouter) GetRegisteredAgents() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	agents := make([]string, 0, len(r.agents))
	for id := range r.agents {
		agents = append(agents, id)
	}
	return agents
}

// RouteMessage routes a message to the specified recipient
func (r *MessageRouter) RouteMessage(message agent.Message) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	// Validate message
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}
	
	// Set timestamp if not set
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	
	// Find recipient agent
	recipient, exists := r.agents[message.To]
	if !exists {
		return fmt.Errorf("recipient agent %s not found", message.To)
	}
	
	// Log the message
	if r.logger != nil {
		r.logger.LogMessage(message.From, message.To, message)
	}
	
	// Deliver the message
	if err := recipient.ReceiveMessage(message); err != nil {
		return fmt.Errorf("failed to deliver message to %s: %w", message.To, err)
	}
	
	return nil
}

// BroadcastMessage sends a message to all registered agents except those in the exclude list
func (r *MessageRouter) BroadcastMessage(message agent.Message, excludeIDs []string) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	// Create exclude map for quick lookup
	excludeMap := make(map[string]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}
	
	// Set timestamp if not set
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	
	var errors []error
	
	// Send to all agents except excluded ones
	for agentID, a := range r.agents {
		if excludeMap[agentID] {
			continue
		}
		
		// Create a copy of the message with the correct recipient
		msgCopy := message
		msgCopy.To = agentID
		
		// Log the message
		if r.logger != nil {
			r.logger.LogMessage(message.From, agentID, msgCopy)
		}
		
		// Deliver the message
		if err := a.ReceiveMessage(msgCopy); err != nil {
			errors = append(errors, fmt.Errorf("failed to deliver to %s: %w", agentID, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed with %d errors: %v", len(errors), errors)
	}
	
	return nil
}

// GetAgent returns an agent by ID
func (r *MessageRouter) GetAgent(agentID string) (agent.Agent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	a, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}
	
	return a, nil
}

// GetLogger returns the current communication logger
func (r *MessageRouter) GetLogger() CommunicationLogger {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.logger
}