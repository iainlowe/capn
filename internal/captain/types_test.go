package captain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_JSONSerialization(t *testing.T) {
	tests := []struct {
		name string
		task Task
	}{
		{
			name: "basic task",
			task: Task{
				ID:           "task-1",
				Type:         TaskTypeAnalysis,
				Priority:     PriorityHigh,
				Dependencies: []string{"task-0"},
				Payload: map[string]any{
					"goal": "analyze code quality",
					"path": "/src",
				},
				Deadline: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Metadata: map[string]string{
					"agent": "code-agent",
					"retry": "3",
				},
			},
		},
		{
			name: "minimal task",
			task: Task{
				ID:       "task-minimal",
				Type:     TaskTypeExecution,
				Priority: PriorityMedium,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.task)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test unmarshaling
			var unmarshaled Task
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify all fields match
			assert.Equal(t, tt.task.ID, unmarshaled.ID)
			assert.Equal(t, tt.task.Type, unmarshaled.Type)
			assert.Equal(t, tt.task.Priority, unmarshaled.Priority)
			assert.Equal(t, tt.task.Dependencies, unmarshaled.Dependencies)
			assert.Equal(t, tt.task.Payload, unmarshaled.Payload)
			assert.Equal(t, tt.task.Deadline.Unix(), unmarshaled.Deadline.Unix())
			assert.Equal(t, tt.task.Metadata, unmarshaled.Metadata)
		})
	}
}

func TestExecutionPlan_JSONSerialization(t *testing.T) {
	plan := ExecutionPlan{
		ID:   "plan-1",
		Goal: "improve code quality",
		Tasks: []Task{
			{
				ID:       "task-1",
				Type:     TaskTypeAnalysis,
				Priority: PriorityHigh,
				Payload: map[string]any{
					"action": "analyze",
				},
			},
			{
				ID:       "task-2",
				Type:     TaskTypeExecution,
				Priority: PriorityMedium,
				Dependencies: []string{"task-1"},
				Payload: map[string]any{
					"action": "fix",
				},
			},
		},
		Timeline: ExecutionTimeline{
			EstimatedDuration: 30 * time.Minute,
			StartTime:         time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			EndTime:           time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		Resources: ResourceAllocation{
			MaxAgents:     5,
			RequiredTools: []string{"code-analyzer", "test-runner"},
			EstimatedCost: 0.50,
		},
		Strategy: ExecutionStrategy{
			Type:        StrategyParallel,
			Description: "Run analysis and fixes in parallel where possible",
			Options: map[string]any{
				"max_parallel": 3,
				"retry_count":  2,
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(plan)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test unmarshaling
	var unmarshaled ExecutionPlan
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, plan.ID, unmarshaled.ID)
	assert.Equal(t, plan.Goal, unmarshaled.Goal)
	assert.Len(t, unmarshaled.Tasks, 2)
	assert.Equal(t, plan.Tasks[0].ID, unmarshaled.Tasks[0].ID)
	assert.Equal(t, plan.Tasks[1].Dependencies, unmarshaled.Tasks[1].Dependencies)
	assert.Equal(t, plan.Timeline.EstimatedDuration, unmarshaled.Timeline.EstimatedDuration)
	assert.Equal(t, plan.Resources.MaxAgents, unmarshaled.Resources.MaxAgents)
	assert.Equal(t, plan.Strategy.Type, unmarshaled.Strategy.Type)
}

func TestTaskType_String(t *testing.T) {
	tests := []struct {
		taskType TaskType
		expected string
	}{
		{TaskTypeAnalysis, "analysis"},
		{TaskTypeExecution, "execution"},
		{TaskTypeValidation, "validation"},
		{TaskTypeReporting, "reporting"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.taskType))
		})
	}
}

func TestPriority_String(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{PriorityLow, "low"},
		{PriorityMedium, "medium"},
		{PriorityHigh, "high"},
		{PriorityCritical, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.priority))
		})
	}
}

func TestStrategyType_String(t *testing.T) {
	tests := []struct {
		strategy StrategyType
		expected string
	}{
		{StrategySequential, "sequential"},
		{StrategyParallel, "parallel"},
		{StrategyHybrid, "hybrid"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.strategy))
		})
	}
}