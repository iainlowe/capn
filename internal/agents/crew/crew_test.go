package crew

import (
	"context"
	"testing"
	"time"

	"github.com/iainlowe/capn/internal/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileAgent_BasicProperties(t *testing.T) {
	agent := NewFileAgent("file-1", "FileAgent-1")
	
	assert.Equal(t, "file-1", agent.ID())
	assert.Equal(t, "FileAgent-1", agent.Name())
	assert.Equal(t, agents.AgentTypeFile, agent.Type())
	assert.Equal(t, agents.AgentStatusIdle, agent.Status())
}

func TestFileAgent_Execute(t *testing.T) {
	agent := NewFileAgent("file-1", "FileAgent-1")
	
	task := agents.Task{
		ID:          "task-1",
		Type:        "file_analysis",
		Description: "Analyze Go files in ./src directory",
		Priority:    agents.PriorityMedium,
		Data: map[string]interface{}{
			"path": "./src",
			"pattern": "*.go",
		},
	}
	
	ctx := context.Background()
	result := agent.Execute(ctx, task)
	
	assert.Equal(t, "task-1", result.TaskID)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "file operation")
	assert.Contains(t, result.Output, "./src")
	assert.NotZero(t, result.Timestamp)
}

func TestNetworkAgent_BasicProperties(t *testing.T) {
	agent := NewNetworkAgent("net-1", "NetworkAgent-1")
	
	assert.Equal(t, "net-1", agent.ID())
	assert.Equal(t, "NetworkAgent-1", agent.Name())
	assert.Equal(t, agents.AgentTypeNetwork, agent.Type())
	assert.Equal(t, agents.AgentStatusIdle, agent.Status())
}

func TestNetworkAgent_Execute(t *testing.T) {
	agent := NewNetworkAgent("net-1", "NetworkAgent-1")
	
	task := agents.Task{
		ID:          "task-1",
		Type:        "api_call",
		Description: "Make API call to get user data",
		Priority:    agents.PriorityHigh,
		Data: map[string]interface{}{
			"url": "https://api.example.com/users",
			"method": "GET",
		},
	}
	
	ctx := context.Background()
	result := agent.Execute(ctx, task)
	
	assert.Equal(t, "task-1", result.TaskID)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "network operation")
	assert.Contains(t, result.Output, "https://api.example.com/users")
	assert.NotZero(t, result.Timestamp)
}

func TestResearchAgent_BasicProperties(t *testing.T) {
	agent := NewResearchAgent("research-1", "ResearchAgent-1")
	
	assert.Equal(t, "research-1", agent.ID())
	assert.Equal(t, "ResearchAgent-1", agent.Name())
	assert.Equal(t, agents.AgentTypeResearch, agent.Type())
	assert.Equal(t, agents.AgentStatusIdle, agent.Status())
}

func TestResearchAgent_Execute(t *testing.T) {
	agent := NewResearchAgent("research-1", "ResearchAgent-1")
	
	task := agents.Task{
		ID:          "task-1",
		Type:        "research",
		Description: "Research Go best practices for error handling",
		Priority:    agents.PriorityMedium,
		Data: map[string]interface{}{
			"topic": "Go error handling",
			"depth": "detailed",
		},
	}
	
	ctx := context.Background()
	result := agent.Execute(ctx, task)
	
	assert.Equal(t, "task-1", result.TaskID)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "research operation")
	assert.Contains(t, result.Output, "Go error handling")
	assert.NotZero(t, result.Timestamp)
}

func TestCrewAgents_CommunicationScenario(t *testing.T) {
	// Set up communication infrastructure
	router := agents.NewMessageRouter()
	logger := agents.NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	
	// Create crew agents
	fileAgent := NewFileAgent("file-1", "FileAgent-1")
	networkAgent := NewNetworkAgent("net-1", "NetworkAgent-1")
	researchAgent := NewResearchAgent("research-1", "ResearchAgent-1")
	
	fileAgent.SetRouter(router)
	networkAgent.SetRouter(router)
	researchAgent.SetRouter(router)
	
	// Register agents
	err := router.RegisterAgent(fileAgent)
	require.NoError(t, err)
	err = router.RegisterAgent(networkAgent) 
	require.NoError(t, err)
	err = router.RegisterAgent(researchAgent)
	require.NoError(t, err)
	
	// Simulate the communication scenario from the issue
	timestamp1 := time.Date(2025, 8, 3, 10, 15, 32, 0, time.UTC)
	msg1 := agents.Message{
		ID:        "msg-1",
		From:      "Captain",
		To:        "file-1",
		Content:   "Please analyze the Go files in ./src directory",
		Type:      agents.MessageTypeText,
		Timestamp: timestamp1,
	}
	
	// Manually log captain message (captain not implemented yet)
	logger.LogMessage("Captain", "file-1", msg1)
	
	// FileAgent responds to Captain (manually log since Captain isn't registered)
	timestamp2 := time.Date(2025, 8, 3, 10, 15, 45, 0, time.UTC)
	msg2 := agents.Message{
		ID:        "msg-2",
		From:      "file-1",
		To:        "Captain",
		Content:   "Found 23 Go files, analyzing structure...",
		Type:      agents.MessageTypeText,
		Timestamp: timestamp2,
	}
	// Manually log this message since Captain is not registered
	logger.LogMessage("file-1", "Captain", msg2)
	
	// FileAgent asks ResearchAgent for help
	timestamp3 := time.Date(2025, 8, 3, 10, 16, 12, 0, time.UTC)
	msg3 := agents.Message{
		ID:        "msg-3",
		Content:   "Hey Research, can you look up best practices for this pattern?",
		Type:      agents.MessageTypeText,
		Timestamp: timestamp3,
	}
	err = fileAgent.SendMessage("research-1", msg3)
	require.NoError(t, err)
	
	// ResearchAgent responds
	timestamp4 := time.Date(2025, 8, 3, 10, 16, 28, 0, time.UTC)
	msg4 := agents.Message{
		ID:        "msg-4", 
		Content:   "Sure! Found 5 relevant patterns in Go documentation",
		Type:      agents.MessageTypeText,
		Timestamp: timestamp4,
	}
	err = researchAgent.SendMessage("file-1", msg4)
	require.NoError(t, err)
	
	// Verify communication was logged
	allMessages := logger.GetAllMessages()
	assert.Len(t, allMessages, 4)
	
	// Check the IRC/Slack format
	assert.Contains(t, allMessages[0].Formatted, "[2025-08-03 10:15:32] Captain -> file-1:")
	assert.Contains(t, allMessages[0].Formatted, "Please analyze the Go files in ./src directory")
	
	// Verify agents received their messages
	fileMessages := fileAgent.GetReceivedMessages()
	researchMessages := researchAgent.GetReceivedMessages()
	
	assert.Len(t, fileMessages, 1) // Received response from research agent
	assert.Len(t, researchMessages, 1) // Received request from file agent
	
	// Search functionality
	searchResults := logger.SearchMessages("Go files")
	assert.Len(t, searchResults, 2) // Should find messages mentioning "Go files"
}