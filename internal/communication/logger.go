package communication

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/iainlowe/capn/internal/agent"
)

// MessageLog represents a logged message
type MessageLog struct {
	MessageID string                 `json:"message_id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Content   string                 `json:"content"`
	Type      agent.MessageType      `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewMessageLog creates a new MessageLog from a Message
func NewMessageLog(message agent.Message) MessageLog {
	return MessageLog{
		MessageID: message.ID,
		From:      message.From,
		To:        message.To,
		Content:   message.Content,
		Type:      message.Type,
		Data:      message.Data,
		Timestamp: message.Timestamp,
	}
}

// CommunicationLogger interface defines the contract for logging agent communications
type CommunicationLogger interface {
	LogMessage(from, to string, message agent.Message)
	GetHistory(agentID string) []MessageLog
	SearchMessages(query string) []MessageLog
	FormatMessage(message agent.Message) string
	GetAllMessages() []MessageLog
}

// InMemoryLogger is an in-memory implementation of CommunicationLogger
type InMemoryLogger struct {
	messages []MessageLog
	mutex    sync.RWMutex
}

// NewInMemoryLogger creates a new in-memory logger
func NewInMemoryLogger() *InMemoryLogger {
	return &InMemoryLogger{
		messages: make([]MessageLog, 0),
	}
}

// LogMessage logs a message from one agent to another
func (l *InMemoryLogger) LogMessage(from, to string, message agent.Message) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	log := NewMessageLog(message)
	l.messages = append(l.messages, log)
}

// GetHistory returns the message history for a specific agent
func (l *InMemoryLogger) GetHistory(agentID string) []MessageLog {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	var history []MessageLog
	for _, msg := range l.messages {
		if msg.From == agentID {
			history = append(history, msg)
		}
	}
	return history
}

// SearchMessages searches for messages containing the query string
func (l *InMemoryLogger) SearchMessages(query string) []MessageLog {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	var results []MessageLog
	queryLower := strings.ToLower(query)

	for _, msg := range l.messages {
		contentLower := strings.ToLower(msg.Content)
		fromLower := strings.ToLower(msg.From)
		toLower := strings.ToLower(msg.To)

		if strings.Contains(contentLower, queryLower) ||
			strings.Contains(fromLower, queryLower) ||
			strings.Contains(toLower, queryLower) {
			results = append(results, msg)
		}
	}
	return results
}

// FormatMessage formats a message in IRC/Slack style
func (l *InMemoryLogger) FormatMessage(message agent.Message) string {
	timestamp := message.Timestamp.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] %s -> %s: \"%s\"", timestamp, message.From, message.To, message.Content)
}

// GetAllMessages returns all messages ordered by timestamp
func (l *InMemoryLogger) GetAllMessages() []MessageLog {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	// Create a copy of the messages slice
	messages := make([]MessageLog, len(l.messages))
	copy(messages, l.messages)

	// Sort by timestamp
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	return messages
}