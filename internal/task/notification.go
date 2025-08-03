package task

import (
	"fmt"
	"sync"
)

// BasicNotificationService provides a simple console-based notification service
type BasicNotificationService struct {
	mu          sync.RWMutex
	preferences NotificationPreferences
}

// NewBasicNotificationService creates a new basic notification service
func NewBasicNotificationService() *BasicNotificationService {
	return &BasicNotificationService{
		preferences: NotificationPreferences{
			EnableCompletion: true,
			EnableErrors:     true,
			OutputFormat:     "console",
		},
	}
}

// NotifyCompletion notifies when a task completes successfully
func (s *BasicNotificationService) NotifyCompletion(task *TaskExecution) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if !s.preferences.EnableCompletion {
		return nil
	}
	
	switch s.preferences.OutputFormat {
	case "console":
		fmt.Printf("[COMPLETED] %s: %s\n", task.ID, task.Goal)
	default:
		// Default to console output
		fmt.Printf("[COMPLETED] %s: %s\n", task.ID, task.Goal)
	}
	
	return nil
}

// NotifyError notifies when a task encounters an error
func (s *BasicNotificationService) NotifyError(task *TaskExecution, err error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if !s.preferences.EnableErrors {
		return nil
	}
	
	switch s.preferences.OutputFormat {
	case "console":
		fmt.Printf("[ERROR] %s: %s - %v\n", task.ID, task.Goal, err)
	default:
		// Default to console output
		fmt.Printf("[ERROR] %s: %s - %v\n", task.ID, task.Goal, err)
	}
	
	return nil
}

// ConfigureNotifications configures notification preferences
func (s *BasicNotificationService) ConfigureNotifications(prefs NotificationPreferences) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.preferences = prefs
	return nil
}