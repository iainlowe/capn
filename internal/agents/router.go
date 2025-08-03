package agents

import (
	"fmt"
	"sync"
)

// MessageRouter handles routing messages between agents and logging communications
type MessageRouter struct {
	mu      sync.RWMutex
	agents  map[string]Agent
	logger  CommunicationLogger
}

// NewMessageRouter creates a new message router
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		agents: make(map[string]Agent),
	}
}

// SetLogger sets the communication logger for this router
func (r *MessageRouter) SetLogger(logger CommunicationLogger) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger = logger
}

// RegisterAgent registers an agent with the router
func (r *MessageRouter) RegisterAgent(agent Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	id := agent.ID()
	if _, exists := r.agents[id]; exists {
		return fmt.Errorf("agent with ID %s is already registered", id)
	}
	
	r.agents[id] = agent
	return nil
}

// UnregisterAgent removes an agent from the router
func (r *MessageRouter) UnregisterAgent(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent with ID %s is not registered", agentID)
	}
	
	delete(r.agents, agentID)
	return nil
}

// RouteMessage routes a message to the specified recipient
func (r *MessageRouter) RouteMessage(message Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Validate the message
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}
	
	// Find the recipient agent
	recipient, exists := r.agents[message.To]
	if !exists {
		return fmt.Errorf("recipient agent not found: %s", message.To)
	}
	
	// Deliver the message
	if err := recipient.ReceiveMessage(message); err != nil {
		return fmt.Errorf("failed to deliver message to %s: %w", message.To, err)
	}
	
	// Log the message if logger is available
	if r.logger != nil {
		r.logger.LogMessage(message.From, message.To, message)
	}
	
	return nil
}

// BroadcastMessage sends a message to all registered agents except the sender
func (r *MessageRouter) BroadcastMessage(message Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Validate the message
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}
	
	var deliveryErrors []error
	
	// Send to all agents except the sender
	for agentID, agent := range r.agents {
		if agentID == message.From {
			continue // Don't send to sender
		}
		
		// Create a copy of the message with the specific recipient
		msgCopy := message
		msgCopy.To = agentID
		
		// Deliver the message
		if err := agent.ReceiveMessage(msgCopy); err != nil {
			deliveryErrors = append(deliveryErrors, fmt.Errorf("failed to deliver to %s: %w", agentID, err))
			continue
		}
		
		// Log the message if logger is available
		if r.logger != nil {
			r.logger.LogMessage(message.From, agentID, msgCopy)
		}
	}
	
	// Return error if any deliveries failed
	if len(deliveryErrors) > 0 {
		return fmt.Errorf("broadcast delivery failures: %v", deliveryErrors)
	}
	
	return nil
}

// GetRegisteredAgents returns a copy of all registered agents
func (r *MessageRouter) GetRegisteredAgents() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	agents := make([]Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	
	return agents
}

// GetAgent returns a specific agent by ID
func (r *MessageRouter) GetAgent(agentID string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	agent, exists := r.agents[agentID]
	return agent, exists
}