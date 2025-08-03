package agents

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryCommunicationLogger is an in-memory implementation of CommunicationLogger
type MemoryCommunicationLogger struct {
	mu       sync.RWMutex
	messages []MessageLog
}

// NewMemoryCommunicationLogger creates a new in-memory communication logger
func NewMemoryCommunicationLogger() *MemoryCommunicationLogger {
	return &MemoryCommunicationLogger{
		messages: make([]MessageLog, 0),
	}
}

// LogMessage logs a message between agents
func (l *MemoryCommunicationLogger) LogMessage(from, to string, message Message) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Format the message in IRC/Slack style
	formatted := fmt.Sprintf("[%s] %s -> %s: \"%s\"",
		message.Timestamp.Format("2006-01-02 15:04:05"),
		from,
		to,
		message.Content,
	)
	
	log := MessageLog{
		Message:   message,
		Formatted: formatted,
		Indexed:   time.Now(),
	}
	
	l.messages = append(l.messages, log)
}

// GetHistory returns message history for a specific agent (messages to or from the agent)
func (l *MemoryCommunicationLogger) GetHistory(agentID string) []MessageLog {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	var history []MessageLog
	for _, log := range l.messages {
		if log.Message.From == agentID || log.Message.To == agentID {
			history = append(history, log)
		}
	}
	
	return history
}

// SearchMessages searches for messages containing the query string
func (l *MemoryCommunicationLogger) SearchMessages(query string) []MessageLog {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	var results []MessageLog
	queryLower := strings.ToLower(query)
	
	for _, log := range l.messages {
		// Search in message content, from, and to fields
		if strings.Contains(strings.ToLower(log.Message.Content), queryLower) ||
			strings.Contains(strings.ToLower(log.Message.From), queryLower) ||
			strings.Contains(strings.ToLower(log.Message.To), queryLower) {
			results = append(results, log)
		}
	}
	
	return results
}

// GetAllMessages returns all logged messages
func (l *MemoryCommunicationLogger) GetAllMessages() []MessageLog {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Return a copy to prevent external modification
	messages := make([]MessageLog, len(l.messages))
	copy(messages, l.messages)
	return messages
}

// Clear removes all logged messages
func (l *MemoryCommunicationLogger) Clear() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.messages = make([]MessageLog, 0)
	return nil
}