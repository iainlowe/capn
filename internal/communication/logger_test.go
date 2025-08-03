package communication

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iainlowe/capn/internal/agent"
)

func TestMessageLog_CreatedCorrectly(t *testing.T) {
	now := time.Now()
	message := agent.Message{
		ID:        "msg-001",
		From:      "Captain",
		To:        "FileAgent-1",
		Content:   "Please analyze files",
		Type:      agent.TaskMessage,
		Timestamp: now,
	}

	log := NewMessageLog(message)

	assert.Equal(t, message.ID, log.MessageID)
	assert.Equal(t, message.From, log.From)
	assert.Equal(t, message.To, log.To)
	assert.Equal(t, message.Content, log.Content)
	assert.Equal(t, message.Type, log.Type)
	assert.Equal(t, now, log.Timestamp)
}

func TestInMemoryLogger_LogMessage(t *testing.T) {
	logger := NewInMemoryLogger()

	message := agent.Message{
		ID:        "msg-001",
		From:      "Captain",
		To:        "FileAgent-1",
		Content:   "Please analyze files",
		Type:      agent.TaskMessage,
		Timestamp: time.Now(),
	}

	logger.LogMessage(message.From, message.To, message)

	history := logger.GetHistory("Captain")
	require.Len(t, history, 1)
	assert.Equal(t, message.ID, history[0].MessageID)
	assert.Equal(t, message.From, history[0].From)
	assert.Equal(t, message.To, history[0].To)
}

func TestInMemoryLogger_GetHistory(t *testing.T) {
	logger := NewInMemoryLogger()

	// Log multiple messages from different agents
	messages := []agent.Message{
		{
			ID:        "msg-001",
			From:      "Captain",
			To:        "FileAgent-1",
			Content:   "Task 1",
			Type:      agent.TaskMessage,
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-002",
			From:      "FileAgent-1",
			To:        "Captain",
			Content:   "Response 1",
			Type:      agent.ResponseMessage,
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-003",
			From:      "Captain",
			To:        "ResearchAgent-1",
			Content:   "Task 2",
			Type:      agent.TaskMessage,
			Timestamp: time.Now(),
		},
	}

	for _, msg := range messages {
		logger.LogMessage(msg.From, msg.To, msg)
	}

	// Check Captain's history (should have 2 messages sent)
	captainHistory := logger.GetHistory("Captain")
	assert.Len(t, captainHistory, 2)

	// Check FileAgent-1's history (should have 1 message sent)
	fileAgentHistory := logger.GetHistory("FileAgent-1")
	assert.Len(t, fileAgentHistory, 1)

	// Check non-existent agent
	unknownHistory := logger.GetHistory("UnknownAgent")
	assert.Len(t, unknownHistory, 0)
}

func TestInMemoryLogger_SearchMessages(t *testing.T) {
	logger := NewInMemoryLogger()

	messages := []agent.Message{
		{
			ID:        "msg-001",
			From:      "Captain",
			To:        "FileAgent-1",
			Content:   "Please analyze the Go files",
			Type:      agent.TaskMessage,
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-002",
			From:      "FileAgent-1",
			To:        "Captain",
			Content:   "Found 23 files to analyze",
			Type:      agent.ResponseMessage,
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-003",
			From:      "Captain",
			To:        "NetworkAgent-1",
			Content:   "Fetch API documentation",
			Type:      agent.TaskMessage,
			Timestamp: time.Now(),
		},
	}

	for _, msg := range messages {
		logger.LogMessage(msg.From, msg.To, msg)
	}

	// Search for messages containing "files"
	results := logger.SearchMessages("files")
	assert.Len(t, results, 2)

	// Search for messages containing "API"
	results = logger.SearchMessages("API")
	assert.Len(t, results, 1)
	assert.Equal(t, "msg-003", results[0].MessageID)

	// Search for non-existent content
	results = logger.SearchMessages("nonexistent")
	assert.Len(t, results, 0)
}

func TestInMemoryLogger_FormatMessage(t *testing.T) {
	logger := NewInMemoryLogger()

	timestamp := time.Date(2023, 8, 3, 10, 15, 32, 0, time.UTC)
	message := agent.Message{
		ID:        "msg-001",
		From:      "Captain",
		To:        "FileAgent-1",
		Content:   "Please analyze files",
		Type:      agent.TaskMessage,
		Timestamp: timestamp,
	}

	formatted := logger.FormatMessage(message)
	expected := "[2023-08-03 10:15:32] Captain -> FileAgent-1: \"Please analyze files\""
	assert.Equal(t, expected, formatted)
}

func TestInMemoryLogger_GetAllMessages(t *testing.T) {
	logger := NewInMemoryLogger()

	messages := []agent.Message{
		{
			ID:        "msg-001",
			From:      "Captain",
			To:        "FileAgent-1",
			Content:   "Task 1",
			Type:      agent.TaskMessage,
			Timestamp: time.Now().Add(-2 * time.Minute),
		},
		{
			ID:        "msg-002",
			From:      "FileAgent-1",
			To:        "Captain",
			Content:   "Response 1",
			Type:      agent.ResponseMessage,
			Timestamp: time.Now().Add(-1 * time.Minute),
		},
		{
			ID:        "msg-003",
			From:      "Captain",
			To:        "ResearchAgent-1",
			Content:   "Task 2",
			Type:      agent.TaskMessage,
			Timestamp: time.Now(),
		},
	}

	for _, msg := range messages {
		logger.LogMessage(msg.From, msg.To, msg)
	}

	allMessages := logger.GetAllMessages()
	assert.Len(t, allMessages, 3)

	// Messages should be ordered by timestamp (oldest first)
	assert.Equal(t, "msg-001", allMessages[0].MessageID)
	assert.Equal(t, "msg-002", allMessages[1].MessageID)
	assert.Equal(t, "msg-003", allMessages[2].MessageID)
}