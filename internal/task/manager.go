package task

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// taskManager implements the TaskManager interface
type taskManager struct {
	mu     sync.RWMutex
	tasks  map[string]*TaskExecution
	closed bool
}

// NewManager creates a new TaskManager instance
func NewManager() TaskManager {
	return &taskManager{
		tasks: make(map[string]*TaskExecution),
	}
}

// generateTaskID creates a unique task ID
func generateTaskID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "task-" + hex.EncodeToString(bytes)
}

// StartTask creates and starts a new task execution
func (tm *taskManager) StartTask(ctx context.Context, goal string) (*TaskExecution, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("cannot start task: %w", ctx.Err())
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return nil, fmt.Errorf("manager is closed")
	}

	taskCtx, cancel := context.WithCancel(ctx)
	
	task := &TaskExecution{
		ID:        generateTaskID(),
		Goal:      goal,
		Status:    StatusQueued,
		StartTime: time.Now(),
		Results:   make([]TaskResult, 0),
		Messages:  make([]CommunicationLog, 0),
		ctx:       taskCtx,
		cancel:    cancel,
	}

	tm.tasks[task.ID] = task

	// Add initial log message
	task.Messages = append(task.Messages, CommunicationLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Task created: %s", goal),
		Source:    "task-manager",
	})

	// Start background processing (simplified for now)
	go tm.processTask(task)

	return task, nil
}

// GetTask retrieves a task by ID
func (tm *taskManager) GetTask(taskID string) (*TaskExecution, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.closed {
		return nil, fmt.Errorf("manager is closed")
	}

	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Return a copy to prevent external modifications
	return tm.copyTask(task), nil
}

// ListTasks returns tasks matching the given filter
func (tm *taskManager) ListTasks(filter TaskFilter) ([]*TaskExecution, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.closed {
		return nil, fmt.Errorf("manager is closed")
	}

	var tasks []*TaskExecution

	// Collect tasks that match the filter
	for _, task := range tm.tasks {
		if tm.matchesFilter(task, filter) {
			tasks = append(tasks, tm.copyTask(task))
		}
	}

	// Sort tasks by start time (newest first)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].StartTime.After(tasks[j].StartTime)
	})

	// Apply offset and limit
	if filter.Offset > 0 && filter.Offset < len(tasks) {
		tasks = tasks[filter.Offset:]
	} else if filter.Offset >= len(tasks) {
		tasks = []*TaskExecution{}
	}

	if filter.Limit > 0 && filter.Limit < len(tasks) {
		tasks = tasks[:filter.Limit]
	}

	return tasks, nil
}

// CancelTask cancels a running task
func (tm *taskManager) CancelTask(taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return fmt.Errorf("manager is closed")
	}

	task, exists := tm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Check if task can be cancelled
	if task.Status == StatusCompleted || task.Status == StatusFailed || task.Status == StatusCancelled {
		return fmt.Errorf("cannot cancel task in status: %s", task.Status)
	}

	// Cancel the task
	task.Status = StatusCancelled
	now := time.Now()
	task.EndTime = &now
	task.Messages = append(task.Messages, CommunicationLog{
		Timestamp: now,
		Level:     "info",
		Message:   "Task cancelled by user",
		Source:    "task-manager",
	})

	// Cancel the context to stop background processing
	if task.cancel != nil {
		task.cancel()
	}

	return nil
}

// Close shuts down the task manager
func (tm *taskManager) Close() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return nil
	}

	tm.closed = true

	// Cancel all running tasks
	for _, task := range tm.tasks {
		if task.cancel != nil && (task.Status == StatusQueued || task.Status == StatusPlanning || task.Status == StatusRunning) {
			task.cancel()
		}
	}

	return nil
}

// processTask handles background task processing (simplified implementation)
func (tm *taskManager) processTask(task *TaskExecution) {
	defer func() {
		if r := recover(); r != nil {
			tm.mu.Lock()
			task.Status = StatusFailed
			task.Error = fmt.Sprintf("task panicked: %v", r)
			now := time.Now()
			task.EndTime = &now
			task.Messages = append(task.Messages, CommunicationLog{
				Timestamp: now,
				Level:     "error",
				Message:   fmt.Sprintf("Task failed with panic: %v", r),
				Source:    "task-manager",
			})
			tm.mu.Unlock()
		}
	}()

	// Simulate task planning phase
	tm.mu.Lock()
	task.Status = StatusPlanning
	task.Messages = append(task.Messages, CommunicationLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Captain has begun planning...",
		Source:    "captain",
	})
	tm.mu.Unlock()

	// Check for cancellation
	select {
	case <-task.ctx.Done():
		return
	case <-time.After(100 * time.Millisecond): // Simulate planning time
	}

	// Create a simple plan
	tm.mu.Lock()
	task.Plan = &ExecutionPlan{
		Steps:             []string{"Analyze goal", "Create execution plan", "Execute plan"},
		CreatedAt:         time.Now(),
		EstimatedDuration: 30 * time.Second,
	}
	task.Status = StatusRunning
	task.Messages = append(task.Messages, CommunicationLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Plan created, beginning execution...",
		Source:    "captain",
	})
	tm.mu.Unlock()

	// Simulate task execution
	for i, step := range task.Plan.Steps {
		select {
		case <-task.ctx.Done():
			return
		case <-time.After(50 * time.Millisecond): // Simulate step execution time
		}

		tm.mu.Lock()
		task.Results = append(task.Results, TaskResult{
			Step:        step,
			Status:      "completed",
			Output:      fmt.Sprintf("Step %d completed successfully", i+1),
			CompletedAt: time.Now(),
		})
		task.Messages = append(task.Messages, CommunicationLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Completed step: %s", step),
			Source:    "crew",
		})
		tm.mu.Unlock()
	}

	// Mark task as completed
	tm.mu.Lock()
	task.Status = StatusCompleted
	now := time.Now()
	task.EndTime = &now
	task.Messages = append(task.Messages, CommunicationLog{
		Timestamp: now,
		Level:     "info",
		Message:   "Task completed successfully",
		Source:    "captain",
	})
	tm.mu.Unlock()
}

// copyTask creates a deep copy of a task for safe external access
func (tm *taskManager) copyTask(task *TaskExecution) *TaskExecution {
	copy := &TaskExecution{
		ID:        task.ID,
		Goal:      task.Goal,
		Status:    task.Status,
		StartTime: task.StartTime,
		EndTime:   task.EndTime,
		Error:     task.Error,
		Results:   make([]TaskResult, len(task.Results)),
		Messages:  make([]CommunicationLog, len(task.Messages)),
	}

	// Copy results
	for i, result := range task.Results {
		copy.Results[i] = result
	}

	// Copy messages
	for i, message := range task.Messages {
		copy.Messages[i] = message
	}

	// Copy plan if it exists
	if task.Plan != nil {
		copy.Plan = &ExecutionPlan{
			Steps:             make([]string, len(task.Plan.Steps)),
			CreatedAt:         task.Plan.CreatedAt,
			EstimatedDuration: task.Plan.EstimatedDuration,
		}
		for i, step := range task.Plan.Steps {
			copy.Plan.Steps[i] = step
		}
	}

	return copy
}

// matchesFilter checks if a task matches the given filter criteria
func (tm *taskManager) matchesFilter(task *TaskExecution, filter TaskFilter) bool {
	// Check status filter
	if filter.Status != nil && task.Status != *filter.Status {
		return false
	}

	// Check time filter
	if filter.SinceTime != nil && task.StartTime.Before(*filter.SinceTime) {
		return false
	}

	return true
}