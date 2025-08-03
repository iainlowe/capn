package cli

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iainlowe/capn/internal/task"
)

func TestTasksCmd_List(t *testing.T) {
	// Create a CLI with task storage
	cli := NewCLI()
	var buf bytes.Buffer
	cli.SetOutput(&buf)
	cli.SetSkipConfigForTests(true)
	
	// Set up test data
	storage := task.NewInMemoryStorage()
	ctx := context.Background()
	
	// Add test tasks
	testTasks := []*task.TaskExecution{
		{
			ID:      "task-001",
			Status:  task.TaskStatusRunning,
			Goal:    "security analysis",
			Started: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:       "task-002",
			Status:   task.TaskStatusCompleted,
			Goal:     "deploy application",
			Started:  time.Now().Add(-1 * time.Hour),
			Finished: timePtr(time.Now().Add(-30 * time.Minute)),
		},
	}
	
	for _, testTask := range testTasks {
		err := storage.StoreTask(ctx, testTask)
		require.NoError(t, err)
	}
	
	// Set the storage in CLI (we'll need to modify CLI to support this)
	cli.taskStorage = storage
	
	args := []string{"tasks", "list"}
	err := cli.Parse(args)
	
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "task-001")
	assert.Contains(t, output, "task-002")
	assert.Contains(t, output, "security analysis")
	assert.Contains(t, output, "deploy application")
	assert.Contains(t, output, "running")
	assert.Contains(t, output, "completed")
}

func TestTasksCmd_Show(t *testing.T) {
	cli := NewCLI()
	var buf bytes.Buffer
	cli.SetOutput(&buf)
	cli.SetSkipConfigForTests(true)
	
	// Set up test data
	storage := task.NewInMemoryStorage()
	ctx := context.Background()
	
	testTask := &task.TaskExecution{
		ID:           "task-001",
		Status:       task.TaskStatusRunning,
		Goal:         "security analysis",
		Started:      time.Now().Add(-1 * time.Hour),
		Plan:         "1. Scan code\n2. Check dependencies\n3. Generate report",
		Progress:     75.0,
		CurrentStep:  3,
		TotalSteps:   4,
		ActiveAgents: []string{"SecurityAgent", "ScanAgent"},
	}
	
	err := storage.StoreTask(ctx, testTask)
	require.NoError(t, err)
	
	cli.taskStorage = storage
	
	args := []string{"tasks", "show", "task-001"}
	err = cli.Parse(args)
	
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "task-001")
	assert.Contains(t, output, "security analysis")
	assert.Contains(t, output, "75% complete")
	assert.Contains(t, output, "SecurityAgent")
	assert.Contains(t, output, "ScanAgent")
}

func TestTasksCmd_Logs(t *testing.T) {
	cli := NewCLI()
	var buf bytes.Buffer
	cli.SetOutput(&buf)
	cli.SetSkipConfigForTests(true)
	
	// Set up test data
	storage := task.NewInMemoryStorage()
	ctx := context.Background()
	
	testTask := &task.TaskExecution{
		ID:      "task-001",
		Status:  task.TaskStatusRunning,
		Goal:    "test task",
		Started: time.Now(),
	}
	
	err := storage.StoreTask(ctx, testTask)
	require.NoError(t, err)
	
	// Add log entries
	logEntries := []task.LogEntry{
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			Level:     task.LogLevelInfo,
			Agent:     "TestAgent",
			Message:   "Starting task execution",
		},
		{
			Timestamp: time.Now(),
			Level:     task.LogLevelWarn,
			Agent:     "TestAgent",
			Message:   "Warning: potential issue detected",
		},
	}
	
	for _, entry := range logEntries {
		err = storage.AddLogEntry(ctx, "task-001", entry)
		require.NoError(t, err)
	}
	
	cli.taskStorage = storage
	
	args := []string{"tasks", "logs", "task-001"}
	err = cli.Parse(args)
	
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Starting task execution")
	assert.Contains(t, output, "Warning: potential issue detected")
	assert.Contains(t, output, "TestAgent")
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "warn")
}

func TestStatusCmd_WithTasks(t *testing.T) {
	cli := NewCLI()
	var buf bytes.Buffer
	cli.SetOutput(&buf)
	cli.SetSkipConfigForTests(true)
	
	// Set up test data
	storage := task.NewInMemoryStorage()
	ctx := context.Background()
	
	// Add test tasks
	testTasks := []*task.TaskExecution{
		{ID: "task-001", Status: task.TaskStatusRunning, Goal: "task 1", Started: time.Now()},
		{ID: "task-002", Status: task.TaskStatusCompleted, Goal: "task 2", Started: time.Now()},
		{ID: "task-003", Status: task.TaskStatusQueued, Goal: "task 3", Started: time.Now()},
	}
	
	for _, testTask := range testTasks {
		err := storage.StoreTask(ctx, testTask)
		require.NoError(t, err)
	}
	
	cli.taskStorage = storage
	
	args := []string{"status"}
	err := cli.Parse(args)
	
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "3 tasks total")
	assert.Contains(t, output, "1 running")
	assert.Contains(t, output, "1 completed")
	assert.Contains(t, output, "1 queued")
}

func TestTasksCmd_ListWithFilter(t *testing.T) {
	cli := NewCLI()
	var buf bytes.Buffer
	cli.SetOutput(&buf)
	cli.SetSkipConfigForTests(true)
	
	// Set up test data
	storage := task.NewInMemoryStorage()
	ctx := context.Background()
	
	// Add test tasks
	testTasks := []*task.TaskExecution{
		{ID: "task-001", Status: task.TaskStatusRunning, Goal: "security analysis", Started: time.Now()},
		{ID: "task-002", Status: task.TaskStatusCompleted, Goal: "deploy app", Started: time.Now()},
		{ID: "task-003", Status: task.TaskStatusRunning, Goal: "code review", Started: time.Now()},
	}
	
	for _, testTask := range testTasks {
		err := storage.StoreTask(ctx, testTask)
		require.NoError(t, err)
	}
	
	cli.taskStorage = storage
	
	// Test filtering by status
	args := []string{"tasks", "list", "--status", "running"}
	err := cli.Parse(args)
	
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "task-001")
	assert.Contains(t, output, "task-003")
	assert.NotContains(t, output, "task-002")
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}