package captain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iainlowe/capn/internal/config"
)

func TestNewCaptain(t *testing.T) {
	cfg := &config.Config{
		Captain: config.CaptainConfig{
			MaxConcurrentAgents: 5,
			PlanningTimeout:     30000000000, // 30 seconds in nanoseconds
		},
	}

	openaiConfig := OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	captain, err := NewCaptain("captain-1", cfg, openaiConfig)
	require.NoError(t, err)
	assert.NotNil(t, captain)
	assert.Equal(t, "captain-1", captain.ID())
	assert.NotNil(t, captain.planner)
	assert.NotNil(t, captain.taskQueue)
	assert.NotNil(t, captain.resultChan)
}

func TestNewCaptain_InvalidConfig(t *testing.T) {
	cfg := &config.Config{
		Captain: config.CaptainConfig{
			MaxConcurrentAgents: 5,
			PlanningTimeout:     30000000000,
		},
	}

	openaiConfig := OpenAIConfig{
		APIKey: "", // Invalid - empty API key
		Model:  "gpt-3.5-turbo",
	}

	captain, err := NewCaptain("captain-1", cfg, openaiConfig)
	assert.Error(t, err)
	assert.Nil(t, captain)
}

func TestCaptain_CreatePlan(t *testing.T) {
	cfg := &config.Config{
		Captain: config.CaptainConfig{
			MaxConcurrentAgents: 5,
			PlanningTimeout:     30000000000,
		},
	}

	mockLLM := &MockLLMProvider{}

	// Setup mock response
	response := &CompletionResponse{
		Content: `{
			"tasks": [
				{
					"id": "task-1",
					"type": "analysis",
					"priority": "high",
					"description": "Analyze code quality",
					"dependencies": []
				}
			],
			"strategy": "sequential",
			"estimated_duration": "15m"
		}`,
		TokensUsed: 100,
		Model:      "gpt-3.5-turbo",
	}

	mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything).Return(response, nil)

	captain := &Captain{
		id:          "captain-1",
		config:      cfg,
		llmProvider: mockLLM,
		planner:     NewPlanningEngine(mockLLM),
		taskQueue:   make(chan Task, 100),
		resultChan:  make(chan Result, 100),
	}

	ctx := context.Background()
	goal := "analyze code quality"

	plan, err := captain.CreatePlan(ctx, goal)
	require.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, goal, plan.Goal)
	assert.Len(t, plan.Tasks, 1)
	assert.Equal(t, "task-1", plan.Tasks[0].ID)

	mockLLM.AssertExpectations(t)
}

func TestCaptain_ExecutePlan_DryRun(t *testing.T) {
	cfg := &config.Config{
		Captain: config.CaptainConfig{
			MaxConcurrentAgents: 5,
			PlanningTimeout:     30000000000,
		},
	}

	mockLLM := &MockLLMProvider{}
	captain := &Captain{
		id:          "captain-1",
		config:      cfg,
		llmProvider: mockLLM,
		planner:     NewPlanningEngine(mockLLM),
		taskQueue:   make(chan Task, 100),
		resultChan:  make(chan Result, 100),
	}

	plan := &ExecutionPlan{
		ID:   "plan-1",
		Goal: "test goal",
		Tasks: []Task{
			{ID: "task-1", Type: TaskTypeAnalysis, Priority: PriorityHigh},
		},
	}

	ctx := context.Background()
	result, err := captain.ExecutePlan(ctx, plan, true) // dry run = true

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.DryRun)
	assert.Equal(t, plan.ID, result.PlanID)
	assert.Len(t, result.TaskResults, 1)
	assert.True(t, result.TaskResults[0].Success)
	assert.Contains(t, result.TaskResults[0].Output, "DRY RUN")
}

func TestCaptain_ExecutePlan_Execution(t *testing.T) {
	cfg := &config.Config{
		Captain: config.CaptainConfig{
			MaxConcurrentAgents: 5,
			PlanningTimeout:     30 * time.Second,
		},
	}

	mockLLM := &MockLLMProvider{}
	captain := &Captain{
		id:          "captain-1",
		config:      cfg,
		llmProvider: mockLLM,
		planner:     NewPlanningEngine(mockLLM),
		taskQueue:   make(chan Task, 100),
		resultChan:  make(chan Result, 100),
	}

	plan := &ExecutionPlan{
		ID:   "plan-1",
		Goal: "test goal",
		Tasks: []Task{
			{ID: "task-1", Type: TaskTypeAnalysis, Priority: PriorityHigh},
		},
	}

	ctx := context.Background()
	result, err := captain.ExecutePlan(ctx, plan, false)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.DryRun)
	assert.Equal(t, plan.ID, result.PlanID)
	assert.Len(t, result.TaskResults, 1)
	assert.True(t, result.TaskResults[0].Success)
	assert.Contains(t, result.TaskResults[0].Output, "executed successfully")
}

func TestCaptain_Status(t *testing.T) {
	cfg := &config.Config{
		Captain: config.CaptainConfig{
			MaxConcurrentAgents: 5,
			PlanningTimeout:     30000000000,
		},
	}

	openaiConfig := OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	captain, err := NewCaptain("captain-1", cfg, openaiConfig)
	require.NoError(t, err)

	status := captain.Status()
	assert.Equal(t, "captain-1", status.ID)
	assert.Equal(t, AgentStatusIdle, status.Status)
	assert.Equal(t, 0, status.ActiveTasks)
	assert.Equal(t, 0, status.QueuedTasks)
}

func TestCaptain_Stop(t *testing.T) {
	cfg := &config.Config{
		Captain: config.CaptainConfig{
			MaxConcurrentAgents: 5,
			PlanningTimeout:     30000000000,
		},
	}

	openaiConfig := OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	captain, err := NewCaptain("captain-1", cfg, openaiConfig)
	require.NoError(t, err)

	err = captain.Stop()
	assert.NoError(t, err)
}
