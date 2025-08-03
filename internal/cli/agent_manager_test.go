package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentManager_NewAgentManager(t *testing.T) {
	am := NewAgentManager()
	assert.NotNil(t, am)
	assert.NotNil(t, am.router)
	assert.NotNil(t, am.logger)
}

func TestAgentManager_ListAgents_Empty(t *testing.T) {
	am := NewAgentManager()
	
	result := am.ListAgents()
	assert.Contains(t, result, "No agents currently registered")
}

func TestAgentManager_CreateExampleAgents(t *testing.T) {
	am := NewAgentManager()
	
	err := am.CreateExampleAgents()
	require.NoError(t, err)
	
	// Test listing agents
	result := am.ListAgents()
	assert.Contains(t, result, "Registered Agents (4)")
	assert.Contains(t, result, "Captain")
	assert.Contains(t, result, "FileAgent-1")
	assert.Contains(t, result, "NetworkAgent-1")
	assert.Contains(t, result, "ResearchAgent-1")
	assert.Contains(t, result, "Type: Captain")
	assert.Contains(t, result, "Status: Idle")
	assert.Contains(t, result, "Health: Healthy")
}

func TestAgentManager_SimulateConversation(t *testing.T) {
	am := NewAgentManager()
	
	// Create agents first
	err := am.CreateExampleAgents()
	require.NoError(t, err)
	
	// Simulate conversation
	err = am.SimulateConversation()
	require.NoError(t, err)
	
	// Check history
	history := am.GetCommunicationHistory(0) // No limit
	assert.Contains(t, history, "Communication History")
	assert.Contains(t, history, "captain-001 -> file-001")
	assert.Contains(t, history, "analyze the Go files")
	assert.Contains(t, history, "Found 23 Go files")
	assert.Contains(t, history, "Hey Research")
	assert.Contains(t, history, "Found 5 relevant patterns")
}

func TestAgentManager_GetCommunicationHistory_Empty(t *testing.T) {
	am := NewAgentManager()
	
	result := am.GetCommunicationHistory(10)
	assert.Contains(t, result, "No communication history available")
}

func TestAgentManager_GetCommunicationHistory_WithLimit(t *testing.T) {
	am := NewAgentManager()
	
	// Create agents and simulate conversation
	err := am.CreateExampleAgents()
	require.NoError(t, err)
	err = am.SimulateConversation()
	require.NoError(t, err)
	
	// Test with limit
	result := am.GetCommunicationHistory(2)
	assert.Contains(t, result, "Communication History")
	assert.Contains(t, result, "(Showing last 2 messages)")
	
	// Count lines with timestamps (should be 2)
	lines := strings.Split(result, "\n")
	timestampLines := 0
	for _, line := range lines {
		if strings.Contains(line, "] ") && strings.Contains(line, " -> ") {
			timestampLines++
		}
	}
	assert.Equal(t, 2, timestampLines)
}

func TestAgentManager_SearchMessages(t *testing.T) {
	am := NewAgentManager()
	
	// Test search with no messages
	result := am.SearchMessages("test")
	assert.Contains(t, result, "No messages found containing 'test'")
	
	// Create agents and simulate conversation
	err := am.CreateExampleAgents()
	require.NoError(t, err)
	err = am.SimulateConversation()
	require.NoError(t, err)
	
	// Test search with results
	result = am.SearchMessages("Go files")
	assert.Contains(t, result, "Messages containing 'Go files'")
	assert.Contains(t, result, "2 results")
	assert.Contains(t, result, "captain-001 -> file-001")
	assert.Contains(t, result, "file-001 -> captain-001")
	
	// Test search with no results
	result = am.SearchMessages("nonexistent")
	assert.Contains(t, result, "No messages found containing 'nonexistent'")
}

func TestAgentManager_SimulateConversation_NoAgents(t *testing.T) {
	am := NewAgentManager()
	
	// Try to simulate without creating agents first
	err := am.SimulateConversation()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}