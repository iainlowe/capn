package task

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskManager_StartTask(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	task, err := manager.StartTask(ctx, "test goal")
	
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "test goal", task.Goal)
	assert.Equal(t, StatusQueued, task.Status)
	assert.False(t, task.StartTime.IsZero())
	assert.Nil(t, task.EndTime)
	assert.Empty(t, task.Results)
	assert.Len(t, task.Messages, 1) // Should have initial creation message
	assert.Empty(t, task.Error)
}

func TestTaskManager_StartTask_WithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	manager := NewManager()
	defer manager.Close()

	task, err := manager.StartTask(ctx, "test goal")
	
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestTaskManager_GetTask(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Start a task first
	originalTask, err := manager.StartTask(ctx, "test goal")
	require.NoError(t, err)
	require.NotNil(t, originalTask)

	// Get the task
	retrievedTask, err := manager.GetTask(originalTask.ID)
	
	require.NoError(t, err)
	require.NotNil(t, retrievedTask)
	assert.Equal(t, originalTask.ID, retrievedTask.ID)
	assert.Equal(t, originalTask.Goal, retrievedTask.Goal)
	assert.Equal(t, originalTask.Status, retrievedTask.Status)
}

func TestTaskManager_GetTask_NotFound(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	task, err := manager.GetTask("non-existent-id")
	
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskManager_ListTasks(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Start multiple tasks
	task1, err := manager.StartTask(ctx, "goal 1")
	require.NoError(t, err)
	
	task2, err := manager.StartTask(ctx, "goal 2")
	require.NoError(t, err)

	// List all tasks
	tasks, err := manager.ListTasks(TaskFilter{})
	
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
	
	taskIDs := []string{tasks[0].ID, tasks[1].ID}
	assert.Contains(t, taskIDs, task1.ID)
	assert.Contains(t, taskIDs, task2.ID)
}

func TestTaskManager_ListTasks_WithStatusFilter(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Start a task
	task, err := manager.StartTask(ctx, "test goal")
	require.NoError(t, err)
	
	// Filter by status
	status := StatusQueued
	tasks, err := manager.ListTasks(TaskFilter{Status: &status})
	
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, task.ID, tasks[0].ID)
	assert.Equal(t, StatusQueued, tasks[0].Status)
}

func TestTaskManager_ListTasks_WithLimit(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Start multiple tasks
	for i := 0; i < 5; i++ {
		_, err := manager.StartTask(ctx, "goal")
		require.NoError(t, err)
	}

	// List with limit
	tasks, err := manager.ListTasks(TaskFilter{Limit: 3})
	
	require.NoError(t, err)
	assert.Len(t, tasks, 3)
}

func TestTaskManager_CancelTask(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Start a task
	task, err := manager.StartTask(ctx, "test goal")
	require.NoError(t, err)

	// Cancel the task
	err = manager.CancelTask(task.ID)
	require.NoError(t, err)

	// Verify task is cancelled
	retrievedTask, err := manager.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, retrievedTask.Status)
	assert.NotNil(t, retrievedTask.EndTime)
}

func TestTaskManager_CancelTask_NotFound(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	err := manager.CancelTask("non-existent-id")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskManager_CancelTask_AlreadyCompleted(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	// Start a task
	task, err := manager.StartTask(ctx, "test goal")
	require.NoError(t, err)

	// Manually set task as completed (simulating completion)
	if m, ok := manager.(*taskManager); ok {
		m.mu.Lock()
		storedTask := m.tasks[task.ID]
		storedTask.Status = StatusCompleted
		now := time.Now()
		storedTask.EndTime = &now
		m.mu.Unlock()
	}

	// Try to cancel completed task
	err = manager.CancelTask(task.ID)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel")
}

func TestTaskManager_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	defer manager.Close()

	var wg sync.WaitGroup
	taskIDs := make([]string, 10)
	
	// Start multiple tasks concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			task, err := manager.StartTask(ctx, "concurrent goal")
			require.NoError(t, err)
			taskIDs[idx] = task.ID
		}(i)
	}
	
	wg.Wait()

	// Verify all tasks were created
	tasks, err := manager.ListTasks(TaskFilter{})
	require.NoError(t, err)
	assert.Len(t, tasks, 10)
	
	// Verify all task IDs are unique
	uniqueIDs := make(map[string]bool)
	for _, id := range taskIDs {
		assert.NotEmpty(t, id)
		assert.False(t, uniqueIDs[id], "Duplicate task ID: %s", id)
		uniqueIDs[id] = true
	}
}

func TestTaskManager_Close(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()

	// Start a task
	_, err := manager.StartTask(ctx, "test goal")
	require.NoError(t, err)

	// Close manager
	err = manager.Close()
	assert.NoError(t, err)

	// Operations after close should fail
	_, err = manager.StartTask(ctx, "another goal")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "manager is closed")
}