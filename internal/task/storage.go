package task

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// InMemoryStorage provides an in-memory implementation of TaskStorage
type InMemoryStorage struct {
	mu    sync.RWMutex
	tasks map[string]*TaskExecution
	logs  map[string][]LogEntry
}

// NewInMemoryStorage creates a new in-memory storage instance
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		tasks: make(map[string]*TaskExecution),
		logs:  make(map[string][]LogEntry),
	}
}

// StoreTask stores a task execution
func (s *InMemoryStorage) StoreTask(ctx context.Context, task *TaskExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.tasks[task.ID] = task
	
	// Initialize logs slice if it doesn't exist
	if _, exists := s.logs[task.ID]; !exists {
		s.logs[task.ID] = make([]LogEntry, 0)
	}
	
	return nil
}

// UpdateTask updates an existing task
func (s *InMemoryStorage) UpdateTask(ctx context.Context, task *TaskExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.tasks[task.ID]; !exists {
		return fmt.Errorf("task not found: %s", task.ID)
	}
	
	s.tasks[task.ID] = task
	return nil
}

// ListTasks returns a list of task summaries matching the filter
func (s *InMemoryStorage) ListTasks(ctx context.Context, filter TaskFilter) ([]*TaskSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var results []*TaskSummary
	
	for _, task := range s.tasks {
		summary := task.ToSummary()
		if filter.Matches(summary) {
			results = append(results, summary)
		}
	}
	
	// Sort by started time (most recent first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Started.After(results[j].Started)
	})
	
	return results, nil
}

// GetTaskDetails returns detailed information about a specific task
func (s *InMemoryStorage) GetTaskDetails(ctx context.Context, taskID string) (*TaskDetails, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	return task.ToDetails(), nil
}

// GetTaskLogs returns the log entries for a specific task
func (s *InMemoryStorage) GetTaskLogs(ctx context.Context, taskID string) ([]LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	logs, exists := s.logs[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	// Return a copy to avoid concurrent modification
	result := make([]LogEntry, len(logs))
	copy(result, logs)
	
	return result, nil
}

// AddLogEntry adds a log entry to a task
func (s *InMemoryStorage) AddLogEntry(ctx context.Context, taskID string, entry LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check if task exists
	if _, exists := s.tasks[taskID]; !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	// Initialize logs slice if it doesn't exist
	if _, exists := s.logs[taskID]; !exists {
		s.logs[taskID] = make([]LogEntry, 0)
	}
	
	s.logs[taskID] = append(s.logs[taskID], entry)
	
	// Sort logs by timestamp
	sort.Slice(s.logs[taskID], func(i, j int) bool {
		return s.logs[taskID][i].Timestamp.Before(s.logs[taskID][j].Timestamp)
	})
	
	return nil
}

// SearchTasks searches for tasks matching the query string
func (s *InMemoryStorage) SearchTasks(ctx context.Context, query string) ([]*TaskSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	query = strings.ToLower(query)
	var results []*TaskSummary
	
	for _, task := range s.tasks {
		goalLower := strings.ToLower(task.Goal)
		if strings.Contains(goalLower, query) {
			results = append(results, task.ToSummary())
		}
	}
	
	// Sort by ID for consistent ordering
	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})
	
	return results, nil
}