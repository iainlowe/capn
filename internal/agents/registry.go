package agents

import (
	"fmt"
	"sync"
)

// AgentCreator is a function that creates an agent of a specific type
type AgentCreator func(id, name string) (Agent, error)

// AgentRegistry manages agent creators by type
type AgentRegistry struct {
	mu       sync.RWMutex
	creators map[AgentType]AgentCreator
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		creators: make(map[AgentType]AgentCreator),
	}
}

// Register registers an agent creator for a specific type
func (r *AgentRegistry) Register(agentType AgentType, creator AgentCreator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.creators[agentType] = creator
}

// CreateAgent creates an agent of the specified type
func (r *AgentRegistry) CreateAgent(id, name string, agentType AgentType) (Agent, error) {
	r.mu.RLock()
	creator, exists := r.creators[agentType]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}

	return creator(id, name)
}

// GetSupportedTypes returns all supported agent types
func (r *AgentRegistry) GetSupportedTypes() []AgentType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]AgentType, 0, len(r.creators))
	for t := range r.creators {
		types = append(types, t)
	}
	return types
}

// RegisteredAgentFactory uses the registry to create agents
type RegisteredAgentFactory struct {
	registry *AgentRegistry
}

// NewRegisteredAgentFactory creates a factory using the provided registry
func NewRegisteredAgentFactory(registry *AgentRegistry) *RegisteredAgentFactory {
	return &RegisteredAgentFactory{
		registry: registry,
	}
}

// CreateAgent creates an agent using the registry
func (f *RegisteredAgentFactory) CreateAgent(id, name string, agentType AgentType) (Agent, error) {
	return f.registry.CreateAgent(id, name, agentType)
}

// DefaultAgentCreators provides default creators for basic agent types
func DefaultAgentCreators() map[AgentType]AgentCreator {
	return map[AgentType]AgentCreator{
		AgentTypeCaptain: func(id, name string) (Agent, error) {
			return NewBaseAgent(id, name, AgentTypeCaptain), nil
		},
		AgentTypeFile: func(id, name string) (Agent, error) {
			return NewBaseAgent(id, name, AgentTypeFile), nil
		},
		AgentTypeNetwork: func(id, name string) (Agent, error) {
			return NewBaseAgent(id, name, AgentTypeNetwork), nil
		},
		AgentTypeResearch: func(id, name string) (Agent, error) {
			return NewBaseAgent(id, name, AgentTypeResearch), nil
		},
	}
}

// NewDefaultAgentRegistry creates a registry with default agent creators
func NewDefaultAgentRegistry() *AgentRegistry {
	registry := NewAgentRegistry()

	for agentType, creator := range DefaultAgentCreators() {
		registry.Register(agentType, creator)
	}

	return registry
}
