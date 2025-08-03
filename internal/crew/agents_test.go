package crew

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iainlowe/capn/internal/agent"
)

func TestFileAgent_Creation(t *testing.T) {
	fileAgent := NewFileAgent("file-001", "FileAgent-1")
	
	assert.Equal(t, "file-001", fileAgent.ID())
	assert.Equal(t, "FileAgent-1", fileAgent.Name())
	assert.Equal(t, agent.FileAgent, fileAgent.Type())
	assert.Equal(t, agent.Idle, fileAgent.Status())
	assert.Equal(t, agent.Healthy, fileAgent.Health())
}

func TestFileAgent_Execute(t *testing.T) {
	fileAgent := NewFileAgent("file-001", "FileAgent-1")
	
	task := agent.Task{
		ID:          "task-001",
		Description: "List files in directory",
		Parameters:  map[string]interface{}{"path": "/tmp"},
	}
	
	result := fileAgent.Execute(context.Background(), task)
	
	assert.True(t, result.IsSuccess())
	assert.Equal(t, task.ID, result.TaskID)
	assert.Contains(t, result.Data, "FileAgent")
	assert.Contains(t, result.Data, "file system operations")
}

func TestNetworkAgent_Creation(t *testing.T) {
	networkAgent := NewNetworkAgent("network-001", "NetworkAgent-1")
	
	assert.Equal(t, "network-001", networkAgent.ID())
	assert.Equal(t, "NetworkAgent-1", networkAgent.Name())
	assert.Equal(t, agent.NetworkAgent, networkAgent.Type())
	assert.Equal(t, agent.Idle, networkAgent.Status())
	assert.Equal(t, agent.Healthy, networkAgent.Health())
}

func TestNetworkAgent_Execute(t *testing.T) {
	networkAgent := NewNetworkAgent("network-001", "NetworkAgent-1")
	
	task := agent.Task{
		ID:          "task-001",
		Description: "Fetch API data",
		Parameters:  map[string]interface{}{"url": "https://api.example.com"},
	}
	
	result := networkAgent.Execute(context.Background(), task)
	
	assert.True(t, result.IsSuccess())
	assert.Equal(t, task.ID, result.TaskID)
	assert.Contains(t, result.Data, "NetworkAgent")
	assert.Contains(t, result.Data, "API interactions")
}

func TestResearchAgent_Creation(t *testing.T) {
	researchAgent := NewResearchAgent("research-001", "ResearchAgent-1")
	
	assert.Equal(t, "research-001", researchAgent.ID())
	assert.Equal(t, "ResearchAgent-1", researchAgent.Name())
	assert.Equal(t, agent.ResearchAgent, researchAgent.Type())
	assert.Equal(t, agent.Idle, researchAgent.Status())
	assert.Equal(t, agent.Healthy, researchAgent.Health())
}

func TestResearchAgent_Execute(t *testing.T) {
	researchAgent := NewResearchAgent("research-001", "ResearchAgent-1")
	
	task := agent.Task{
		ID:          "task-001",
		Description: "Research Go best practices",
		Parameters:  map[string]interface{}{"topic": "Go error handling"},
	}
	
	result := researchAgent.Execute(context.Background(), task)
	
	assert.True(t, result.IsSuccess())
	assert.Equal(t, task.ID, result.TaskID)
	assert.Contains(t, result.Data, "ResearchAgent")
	assert.Contains(t, result.Data, "information gathering")
}

func TestAllAgents_Communication(t *testing.T) {
	fileAgent := NewFileAgent("file-001", "FileAgent-1")
	networkAgent := NewNetworkAgent("network-001", "NetworkAgent-1")
	researchAgent := NewResearchAgent("research-001", "ResearchAgent-1")
	
	// Mock router for testing communication
	mockRouter := &MockRouter{
		routedMessages: make(map[string]agent.Message),
	}
	
	fileAgent.SetRouter(mockRouter)
	networkAgent.SetRouter(mockRouter)
	researchAgent.SetRouter(mockRouter)
	
	// Test FileAgent sending message to NetworkAgent
	message := agent.Message{
		ID:      "msg-001",
		Content: "Please fetch data from API",
		Type:    agent.TaskMessage,
	}
	
	err := fileAgent.SendMessage("network-001", message)
	assert.NoError(t, err)
	
	// Check if message was routed
	routedMsg, exists := mockRouter.routedMessages["network-001"]
	assert.True(t, exists)
	assert.Equal(t, "file-001", routedMsg.From)
	assert.Equal(t, "network-001", routedMsg.To)
}

// MockRouter for testing
type MockRouter struct {
	routedMessages map[string]agent.Message
}

func (m *MockRouter) RouteMessage(message agent.Message) error {
	m.routedMessages[message.To] = message
	return nil
}