package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected string
	}{
		{"queued", StatusQueued, "queued"},
		{"planning", StatusPlanning, "planning"},
		{"running", StatusRunning, "running"},
		{"completed", StatusCompleted, "completed"},
		{"failed", StatusFailed, "failed"},
		{"cancelled", StatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestTaskExecution_Creation(t *testing.T) {
	now := time.Now()
	
	task := &TaskExecution{
		ID:        "test-001",
		Goal:      "test goal",
		Status:    StatusQueued,
		StartTime: now,
		Results:   []TaskResult{},
		Messages:  []CommunicationLog{},
	}
	
	assert.Equal(t, "test-001", task.ID)
	assert.Equal(t, "test goal", task.Goal)
	assert.Equal(t, StatusQueued, task.Status)
	assert.Equal(t, now, task.StartTime)
	assert.Nil(t, task.EndTime)
	assert.Nil(t, task.Plan)
	assert.Empty(t, task.Results)
	assert.Empty(t, task.Messages)
	assert.Empty(t, task.Error)
}

func TestTaskResult_Structure(t *testing.T) {
	now := time.Now()
	
	result := TaskResult{
		Step:        "test step",
		Status:      "completed",
		Output:      "test output",
		Error:       "",
		CompletedAt: now,
	}
	
	assert.Equal(t, "test step", result.Step)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, "test output", result.Output)
	assert.Empty(t, result.Error)
	assert.Equal(t, now, result.CompletedAt)
}

func TestCommunicationLog_Structure(t *testing.T) {
	now := time.Now()
	
	log := CommunicationLog{
		Timestamp: now,
		Level:     "info",
		Message:   "test message",
		Source:    "captain",
	}
	
	assert.Equal(t, now, log.Timestamp)
	assert.Equal(t, "info", log.Level)
	assert.Equal(t, "test message", log.Message)
	assert.Equal(t, "captain", log.Source)
}

func TestExecutionPlan_Structure(t *testing.T) {
	now := time.Now()
	duration := 5 * time.Minute
	
	plan := ExecutionPlan{
		Steps:             []string{"step1", "step2", "step3"},
		CreatedAt:         now,
		EstimatedDuration: duration,
	}
	
	assert.Equal(t, []string{"step1", "step2", "step3"}, plan.Steps)
	assert.Equal(t, now, plan.CreatedAt)
	assert.Equal(t, duration, plan.EstimatedDuration)
}

func TestTaskFilter_Structure(t *testing.T) {
	status := StatusRunning
	now := time.Now()
	
	filter := TaskFilter{
		Status:    &status,
		Limit:     10,
		Offset:    5,
		SinceTime: &now,
	}
	
	require.NotNil(t, filter.Status)
	assert.Equal(t, StatusRunning, *filter.Status)
	assert.Equal(t, 10, filter.Limit)
	assert.Equal(t, 5, filter.Offset)
	require.NotNil(t, filter.SinceTime)
	assert.Equal(t, now, *filter.SinceTime)
}