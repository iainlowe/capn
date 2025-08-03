package agents

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseAgent_BasicProperties(t *testing.T) {
	agent := NewBaseAgent("test-1", "TestAgent", AgentTypeFile)
	
	assert.Equal(t, "test-1", agent.ID())
	assert.Equal(t, "TestAgent", agent.Name())
	assert.Equal(t, AgentTypeFile, agent.Type())
	assert.Equal(t, AgentStatusIdle, agent.Status())
}

func TestBaseAgent_Health(t *testing.T) {
	agent := NewBaseAgent("test-1", "TestAgent", AgentTypeFile)
	
	health := agent.Health()
	assert.Equal(t, HealthStatusHealthy, health.Status)
	assert.NotZero(t, health.Timestamp)
}

func TestBaseAgent_Execute(t *testing.T) {
	agent := NewBaseAgent("test-1", "TestAgent", AgentTypeFile)
	
	task := Task{
		ID:          "task-1",
		Type:        "test",
		Description: "Test task",
		Priority:    PriorityMedium,
	}
	
	ctx := context.Background()
	result := agent.Execute(ctx, task)
	
	assert.Equal(t, "task-1", result.TaskID)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "executed by TestAgent")
	assert.NotZero(t, result.Timestamp)
}

func TestBaseAgent_ReceiveMessage(t *testing.T) {
	agent := NewBaseAgent("test-1", "TestAgent", AgentTypeFile)
	
	message := Message{
		ID:        "msg-1",
		From:      "sender",
		To:        "test-1",
		Content:   "Hello TestAgent",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	err := agent.ReceiveMessage(message)
	assert.NoError(t, err)
	
	// Check message was stored
	messages := agent.GetReceivedMessages()
	require.Len(t, messages, 1)
	assert.Equal(t, "Hello TestAgent", messages[0].Content)
}

func TestBaseAgent_SendMessage_NoRouter(t *testing.T) {
	agent := NewBaseAgent("test-1", "TestAgent", AgentTypeFile)
	
	message := Message{
		ID:        "msg-1",
		From:      "test-1",
		To:        "receiver",
		Content:   "Hello receiver",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	// Should return error when no router is set
	err := agent.SendMessage("receiver", message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no router configured")
}

func TestBaseAgent_SendMessage_WithRouter(t *testing.T) {
	router := NewMessageRouter()
	logger := NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	
	sender := NewBaseAgent("sender", "SenderAgent", AgentTypeFile)
	receiver := NewBaseAgent("receiver", "ReceiverAgent", AgentTypeNetwork)
	
	// Set router for sender
	sender.SetRouter(router)
	
	// Register both agents
	err := router.RegisterAgent(sender)
	require.NoError(t, err)
	err = router.RegisterAgent(receiver)
	require.NoError(t, err)
	
	message := Message{
		ID:        "msg-1",
		From:      "sender",
		To:        "receiver",
		Content:   "Hello receiver",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	err = sender.SendMessage("receiver", message)
	assert.NoError(t, err)
	
	// Check that receiver got the message
	messages := receiver.GetReceivedMessages()
	require.Len(t, messages, 1)
	assert.Equal(t, "Hello receiver", messages[0].Content)
}

func TestBaseAgent_Stop(t *testing.T) {
	agent := NewBaseAgent("test-1", "TestAgent", AgentTypeFile)
	
	// Initially idle
	assert.Equal(t, AgentStatusIdle, agent.Status())
	
	// Stop the agent
	err := agent.Stop()
	assert.NoError(t, err)
	assert.Equal(t, AgentStatusStopped, agent.Status())
	
	// Stopping again should be safe
	err = agent.Stop()
	assert.NoError(t, err)
	assert.Equal(t, AgentStatusStopped, agent.Status())
}

func TestBaseAgent_ConcurrentAccess(t *testing.T) {
	agent := NewBaseAgent("test-1", "TestAgent", AgentTypeFile)
	
	// Test concurrent message receiving
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(i int) {
			message := Message{
				ID:        fmt.Sprintf("msg-%d", i),
				From:      "sender",
				To:        "test-1",
				Content:   fmt.Sprintf("Message %d", i),
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			}
			
			err := agent.ReceiveMessage(message)
			assert.NoError(t, err)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	messages := agent.GetReceivedMessages()
	assert.Len(t, messages, 10)
}