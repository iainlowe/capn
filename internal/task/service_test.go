package task

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskQuery_Interface(t *testing.T) {
	// This test verifies that our implementation satisfies the interface
	var _ TaskQuery = (*InMemoryStorage)(nil)
}

func TestNotificationService_Interface(t *testing.T) {
	// This test verifies that our implementation satisfies the interface
	var _ NotificationService = (*BasicNotificationService)(nil)
}

func TestInMemoryStorage_StoreTask(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	task := &TaskExecution{
		ID:      "test-001",
		Status:  TaskStatusRunning,
		Goal:    "test goal",
		Started: time.Now(),
	}

	err := storage.StoreTask(ctx, task)
	assert.NoError(t, err)

	// Verify task was stored
	details, err := storage.GetTaskDetails(ctx, "test-001")
	assert.NoError(t, err)
	assert.Equal(t, "test-001", details.ID)
	assert.Equal(t, TaskStatusRunning, details.Status)
	assert.Equal(t, "test goal", details.Goal)
}

func TestInMemoryStorage_ListTasks(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	// Store test tasks
	tasks := []*TaskExecution{
		{ID: "task-001", Status: TaskStatusRunning, Goal: "goal 1", Started: time.Now()},
		{ID: "task-002", Status: TaskStatusCompleted, Goal: "goal 2", Started: time.Now().Add(-1 * time.Hour)},
		{ID: "task-003", Status: TaskStatusFailed, Goal: "goal 3", Started: time.Now().Add(-2 * time.Hour)},
	}

	for _, task := range tasks {
		err := storage.StoreTask(ctx, task)
		require.NoError(t, err)
	}

	// Test list all tasks
	filter := TaskFilter{}
	results, err := storage.ListTasks(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// Test filter by status
	filter = TaskFilter{Status: []TaskStatus{TaskStatusRunning}}
	results, err = storage.ListTasks(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "task-001", results[0].ID)

	// Test filter by keywords
	filter = TaskFilter{Keywords: []string{"goal 2"}}
	results, err = storage.ListTasks(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "task-002", results[0].ID)
}

func TestInMemoryStorage_GetTaskDetails(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	task := &TaskExecution{
		ID:           "test-001",
		Status:       TaskStatusRunning,
		Goal:         "test goal",
		Started:      time.Now(),
		Plan:         "test plan",
		Progress:     50.0,
		CurrentStep:  2,
		TotalSteps:   4,
		ActiveAgents: []string{"Agent-1"},
	}

	err := storage.StoreTask(ctx, task)
	require.NoError(t, err)

	details, err := storage.GetTaskDetails(ctx, "test-001")
	assert.NoError(t, err)
	assert.Equal(t, "test-001", details.ID)
	assert.Equal(t, "test plan", details.Plan)
	assert.Equal(t, 50.0, details.Progress)
	assert.Equal(t, 2, details.CurrentStep)
	assert.Equal(t, 4, details.TotalSteps)
	assert.Equal(t, []string{"Agent-1"}, details.ActiveAgents)
}

func TestInMemoryStorage_GetTaskDetails_NotFound(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	_, err := storage.GetTaskDetails(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestInMemoryStorage_AddLogEntry(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	// First store a task
	task := &TaskExecution{
		ID:      "test-001",
		Status:  TaskStatusRunning,
		Goal:    "test goal",
		Started: time.Now(),
	}
	err := storage.StoreTask(ctx, task)
	require.NoError(t, err)

	// Add log entry
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelInfo,
		Agent:     "TestAgent",
		Message:   "test message",
	}

	err = storage.AddLogEntry(ctx, "test-001", entry)
	assert.NoError(t, err)

	// Retrieve logs
	logs, err := storage.GetTaskLogs(ctx, "test-001")
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "test message", logs[0].Message)
	assert.Equal(t, "TestAgent", logs[0].Agent)
}

func TestInMemoryStorage_GetTaskLogs(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	// Store task
	task := &TaskExecution{ID: "test-001", Status: TaskStatusRunning, Started: time.Now()}
	err := storage.StoreTask(ctx, task)
	require.NoError(t, err)

	// Add multiple log entries
	entries := []LogEntry{
		{Timestamp: time.Now(), Level: LogLevelInfo, Agent: "Agent1", Message: "msg1"},
		{Timestamp: time.Now().Add(1 * time.Second), Level: LogLevelWarn, Agent: "Agent2", Message: "msg2"},
	}

	for _, entry := range entries {
		err = storage.AddLogEntry(ctx, "test-001", entry)
		require.NoError(t, err)
	}

	logs, err := storage.GetTaskLogs(ctx, "test-001")
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	// Logs should be ordered by timestamp
	assert.Equal(t, "msg1", logs[0].Message)
	assert.Equal(t, "msg2", logs[1].Message)
}

func TestInMemoryStorage_SearchTasks(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	// Store test tasks
	tasks := []*TaskExecution{
		{ID: "task-001", Status: TaskStatusRunning, Goal: "security analysis", Started: time.Now()},
		{ID: "task-002", Status: TaskStatusCompleted, Goal: "deploy to production", Started: time.Now()},
		{ID: "task-003", Status: TaskStatusFailed, Goal: "security audit", Started: time.Now()},
	}

	for _, task := range tasks {
		err := storage.StoreTask(ctx, task)
		require.NoError(t, err)
	}

	// Test search
	results, err := storage.SearchTasks(ctx, "security")
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Results should be ordered by relevance/ID
	assert.Equal(t, "task-001", results[0].ID)
	assert.Equal(t, "task-003", results[1].ID)
}

func TestBasicNotificationService_NotifyCompletion(t *testing.T) {
	service := NewBasicNotificationService()

	task := &TaskExecution{
		ID:     "test-001",
		Status: TaskStatusCompleted,
		Goal:   "test goal",
	}

	err := service.NotifyCompletion(task)
	assert.NoError(t, err)
}

func TestBasicNotificationService_NotifyError(t *testing.T) {
	service := NewBasicNotificationService()

	task := &TaskExecution{
		ID:     "test-001",
		Status: TaskStatusFailed,
		Goal:   "test goal",
	}

	testErr := assert.AnError
	err := service.NotifyError(task, testErr)
	assert.NoError(t, err)
}

func TestBasicNotificationService_ConfigureNotifications(t *testing.T) {
	service := NewBasicNotificationService()

	prefs := NotificationPreferences{
		EnableCompletion: true,
		EnableErrors:     true,
		OutputFormat:     "console",
	}

	err := service.ConfigureNotifications(prefs)
	assert.NoError(t, err)
}