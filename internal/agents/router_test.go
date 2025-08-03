package agents

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageRouter_RegisterAgent(t *testing.T) {
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	
	agent := &MockAgent{
		id:     "test-agent",
		name:   "TestAgent",
		status: AgentStatusIdle,
	}
	
	err := router.RegisterAgent(agent)
	assert.NoError(t, err)
	
	// Try to register the same agent again
	err = router.RegisterAgent(agent)
	assert.Error(t, err)
}

func TestMessageRouter_UnregisterAgent(t *testing.T) {
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	
	agent := &MockAgent{
		id:     "test-agent",
		name:   "TestAgent",
		status: AgentStatusIdle,
	}
	
	// Register and then unregister
	err := router.RegisterAgent(agent)
	require.NoError(t, err)
	
	err = router.UnregisterAgent("test-agent")
	assert.NoError(t, err)
	
	// Try to unregister non-existent agent
	err = router.UnregisterAgent("non-existent")
	assert.Error(t, err)
}

func TestMessageRouter_RouteMessage(t *testing.T) {
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	
	sender := &MockAgent{
		id:     "sender",
		name:   "SenderAgent",
		status: AgentStatusIdle,
	}
	
	receiver := &MockAgent{
		id:       "receiver",
		name:     "ReceiverAgent",
		status:   AgentStatusIdle,
		messages: make([]Message, 0),
	}
	
	// Register both agents
	err := router.RegisterAgent(sender)
	require.NoError(t, err)
	err = router.RegisterAgent(receiver)
	require.NoError(t, err)
	
	// Route a message
	message := Message{
		ID:        "msg-1",
		From:      "sender",
		To:        "receiver",
		Content:   "Hello receiver",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	err = router.RouteMessage(message)
	assert.NoError(t, err)
	
	// Check that receiver got the message
	assert.Len(t, receiver.messages, 1)
	assert.Equal(t, "Hello receiver", receiver.messages[0].Content)
	
	// Check that message was logged
	history := logger.GetAllMessages()
	assert.Len(t, history, 1)
	assert.Contains(t, history[0].Formatted, "sender -> receiver:")
}

func TestMessageRouter_RouteMessage_UnknownReceiver(t *testing.T) {
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	
	message := Message{
		ID:        "msg-1",
		From:      "sender",
		To:        "unknown",
		Content:   "Hello unknown",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	err := router.RouteMessage(message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
}

func TestMessageRouter_BroadcastMessage(t *testing.T) {
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	
	sender := &MockAgent{
		id:     "sender",
		name:   "SenderAgent",
		status: AgentStatusIdle,
	}
	
	receiver1 := &MockAgent{
		id:       "receiver1",
		name:     "Receiver1",
		status:   AgentStatusIdle,
		messages: make([]Message, 0),
	}
	
	receiver2 := &MockAgent{
		id:       "receiver2",
		name:     "Receiver2",
		status:   AgentStatusIdle,
		messages: make([]Message, 0),
	}
	
	// Register all agents
	err := router.RegisterAgent(sender)
	require.NoError(t, err)
	err = router.RegisterAgent(receiver1)
	require.NoError(t, err)
	err = router.RegisterAgent(receiver2)
	require.NoError(t, err)
	
	// Broadcast a message
	message := Message{
		ID:        "msg-1",
		From:      "sender",
		To:        "all",
		Content:   "Hello everyone",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	err = router.BroadcastMessage(message)
	assert.NoError(t, err)
	
	// Check that both receivers got the message (but not the sender)
	assert.Len(t, receiver1.messages, 1)
	assert.Len(t, receiver2.messages, 1)
	
	// Check that messages were logged
	history := logger.GetAllMessages()
	assert.Len(t, history, 2) // One for each receiver
}

func TestMessageRouter_GetRegisteredAgents(t *testing.T) {
	router := NewMessageRouter()
	
	agent1 := &MockAgent{id: "agent1", name: "Agent1", status: AgentStatusIdle}
	agent2 := &MockAgent{id: "agent2", name: "Agent2", status: AgentStatusBusy}
	
	err := router.RegisterAgent(agent1)
	require.NoError(t, err)
	err = router.RegisterAgent(agent2)
	require.NoError(t, err)
	
	agents := router.GetRegisteredAgents()
	assert.Len(t, agents, 2)
	
	// Check that we got both agents
	agentIDs := make([]string, len(agents))
	for i, agent := range agents {
		agentIDs[i] = agent.ID()
	}
	assert.Contains(t, agentIDs, "agent1")
	assert.Contains(t, agentIDs, "agent2")
}

// MockAgent is a test implementation of the Agent interface
type MockAgent struct {
	id       string
	name     string
	agentType AgentType
	status   AgentStatus
	messages []Message
	stopped  bool
}

func (m *MockAgent) ID() string {
	return m.id
}

func (m *MockAgent) Name() string {
	return m.name
}

func (m *MockAgent) Type() AgentType {
	return m.agentType
}

func (m *MockAgent) Status() AgentStatus {
	return m.status
}

func (m *MockAgent) Health() HealthStatus {
	return HealthStatus{
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
	}
}

func (m *MockAgent) Execute(ctx context.Context, task Task) Result {
	return Result{
		TaskID:    task.ID,
		Success:   true,
		Output:    "Task executed successfully",
		Timestamp: time.Now(),
	}
}

func (m *MockAgent) SendMessage(to string, message Message) error {
	// For testing, we don't actually send - this would be handled by the router
	return nil
}

func (m *MockAgent) ReceiveMessage(message Message) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *MockAgent) Stop() error {
	m.stopped = true
	m.status = AgentStatusStopped
	return nil
}