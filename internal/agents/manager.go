package agents

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/iainlowe/capn/internal/common"
)

// AgentStats represents statistics about managed agents
type AgentStats struct {
	Total   int               `json:"total"`
	Idle    int               `json:"idle"`
	Busy    int               `json:"busy"`
	Stopped int               `json:"stopped"`
	Error   int               `json:"error"`
	ByType  map[AgentType]int `json:"by_type"`
}

// AgentManager handles the lifecycle of agents
type AgentManager struct {
	mu       sync.RWMutex
	agents   map[string]Agent
	router   *MessageRouter
	registry *AgentRegistry
}

// NewAgentManager creates a new agent manager
func NewAgentManager() *AgentManager {
	registry := NewAgentRegistry()

	// Register default agent factories
	registry.Register(AgentTypeFile, func(id, name string) (Agent, error) {
		return NewBaseAgent(id, name, AgentTypeFile), nil
	})
	registry.Register(AgentTypeNetwork, func(id, name string) (Agent, error) {
		return NewBaseAgent(id, name, AgentTypeNetwork), nil
	})
	registry.Register(AgentTypeResearch, func(id, name string) (Agent, error) {
		return NewBaseAgent(id, name, AgentTypeResearch), nil
	})
	registry.Register(AgentTypeCaptain, func(id, name string) (Agent, error) {
		return NewBaseAgent(id, name, AgentTypeCaptain), nil
	})

	return &AgentManager{
		agents:   make(map[string]Agent),
		registry: registry,
	}
}

// SetRouter sets the message router for the manager
func (m *AgentManager) SetRouter(router *MessageRouter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.router = router
}

// SpawnAgent creates and starts a new agent
func (m *AgentManager) SpawnAgent(id, name string, agentType AgentType) (Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if agent with this ID already exists
	if _, exists := m.agents[id]; exists {
		return nil, fmt.Errorf("agent with ID %s already exists", id)
	}

	// Create the appropriate agent type using registry
	agent, err := m.registry.CreateAgent(id, name, agentType)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Set router if available and agent supports it
	if m.router != nil {
		if baseAgent, ok := agent.(*BaseAgent); ok {
			baseAgent.SetRouter(m.router)
		}

		// Register with router
		if err := m.router.RegisterAgent(agent); err != nil {
			return nil, fmt.Errorf("failed to register agent with router: %w", err)
		}
	}

	// Store in manager
	m.agents[id] = agent

	return agent, nil
}

// TerminateAgent stops and removes an agent
func (m *AgentManager) TerminateAgent(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}

	// Stop the agent
	if err := agent.Stop(); err != nil {
		return fmt.Errorf("failed to stop agent %s: %w", agentID, err)
	}

	// Unregister from router if available
	if m.router != nil {
		if err := m.router.UnregisterAgent(agentID); err != nil {
			// Log but don't fail - agent might not be registered
			// In a full implementation, we would use a logger here
		}
	}

	// Remove from manager
	delete(m.agents, agentID)

	return nil
}

// TerminateAll stops and removes all managed agents
func (m *AgentManager) TerminateAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error

	// Stop all agents
	for agentID, agent := range m.agents {
		if err := agent.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop agent %s: %w", agentID, err))
		}

		// Unregister from router
		if m.router != nil {
			if err := m.router.UnregisterAgent(agentID); err != nil {
				// Log but don't add to errors - agent might not be registered
			}
		}
	}

	// Clear all agents
	m.agents = make(map[string]Agent)

	if len(errors) > 0 {
		return fmt.Errorf("errors during terminate all: %v", errors)
	}

	return nil
}

// GetManagedAgents returns a copy of all managed agents
func (m *AgentManager) GetManagedAgents() []Agent {
	return common.CollectMapValues(&m.mu, m.agents)
}

// GetAgent returns a specific managed agent by ID
func (m *AgentManager) GetAgent(agentID string) (Agent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.agents[agentID]
	return agent, exists
}

// MonitorAgents periodically checks the health of all managed agents
func (m *AgentManager) MonitorAgents(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkAgentHealth()
		}
	}
}

// checkAgentHealth checks the health of all agents and handles unhealthy ones
func (m *AgentManager) checkAgentHealth() {
	m.mu.RLock()
	agents := make([]Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	m.mu.RUnlock()

	for _, agent := range agents {
		health := agent.Health()

		// In a full implementation, we might take action based on health status
		// For now, just check that the agent is responding
		if health.Status == HealthStatusUnhealthy {
			// Could implement recovery logic here
			// For now, just continue monitoring
		}
	}
}

// GetAgentStats returns statistics about managed agents
func (m *AgentManager) GetAgentStats() AgentStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := AgentStats{
		ByType: make(map[AgentType]int),
	}

	for _, agent := range m.agents {
		stats.Total++

		// Count by status
		switch agent.Status() {
		case AgentStatusIdle:
			stats.Idle++
		case AgentStatusBusy:
			stats.Busy++
		case AgentStatusStopped:
			stats.Stopped++
		case AgentStatusError:
			stats.Error++
		}

		// Count by type
		agentType := agent.Type()
		stats.ByType[agentType]++
	}

	return stats
}
