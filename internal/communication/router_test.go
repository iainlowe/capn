package communication

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iainlowe/capn/internal/agent"
)

// MockAgent is a mock implementation for testing
type MockAgent struct {
	id              string
	name            string
	agentType       agent.AgentType
	status          agent.AgentStatus
	health          agent.HealthStatus
	receivedMessages []agent.Message
	mutex           sync.Mutex
}

func NewMockAgent(id, name string, agentType agent.AgentType) *MockAgent {
	return &MockAgent{
		id:               id,
		name:             name,
		agentType:        agentType,
		status:           agent.Idle,
		health:           agent.Healthy,
		receivedMessages: make([]agent.Message, 0),
	}
}

func (m *MockAgent) ID() string               { return m.id }
func (m *MockAgent) Name() string             { return m.name }
func (m *MockAgent) Type() agent.AgentType    { return m.agentType }
func (m *MockAgent) Status() agent.AgentStatus { return m.status }
func (m *MockAgent) Health() agent.HealthStatus { return m.health }

func (m *MockAgent) Execute(ctx context.Context, task agent.Task) agent.Result {
	return agent.Result{TaskID: task.ID, Success: true, Data: "mock result"}
}

func (m *MockAgent) SendMessage(to string, message agent.Message) error {
	// In real implementation, this would route through the message router
	return nil
}

func (m *MockAgent) ReceiveMessage(message agent.Message) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.receivedMessages = append(m.receivedMessages, message)
	return nil
}

func (m *MockAgent) Stop() error {
	m.status = agent.Stopped
	return nil
}

func (m *MockAgent) GetReceivedMessages() []agent.Message {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]agent.Message(nil), m.receivedMessages...)
}

func TestMessageRouter_RegisterAgent(t *testing.T) {
	router := NewMessageRouter()
	agent1 := NewMockAgent("agent-001", "FileAgent-1", agent.FileAgent)
	
	err := router.RegisterAgent(agent1)
	assert.NoError(t, err)
	
	// Check if agent is registered
	agents := router.GetRegisteredAgents()
	assert.Len(t, agents, 1)
	assert.Equal(t, "agent-001", agents[0])
}

func TestMessageRouter_RegisterDuplicateAgent(t *testing.T) {
	router := NewMessageRouter()
	agent1 := NewMockAgent("agent-001", "FileAgent-1", agent.FileAgent)
	agent2 := NewMockAgent("agent-001", "FileAgent-2", agent.FileAgent)
	
	err := router.RegisterAgent(agent1)
	assert.NoError(t, err)
	
	err = router.RegisterAgent(agent2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestMessageRouter_UnregisterAgent(t *testing.T) {
	router := NewMessageRouter()
	agent1 := NewMockAgent("agent-001", "FileAgent-1", agent.FileAgent)
	
	err := router.RegisterAgent(agent1)
	require.NoError(t, err)
	
	err = router.UnregisterAgent("agent-001")
	assert.NoError(t, err)
	
	// Check if agent is unregistered
	agents := router.GetRegisteredAgents()
	assert.Len(t, agents, 0)
}

func TestMessageRouter_UnregisterNonExistentAgent(t *testing.T) {
	router := NewMessageRouter()
	
	err := router.UnregisterAgent("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMessageRouter_RouteMessage(t *testing.T) {
	router := NewMessageRouter()
	logger := NewInMemoryLogger()
	router.SetLogger(logger)
	
	sender := NewMockAgent("sender-001", "Captain", agent.Captain)
	receiver := NewMockAgent("receiver-001", "FileAgent-1", agent.FileAgent)
	
	err := router.RegisterAgent(sender)
	require.NoError(t, err)
	err = router.RegisterAgent(receiver)
	require.NoError(t, err)
	
	message := agent.Message{
		ID:      "msg-001",
		From:    "sender-001",
		To:      "receiver-001",
		Content: "Test message",
		Type:    agent.TaskMessage,
	}
	
	err = router.RouteMessage(message)
	assert.NoError(t, err)
	
	// Check if receiver got the message
	receivedMessages := receiver.GetReceivedMessages()
	assert.Len(t, receivedMessages, 1)
	assert.Equal(t, "msg-001", receivedMessages[0].ID)
	
	// Check if message was logged
	history := logger.GetHistory("sender-001")
	assert.Len(t, history, 1)
}

func TestMessageRouter_RouteToNonExistentAgent(t *testing.T) {
	router := NewMessageRouter()
	
	message := agent.Message{
		ID:   "msg-001",
		From: "sender-001",
		To:   "non-existent",
	}
	
	err := router.RouteMessage(message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMessageRouter_BroadcastMessage(t *testing.T) {
	router := NewMessageRouter()
	
	sender := NewMockAgent("sender-001", "Captain", agent.Captain)
	receiver1 := NewMockAgent("receiver-001", "FileAgent-1", agent.FileAgent)
	receiver2 := NewMockAgent("receiver-002", "NetworkAgent-1", agent.NetworkAgent)
	
	err := router.RegisterAgent(sender)
	require.NoError(t, err)
	err = router.RegisterAgent(receiver1)
	require.NoError(t, err)
	err = router.RegisterAgent(receiver2)
	require.NoError(t, err)
	
	message := agent.Message{
		ID:      "msg-001",
		From:    "sender-001",
		Content: "Broadcast message",
		Type:    agent.StatusMessage,
	}
	
	excludeIDs := []string{"sender-001"} // Don't send to sender
	err = router.BroadcastMessage(message, excludeIDs)
	assert.NoError(t, err)
	
	// Check if both receivers got the message
	receivedMessages1 := receiver1.GetReceivedMessages()
	assert.Len(t, receivedMessages1, 1)
	
	receivedMessages2 := receiver2.GetReceivedMessages()
	assert.Len(t, receivedMessages2, 1)
	
	// Check sender didn't receive the message
	senderMessages := sender.GetReceivedMessages()
	assert.Len(t, senderMessages, 0)
}