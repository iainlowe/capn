package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBaseAgent_Identity(t *testing.T) {
	agent := NewBaseAgent("agent-001", "TestAgent-1", FileAgent)
	
	assert.Equal(t, "agent-001", agent.ID())
	assert.Equal(t, "TestAgent-1", agent.Name())
	assert.Equal(t, FileAgent, agent.Type())
	assert.Equal(t, Idle, agent.Status())
	assert.Equal(t, Healthy, agent.Health())
}

func TestBaseAgent_MessageHandling(t *testing.T) {
	agent := NewBaseAgent("agent-001", "TestAgent-1", FileAgent)
	
	message := Message{
		ID:        "msg-001",
		From:      "Captain",
		To:        "agent-001",
		Content:   "Test message",
		Type:      TaskMessage,
		Timestamp: time.Now(),
	}
	
	err := agent.ReceiveMessage(message)
	assert.NoError(t, err)
	
	// Check if message was stored
	messages := agent.GetReceivedMessages()
	assert.Len(t, messages, 1)
	assert.Equal(t, "msg-001", messages[0].ID)
}

func TestBaseAgent_SendMessage(t *testing.T) {
	agent := NewBaseAgent("agent-001", "TestAgent-1", FileAgent)
	
	// Mock router
	mockRouter := &MockRouter{
		routedMessages: make(map[string]Message),
	}
	agent.SetRouter(mockRouter)
	
	message := Message{
		ID:      "msg-001",
		From:    "agent-001",
		To:      "target-agent",
		Content: "Test message",
		Type:    TaskMessage,
	}
	
	err := agent.SendMessage("target-agent", message)
	assert.NoError(t, err)
	
	// Check if message was routed
	routedMsg, exists := mockRouter.routedMessages["target-agent"]
	assert.True(t, exists)
	assert.Equal(t, "msg-001", routedMsg.ID)
}

func TestBaseAgent_Stop(t *testing.T) {
	agent := NewBaseAgent("agent-001", "TestAgent-1", FileAgent)
	
	assert.Equal(t, Idle, agent.Status())
	
	err := agent.Stop()
	assert.NoError(t, err)
	assert.Equal(t, Stopped, agent.Status())
}

func TestBaseAgent_Execute_NotImplemented(t *testing.T) {
	agent := NewBaseAgent("agent-001", "TestAgent-1", FileAgent)
	
	task := Task{
		ID:          "task-001",
		Description: "Test task",
		Parameters:  map[string]interface{}{"param": "value"},
	}
	
	result := agent.Execute(context.Background(), task)
	
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "not implemented")
}

func TestBaseAgent_SetStatus(t *testing.T) {
	agent := NewBaseAgent("agent-001", "TestAgent-1", FileAgent)
	
	assert.Equal(t, Idle, agent.Status())
	
	agent.SetStatus(Running)
	assert.Equal(t, Running, agent.Status())
	
	agent.SetStatus(Error)
	assert.Equal(t, Error, agent.Status())
}

func TestBaseAgent_SetHealth(t *testing.T) {
	agent := NewBaseAgent("agent-001", "TestAgent-1", FileAgent)
	
	assert.Equal(t, Healthy, agent.Health())
	
	agent.SetHealth(Unhealthy)
	assert.Equal(t, Unhealthy, agent.Health())
	
	agent.SetHealth(Unknown)
	assert.Equal(t, Unknown, agent.Health())
}

func TestBaseAgent_ConcurrentMessageHandling(t *testing.T) {
	agent := NewBaseAgent("agent-001", "TestAgent-1", FileAgent)
	
	// Send messages concurrently
	const numMessages = 100
	done := make(chan bool, numMessages)
	
	for i := 0; i < numMessages; i++ {
		go func(id int) {
			message := Message{
				ID:        fmt.Sprintf("msg-%03d", id),
				From:      "sender",
				To:        "agent-001",
				Content:   fmt.Sprintf("Message %d", id),
				Type:      TaskMessage,
				Timestamp: time.Now(),
			}
			err := agent.ReceiveMessage(message)
			assert.NoError(t, err)
			done <- true
		}(i)
	}
	
	// Wait for all messages to be processed
	for i := 0; i < numMessages; i++ {
		<-done
	}
	
	// Check that all messages were received
	messages := agent.GetReceivedMessages()
	assert.Len(t, messages, numMessages)
}

// MockRouter for testing
type MockRouter struct {
	routedMessages map[string]Message
}

func (m *MockRouter) RouteMessage(message Message) error {
	m.routedMessages[message.To] = message
	return nil
}