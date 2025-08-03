package agents

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCommunicationLogger_LogMessage(t *testing.T) {
	logger := NewMemoryCommunicationLogger()
	
	message := Message{
		ID:        "msg-1",
		From:      "agent-1",
		To:        "agent-2",
		Content:   "Hello world",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	logger.LogMessage(message.From, message.To, message)
	
	history := logger.GetAllMessages()
	require.Len(t, history, 1)
	
	logged := history[0]
	assert.Equal(t, message.ID, logged.Message.ID)
	assert.Equal(t, message.From, logged.Message.From)
	assert.Equal(t, message.To, logged.Message.To)
	assert.Equal(t, message.Content, logged.Message.Content)
	assert.Contains(t, logged.Formatted, "agent-1 -> agent-2:")
	assert.Contains(t, logged.Formatted, "Hello world")
}

func TestMemoryCommunicationLogger_GetHistory(t *testing.T) {
	logger := NewMemoryCommunicationLogger()
	
	// Add messages involving agent-1
	msg1 := Message{
		ID:        "msg-1",
		From:      "agent-1",
		To:        "agent-2",
		Content:   "Hello from 1",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	msg2 := Message{
		ID:        "msg-2",
		From:      "agent-2",
		To:        "agent-1",
		Content:   "Hello back to 1",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	msg3 := Message{
		ID:        "msg-3",
		From:      "agent-3",
		To:        "agent-4",
		Content:   "Unrelated message",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	logger.LogMessage(msg1.From, msg1.To, msg1)
	logger.LogMessage(msg2.From, msg2.To, msg2)
	logger.LogMessage(msg3.From, msg3.To, msg3)
	
	// Get history for agent-1 (should return messages where agent-1 is from or to)
	history := logger.GetHistory("agent-1")
	require.Len(t, history, 2)
	
	// Verify the messages are related to agent-1
	for _, log := range history {
		assert.True(t, log.Message.From == "agent-1" || log.Message.To == "agent-1")
	}
}

func TestMemoryCommunicationLogger_SearchMessages(t *testing.T) {
	logger := NewMemoryCommunicationLogger()
	
	messages := []Message{
		{
			ID:        "msg-1",
			From:      "agent-1",
			To:        "agent-2",
			Content:   "Please analyze the Go files",
			Type:      MessageTypeText,
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-2",
			From:      "agent-2",
			To:        "agent-1",
			Content:   "Found 23 Go files, analyzing structure",
			Type:      MessageTypeText,
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-3",
			From:      "agent-1",
			To:        "agent-3",
			Content:   "Can you research best practices?",
			Type:      MessageTypeText,
			Timestamp: time.Now(),
		},
	}
	
	for _, msg := range messages {
		logger.LogMessage(msg.From, msg.To, msg)
	}
	
	// Search for "Go files"
	results := logger.SearchMessages("Go files")
	require.Len(t, results, 2)
	
	// Search for "research"
	results = logger.SearchMessages("research")
	require.Len(t, results, 1)
	assert.Equal(t, "msg-3", results[0].Message.ID)
	
	// Search for non-existent term
	results = logger.SearchMessages("nonexistent")
	assert.Len(t, results, 0)
}

func TestMemoryCommunicationLogger_Clear(t *testing.T) {
	logger := NewMemoryCommunicationLogger()
	
	message := Message{
		ID:        "msg-1",
		From:      "agent-1",
		To:        "agent-2",
		Content:   "Hello world",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	
	logger.LogMessage(message.From, message.To, message)
	
	// Verify message was logged
	history := logger.GetAllMessages()
	require.Len(t, history, 1)
	
	// Clear the logger
	err := logger.Clear()
	assert.NoError(t, err)
	
	// Verify messages are cleared
	history = logger.GetAllMessages()
	assert.Len(t, history, 0)
}

func TestMemoryCommunicationLogger_FormattedMessage(t *testing.T) {
	logger := NewMemoryCommunicationLogger()
	
	timestamp := time.Date(2025, 8, 3, 10, 15, 32, 0, time.UTC)
	message := Message{
		ID:        "msg-1",
		From:      "Captain",
		To:        "FileAgent-1",
		Content:   "Please analyze the Go files in ./src directory",
		Type:      MessageTypeText,
		Timestamp: timestamp,
	}
	
	logger.LogMessage(message.From, message.To, message)
	
	history := logger.GetAllMessages()
	require.Len(t, history, 1)
	
	formatted := history[0].Formatted
	expectedStart := "[2025-08-03 10:15:32] Captain -> FileAgent-1:"
	assert.Contains(t, formatted, expectedStart)
	assert.Contains(t, formatted, "Please analyze the Go files in ./src directory")
}