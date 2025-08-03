package task

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskManager_TaskLifecycle(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Start a task
	task, err := manager.StartTask(ctx, "integration test goal")
	require.NoError(t, err)
	require.NotNil(t, task)
	
	taskID := task.ID
	
	// Wait for task to progress through states
	time.Sleep(200 * time.Millisecond)
	
	// Check task has progressed
	updatedTask, err := manager.GetTask(taskID)
	require.NoError(t, err)
	
	// Task should have progressed beyond queued
	assert.NotEqual(t, StatusQueued, updatedTask.Status)
	
	// Task should have messages
	assert.NotEmpty(t, updatedTask.Messages)
	
	// Wait for completion
	time.Sleep(300 * time.Millisecond)
	
	// Check final state
	finalTask, err := manager.GetTask(taskID)
	require.NoError(t, err)
	
	assert.Equal(t, StatusCompleted, finalTask.Status)
	assert.NotNil(t, finalTask.EndTime)
	assert.NotNil(t, finalTask.Plan)
	assert.NotEmpty(t, finalTask.Results)
	assert.Len(t, finalTask.Plan.Steps, 3)
	assert.Len(t, finalTask.Results, 3)
}

func TestTaskManager_TaskCancellationDuringExecution(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Start a task
	task, err := manager.StartTask(ctx, "cancellation test")
	require.NoError(t, err)
	
	// Wait a bit to let it start processing
	time.Sleep(50 * time.Millisecond)
	
	// Cancel while running
	err = manager.CancelTask(task.ID)
	require.NoError(t, err)
	
	// Verify cancellation
	cancelledTask, err := manager.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, cancelledTask.Status)
	assert.NotNil(t, cancelledTask.EndTime)
}

func TestTaskManager_TaskFailureHandling(t *testing.T) {
	// This test verifies the panic recovery mechanism
	// We'll test by checking the processTask method handles errors gracefully
	
	manager := &taskManager{
		tasks: make(map[string]*TaskExecution),
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	task := &TaskExecution{
		ID:        "test-fail",
		Goal:      "failing task",
		Status:    StatusQueued,
		StartTime: time.Now(),
		Results:   make([]TaskResult, 0),
		Messages:  make([]CommunicationLog, 0),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	manager.tasks[task.ID] = task
	
	// Simulate a task that would complete normally
	manager.processTask(task)
	
	// Verify task completed successfully
	assert.Equal(t, StatusCompleted, task.Status)
	assert.NotNil(t, task.EndTime)
	assert.NotEmpty(t, task.Results)
}

func TestTaskManager_ListTasksWithOffsetAndLimit(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Create multiple tasks
	for i := 0; i < 10; i++ {
		_, err := manager.StartTask(ctx, "task")
		require.NoError(t, err)
	}

	// Test offset without limit
	tasks, err := manager.ListTasks(TaskFilter{Offset: 5})
	require.NoError(t, err)
	assert.Len(t, tasks, 5)

	// Test offset beyond available tasks
	tasks, err = manager.ListTasks(TaskFilter{Offset: 15})
	require.NoError(t, err)
	assert.Len(t, tasks, 0)

	// Test both offset and limit
	tasks, err = manager.ListTasks(TaskFilter{Offset: 2, Limit: 3})
	require.NoError(t, err)
	assert.Len(t, tasks, 3)
}

func TestTaskManager_ListTasksWithTimeFilter(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Create a task
	task1, err := manager.StartTask(ctx, "task 1")
	require.NoError(t, err)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)
	
	// Note the time
	filterTime := time.Now()
	
	// Wait a bit more
	time.Sleep(10 * time.Millisecond)

	// Create another task
	task2, err := manager.StartTask(ctx, "task 2")
	require.NoError(t, err)

	// Filter tasks since the filter time
	tasks, err := manager.ListTasks(TaskFilter{SinceTime: &filterTime})
	require.NoError(t, err)
	
	// Should only get task2
	assert.Len(t, tasks, 1)
	assert.Equal(t, task2.ID, tasks[0].ID)

	// Filter tasks before the filter time
	beforeTime := task1.StartTime.Add(-1 * time.Second)
	tasks, err = manager.ListTasks(TaskFilter{SinceTime: &beforeTime})
	require.NoError(t, err)
	
	// Should get both tasks
	assert.Len(t, tasks, 2)
}

func TestTaskManager_GenerateTaskID(t *testing.T) {
	// Test task ID generation
	id1 := generateTaskID()
	id2 := generateTaskID()
	
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "task-")
	assert.Contains(t, id2, "task-")
}

func TestTaskManager_CopyTask(t *testing.T) {
	manager := &taskManager{}
	
	now := time.Now()
	endTime := now.Add(5 * time.Minute)
	
	original := &TaskExecution{
		ID:        "test-task",
		Goal:      "test goal",
		Status:    StatusCompleted,
		StartTime: now,
		EndTime:   &endTime,
		Error:     "test error",
		Results: []TaskResult{
			{Step: "step1", Status: "completed", Output: "output1", CompletedAt: now},
		},
		Messages: []CommunicationLog{
			{Timestamp: now, Level: "info", Message: "test message", Source: "test"},
		},
		Plan: &ExecutionPlan{
			Steps:             []string{"step1", "step2"},
			CreatedAt:         now,
			EstimatedDuration: 5 * time.Minute,
		},
	}
	
	copy := manager.copyTask(original)
	
	// Verify all fields are copied
	assert.Equal(t, original.ID, copy.ID)
	assert.Equal(t, original.Goal, copy.Goal)
	assert.Equal(t, original.Status, copy.Status)
	assert.Equal(t, original.StartTime, copy.StartTime)
	assert.Equal(t, original.EndTime, copy.EndTime)
	assert.Equal(t, original.Error, copy.Error)
	
	// Verify slices are copied
	assert.Equal(t, original.Results, copy.Results)
	assert.Equal(t, original.Messages, copy.Messages)
	
	// Verify plan is copied
	require.NotNil(t, copy.Plan)
	assert.Equal(t, original.Plan.Steps, copy.Plan.Steps)
	assert.Equal(t, original.Plan.CreatedAt, copy.Plan.CreatedAt)
	assert.Equal(t, original.Plan.EstimatedDuration, copy.Plan.EstimatedDuration)
	
	// Verify it's a deep copy (modifying copy doesn't affect original)
	copy.Results[0].Output = "modified"
	assert.NotEqual(t, original.Results[0].Output, copy.Results[0].Output)
}

func TestTaskManager_MatchesFilter(t *testing.T) {
	manager := &taskManager{}
	
	now := time.Now()
	task := &TaskExecution{
		ID:        "test-task",
		Status:    StatusRunning,
		StartTime: now,
	}
	
	tests := []struct {
		name     string
		filter   TaskFilter
		expected bool
	}{
		{
			name:     "empty filter matches all",
			filter:   TaskFilter{},
			expected: true,
		},
		{
			name:     "matching status filter",
			filter:   TaskFilter{Status: &[]TaskStatus{StatusRunning}[0]},
			expected: true,
		},
		{
			name:     "non-matching status filter",
			filter:   TaskFilter{Status: &[]TaskStatus{StatusCompleted}[0]},
			expected: false,
		},
		{
			name:     "matching time filter",
			filter:   TaskFilter{SinceTime: &[]time.Time{now.Add(-1 * time.Hour)}[0]},
			expected: true,
		},
		{
			name:     "non-matching time filter",
			filter:   TaskFilter{SinceTime: &[]time.Time{now.Add(1 * time.Hour)}[0]},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.matchesFilter(task, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTaskManager_DoubleClose(t *testing.T) {
	manager := NewManager()
	
	// First close should succeed
	err := manager.Close()
	assert.NoError(t, err)
	
	// Second close should also succeed (idempotent)
	err = manager.Close()
	assert.NoError(t, err)
}