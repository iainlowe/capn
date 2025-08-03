package agents

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentManager_SpawnAgent(t *testing.T) {
	manager := NewAgentManager()
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	manager.SetRouter(router)
	
	// Spawn a file agent
	agent, err := manager.SpawnAgent("file-1", "FileAgent-1", AgentTypeFile)
	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "file-1", agent.ID())
	assert.Equal(t, "FileAgent-1", agent.Name())
	assert.Equal(t, AgentTypeFile, agent.Type())
	
	// Agent should be registered with router
	_, exists := router.GetAgent("file-1")
	assert.True(t, exists)
	
	// Try to spawn agent with same ID
	_, err = manager.SpawnAgent("file-1", "FileAgent-2", AgentTypeFile)
	assert.Error(t, err)
}

func TestAgentManager_TerminateAgent(t *testing.T) {
	manager := NewAgentManager()
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	manager.SetRouter(router)
	
	// Spawn and then terminate an agent
	agent, err := manager.SpawnAgent("file-1", "FileAgent-1", AgentTypeFile)
	require.NoError(t, err)
	
	err = manager.TerminateAgent("file-1")
	assert.NoError(t, err)
	
	// Agent should be stopped
	assert.Equal(t, AgentStatusStopped, agent.Status())
	
	// Agent should be unregistered from router
	_, exists := router.GetAgent("file-1")
	assert.False(t, exists)
	
	// Try to terminate non-existent agent
	err = manager.TerminateAgent("non-existent")
	assert.Error(t, err)
}

func TestAgentManager_GetManagedAgents(t *testing.T) {
	manager := NewAgentManager()
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	manager.SetRouter(router)
	
	// Spawn multiple agents
	_, err := manager.SpawnAgent("file-1", "FileAgent-1", AgentTypeFile)
	require.NoError(t, err)
	_, err = manager.SpawnAgent("net-1", "NetworkAgent-1", AgentTypeNetwork)
	require.NoError(t, err)
	_, err = manager.SpawnAgent("research-1", "ResearchAgent-1", AgentTypeResearch)
	require.NoError(t, err)
	
	managedAgents := manager.GetManagedAgents()
	assert.Len(t, managedAgents, 3)
	
	// Check that all agent types are represented
	agentTypes := make(map[AgentType]bool)
	for _, agent := range managedAgents {
		agentTypes[agent.Type()] = true
	}
	assert.True(t, agentTypes[AgentTypeFile])
	assert.True(t, agentTypes[AgentTypeNetwork])
	assert.True(t, agentTypes[AgentTypeResearch])
}

func TestAgentManager_MonitorAgents(t *testing.T) {
	manager := NewAgentManager()
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	manager.SetRouter(router)
	
	// Spawn an agent
	agent, err := manager.SpawnAgent("file-1", "FileAgent-1", AgentTypeFile)
	require.NoError(t, err)
	
	// Start monitoring in background
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	
	go manager.MonitorAgents(ctx, 100*time.Millisecond)
	
	// Let monitoring run for a bit
	time.Sleep(200 * time.Millisecond)
	
	// Agent should still be healthy
	assert.Equal(t, AgentStatusIdle, agent.Status())
	health := agent.Health()
	assert.Equal(t, HealthStatusHealthy, health.Status)
}

func TestAgentManager_TerminateAll(t *testing.T) {
	manager := NewAgentManager()
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	manager.SetRouter(router)
	
	// Spawn multiple agents
	_, err := manager.SpawnAgent("file-1", "FileAgent-1", AgentTypeFile)
	require.NoError(t, err)
	_, err = manager.SpawnAgent("net-1", "NetworkAgent-1", AgentTypeNetwork)
	require.NoError(t, err)
	
	// Verify agents exist
	managedAgents := manager.GetManagedAgents()
	assert.Len(t, managedAgents, 2)
	
	// Terminate all
	err = manager.TerminateAll()
	assert.NoError(t, err)
	
	// All agents should be stopped and removed
	managedAgents = manager.GetManagedAgents()
	assert.Len(t, managedAgents, 0)
	
	// Router should have no agents
	registeredAgents := router.GetRegisteredAgents()
	assert.Len(t, registeredAgents, 0)
}

func TestAgentManager_GetAgentStats(t *testing.T) {
	manager := NewAgentManager()
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	manager.SetRouter(router)
	
	// Spawn agents of different types
	_, err := manager.SpawnAgent("file-1", "FileAgent-1", AgentTypeFile)
	require.NoError(t, err)
	_, err = manager.SpawnAgent("file-2", "FileAgent-2", AgentTypeFile)
	require.NoError(t, err)
	_, err = manager.SpawnAgent("net-1", "NetworkAgent-1", AgentTypeNetwork)
	require.NoError(t, err)
	
	stats := manager.GetAgentStats()
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 3, stats.Idle)
	assert.Equal(t, 0, stats.Busy)
	assert.Equal(t, 0, stats.Stopped)
	assert.Equal(t, 0, stats.Error)
	assert.Contains(t, stats.ByType, AgentTypeFile)
	assert.Contains(t, stats.ByType, AgentTypeNetwork)
	assert.Equal(t, 2, stats.ByType[AgentTypeFile])
	assert.Equal(t, 1, stats.ByType[AgentTypeNetwork])
}